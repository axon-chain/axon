package main

import (
	"context"
	"crypto/ecdsa"
	"flag"
	"fmt"
	"log/slog"
	"math/big"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	registryAddr    = "0x0000000000000000000000000000000000000801"
	heartbeatSig    = "0x3defb962" // keccak256("heartbeat()")[:4]
	isAgentSig      = "0x1ffbb064" // keccak256("isAgent(address)")[:4]
	pollInterval    = 2 * time.Second
	receiptTimeout  = 60 * time.Second
	receiptPollWait = 3 * time.Second
)

type Daemon struct {
	rpc               string
	key               *ecdsa.PrivateKey
	addr              common.Address
	heartbeatInterval uint64
	client            *ethclient.Client
	chainID           *big.Int
	registry          common.Address
	logger            *slog.Logger
}

func main() {
	var (
		rpcURL   string
		keyFile  string
		interval uint64
		logLevel string
	)

	flag.StringVar(&rpcURL, "rpc", "", "JSON-RPC endpoint")
	flag.StringVar(&keyFile, "private-key-file", "", "path to file containing hex private key")
	flag.Uint64Var(&interval, "heartbeat-interval", 100, "send heartbeat every N blocks")
	flag.StringVar(&logLevel, "log-level", "info", "log level: debug, info, warn, error")
	flag.Parse()

	if rpcURL == "" {
		fmt.Fprintln(os.Stderr, "error: --rpc is required")
		flag.Usage()
		os.Exit(1)
	}

	if keyFile == "" {
		fmt.Fprintln(os.Stderr, "error: --private-key-file is required")
		flag.Usage()
		os.Exit(1)
	}

	data, err := os.ReadFile(keyFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading key file: %v\n", err)
		os.Exit(1)
	}
	keyHex := strings.TrimSpace(string(data))

	logger := newLogger(logLevel)

	keyHex = strings.TrimPrefix(keyHex, "0x")
	privateKey, err := crypto.HexToECDSA(keyHex)
	if err != nil {
		logger.Error("invalid private key", "error", err)
		os.Exit(1)
	}

	addr := crypto.PubkeyToAddress(privateKey.PublicKey)

	d := &Daemon{
		rpc:               rpcURL,
		key:               privateKey,
		addr:              addr,
		heartbeatInterval: interval,
		registry:          common.HexToAddress(registryAddr),
		logger:            logger,
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := d.Run(ctx); err != nil && ctx.Err() == nil {
		logger.Error("daemon exited with error", "error", err)
		os.Exit(1)
	}

	logger.Info("daemon stopped gracefully")
}

func newLogger(level string) *slog.Logger {
	var lv slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lv = slog.LevelDebug
	case "warn":
		lv = slog.LevelWarn
	case "error":
		lv = slog.LevelError
	default:
		lv = slog.LevelInfo
	}
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: lv}))
}

func (d *Daemon) Run(ctx context.Context) error {
	d.logger.Info("starting agent-daemon",
		"address", d.addr.Hex(),
		"rpc", d.rpc,
		"heartbeat_interval", d.heartbeatInterval,
	)

	client, err := ethclient.DialContext(ctx, d.rpc)
	if err != nil {
		return fmt.Errorf("dial rpc: %w", err)
	}
	defer client.Close()
	d.client = client

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("get chain id: %w", err)
	}
	d.chainID = chainID
	d.logger.Info("connected to chain", "chain_id", chainID)

	registered, err := d.checkIsAgent(ctx)
	if err != nil {
		d.logger.Warn("failed to check agent registration", "error", err)
	} else if !registered {
		d.logger.Warn("account is NOT registered as an agent — heartbeats may fail")
	} else {
		d.logger.Info("account is a registered agent")
	}

	return d.loop(ctx)
}

func (d *Daemon) loop(ctx context.Context) error {
	const maxBackoff = 300 * time.Second

	var lastHeartbeatBlock uint64
	var rpcBackoff = pollInterval
	heartbeatBackoff := pollInterval
	lastHeartbeatAttempt := time.Now().Add(-maxBackoff)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		blockNum, err := d.client.BlockNumber(ctx)
		if err != nil {
			d.logger.Error("failed to get block number", "error", err, "retry_in", rpcBackoff)
			sleepCtx(ctx, rpcBackoff)
			rpcBackoff *= 2
			if rpcBackoff > maxBackoff {
				rpcBackoff = maxBackoff
			}
			continue
		}
		rpcBackoff = pollInterval

		d.logger.Debug("polled block", "block", blockNum)

		if lastHeartbeatBlock == 0 || blockNum-lastHeartbeatBlock >= d.heartbeatInterval {
			if heartbeatBackoff > pollInterval && time.Since(lastHeartbeatAttempt) < heartbeatBackoff {
				// Still in backoff period
			} else {
				d.logger.Info("heartbeat due", "block", blockNum, "last", lastHeartbeatBlock)
				lastHeartbeatAttempt = time.Now()
				if err := d.sendHeartbeat(ctx); err != nil {
					d.logger.Error("heartbeat failed", "error", err, "retry_in", heartbeatBackoff)
					heartbeatBackoff *= 2
					if heartbeatBackoff > maxBackoff {
						heartbeatBackoff = maxBackoff
					}
				} else {
					lastHeartbeatBlock = blockNum
					heartbeatBackoff = pollInterval
				}
			}
		}

		sleepCtx(ctx, pollInterval)
	}
}

func (d *Daemon) sendHeartbeat(ctx context.Context) error {
	nonce, err := d.client.PendingNonceAt(ctx, d.addr)
	if err != nil {
		return fmt.Errorf("get nonce: %w", err)
	}

	gasPrice, err := d.client.SuggestGasPrice(ctx)
	if err != nil {
		return fmt.Errorf("suggest gas price: %w", err)
	}

	calldata := common.FromHex(heartbeatSig)

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &d.registry,
		Value:    big.NewInt(0),
		Gas:      100_000,
		GasPrice: gasPrice,
		Data:     calldata,
	})

	signer := types.NewEIP155Signer(d.chainID)
	signedTx, err := types.SignTx(tx, signer, d.key)
	if err != nil {
		return fmt.Errorf("sign tx: %w", err)
	}

	if err := d.client.SendTransaction(ctx, signedTx); err != nil {
		return fmt.Errorf("send tx: %w", err)
	}

	txHash := signedTx.Hash().Hex()
	d.logger.Info("heartbeat tx sent", "tx", txHash, "nonce", nonce)

	receipt, err := d.waitReceipt(ctx, signedTx.Hash())
	if err != nil {
		return fmt.Errorf("wait receipt for %s: %w", txHash, err)
	}

	if receipt.Status == types.ReceiptStatusSuccessful {
		d.logger.Info("heartbeat confirmed", "tx", txHash, "block", receipt.BlockNumber, "gas_used", receipt.GasUsed)
		return nil
	}

	return fmt.Errorf("heartbeat transaction reverted at block %d", receipt.BlockNumber)
}

func (d *Daemon) waitReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	deadline := time.After(receiptTimeout)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-deadline:
			return nil, fmt.Errorf("timeout waiting for receipt")
		default:
		}

		receipt, err := d.client.TransactionReceipt(ctx, txHash)
		if err == nil {
			return receipt, nil
		}
		sleepCtx(ctx, receiptPollWait)
	}
}

// checkIsAgent calls isAgent(address) on the registry precompile.
func (d *Daemon) checkIsAgent(ctx context.Context) (bool, error) {
	selector := common.FromHex(isAgentSig)
	addrPadded := common.LeftPadBytes(d.addr.Bytes(), 32)
	calldata := append(selector, addrPadded...)

	msg := ethereum.CallMsg{
		To:   &d.registry,
		Data: calldata,
	}

	result, err := d.client.CallContract(ctx, msg, nil)
	if err != nil {
		return false, err
	}

	if len(result) < 32 {
		return false, fmt.Errorf("unexpected result length: %d", len(result))
	}

	return new(big.Int).SetBytes(result).Sign() != 0, nil
}

func sleepCtx(ctx context.Context, dur time.Duration) {
	t := time.NewTimer(dur)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}
