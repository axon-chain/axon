package app

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	nodev1beta1 "cosmossdk.io/api/cosmos/base/node/v1beta1"
	tendermintv1beta1 "cosmossdk.io/api/cosmos/base/tendermint/v1beta1"
	sdkmath "cosmossdk.io/math"

	cmttypes "github.com/cometbft/cometbft/rpc/core/types"
	cmttmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/server/api"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/version"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/gorilla/mux"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"

	axonconfig "github.com/axon-chain/axon/app/config"
	agenttypes "github.com/axon-chain/axon/x/agent/types"
	feemarkettypes "github.com/cosmos/evm/x/feemarket/types"
)

const (
	publicAPIRoot                 = "/axon/public/v1"
	publicAPIDefaultLimit         = 20
	publicAPIMaxLimit             = 100
	publicAPIMaxCacheEntries      = 10000
	publicAPIShortCacheTTL        = 5 * time.Second
	publicAPIExplorerCacheTTL     = 8 * time.Second
	publicAPIChainInfoCacheTTL    = 30 * time.Second
	publicAPIChainParamsCacheTTL  = 30 * time.Second
	publicAPIValidatorCacheTTL    = 10 * time.Second
	publicAPIAgentListCacheTTL    = 8 * time.Second
	publicAPISearchCacheTTL       = 5 * time.Second
	publicAPIAllValidatorFetchMax = 500
)

type publicAPI struct {
	clientCtx        client.Context
	tendermintClient tendermintv1beta1.ServiceClient
	nodeClient       nodev1beta1.ServiceClient
	txClient         txtypes.ServiceClient
	authClient       authtypes.QueryClient
	bankClient       banktypes.QueryClient
	stakingClient    stakingtypes.QueryClient
	distribution     distrtypes.QueryClient
	slashingClient   slashingtypes.QueryClient
	govClient        govv1.QueryClient
	feeMarketClient  feemarkettypes.QueryClient
	agentClient      agenttypes.QueryClient

	cacheMu sync.RWMutex
	cache   map[string]publicAPICacheEntry
}

type publicAPICacheEntry struct {
	expiresAt   time.Time
	generatedAt time.Time
	payload     json.RawMessage
}

type publicAPISuccessResponse struct {
	Source      string          `json:"source"`
	GeneratedAt time.Time       `json:"generated_at"`
	Data        json.RawMessage `json:"data"`
}

type publicAPIErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type publicAPIChainInfo struct {
	ChainID        string                 `json:"chain_id"`
	Network        string                 `json:"network"`
	AppName        string                 `json:"app_name"`
	Version        string                 `json:"version"`
	CosmosSDK      string                 `json:"cosmos_sdk_version,omitempty"`
	EVMChainID     uint64                 `json:"evm_chain_id"`
	NativeToken    map[string]any         `json:"native_token"`
	AddressPrefix  map[string]string      `json:"address_prefix"`
	AdditionalInfo map[string]interface{} `json:"additional_info,omitempty"`
}

type publicAPIChainStatus struct {
	LatestBlockHeight string              `json:"latest_block_height"`
	LatestBlockTime   time.Time           `json:"latest_block_time"`
	CatchingUp        bool                `json:"catching_up"`
	ClientName        string              `json:"client_name"`
	PeerCount         int                 `json:"peer_count"`
	Peers             []publicAPIPeerInfo `json:"peers"`
	NodeVersion       string              `json:"node_version"`
	AppVersion        string              `json:"app_version"`
	TxIndexEnabled    bool                `json:"tx_index_enabled"`
}

type publicAPIPeerInfo struct {
	NodeID     string `json:"node_id"`
	Name       string `json:"name"`
	Moniker    string `json:"moniker"`
	RemoteIP   string `json:"remote_ip"`
	Network    string `json:"network"`
	IsOutbound bool   `json:"is_outbound"`
}

// clientName builds a geth-style client identifier:
//
//	axond/v1.1.1-abc123/linux-amd64/go1.22.1
func clientName() string {
	v := version.Version
	if v == "" {
		v = "dev"
	}
	commit := version.Commit
	if len(commit) > 8 {
		commit = commit[:8]
	}
	if commit != "" {
		v += "-" + commit
	}
	return fmt.Sprintf("axond/%s/%s-%s/%s", v, runtime.GOOS, runtime.GOARCH, runtime.Version())
}

// parseMonikerClientName splits a moniker that may contain an injected client
// name suffix. Format: "my-validator axond/v1.1.1/os-arch/goX.Y"
// Returns (original_moniker, client_name). If no client name is found,
// client_name is empty.
func parseMonikerClientName(raw string) (moniker, name string) {
	if idx := strings.Index(raw, " axond/"); idx >= 0 {
		return raw[:idx], raw[idx+1:]
	}
	return raw, ""
}

type publicAPIValidatorSummary struct {
	Moniker         string      `json:"moniker"`
	OperatorAddress string      `json:"operator_address"`
	BondedStatus    string      `json:"bonded_status"`
	Jailed          bool        `json:"jailed"`
	Tokens          interface{} `json:"tokens"`
	VotingPower     int64       `json:"voting_power"`
	CommissionRate  string      `json:"commission_rate"`
}

type publicAPIAgentSummary struct {
	AgentAddress         string      `json:"agent_address"`
	ValidatorAddress     string      `json:"validator_address"`
	AgentID              string      `json:"agent_id"`
	Status               string      `json:"status"`
	Model                string      `json:"model"`
	Capabilities         []string    `json:"capabilities"`
	ReputationScore      uint64      `json:"reputation_score"`
	AgentStake           interface{} `json:"agent_stake"`
	RegisteredAt         int64       `json:"registered_at"`
	LastHeartbeatHeight  int64       `json:"last_heartbeat_height"`
	LastHeartbeatTime    *time.Time  `json:"last_heartbeat_time,omitempty"`
	RemainingOfflineBars int64       `json:"remaining_blocks_until_offline,omitempty"`
}

type publicAPIPagination struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
	Total  int `json:"total"`
}

func registerPublicAPIRoutes(apiSvr *api.Server) {
	publicAPI := newPublicAPI(apiSvr.ClientCtx)

	router := apiSvr.Router.PathPrefix(publicAPIRoot).Subrouter()
	router.HandleFunc("/chain/info", publicAPI.cachedHandler("chain-info", publicAPIChainInfoCacheTTL, publicAPI.handleChainInfo)).Methods(http.MethodGet)
	router.HandleFunc("/chain/status", publicAPI.cachedHandler("chain-status", publicAPIShortCacheTTL, publicAPI.handleChainStatus)).Methods(http.MethodGet)
	router.HandleFunc("/chain/health", publicAPI.cachedHandler("chain-health", publicAPIShortCacheTTL, publicAPI.handleChainHealth)).Methods(http.MethodGet)
	router.HandleFunc("/chain/fees", publicAPI.cachedHandler("chain-fees", publicAPIShortCacheTTL, publicAPI.handleChainFees)).Methods(http.MethodGet)
	router.HandleFunc("/chain/params", publicAPI.cachedHandler("chain-params", publicAPIChainParamsCacheTTL, publicAPI.handleChainParams)).Methods(http.MethodGet)
	router.HandleFunc("/params/staking", publicAPI.cachedHandler("params-staking", publicAPIChainParamsCacheTTL, publicAPI.handleStakingParams)).Methods(http.MethodGet)
	router.HandleFunc("/params/slashing", publicAPI.cachedHandler("params-slashing", publicAPIChainParamsCacheTTL, publicAPI.handleSlashingParams)).Methods(http.MethodGet)
	router.HandleFunc("/params/distribution", publicAPI.cachedHandler("params-distribution", publicAPIChainParamsCacheTTL, publicAPI.handleDistributionParams)).Methods(http.MethodGet)
	router.HandleFunc("/params/agent", publicAPI.cachedHandler("params-agent", publicAPIChainParamsCacheTTL, publicAPI.handleAgentParams)).Methods(http.MethodGet)

	router.HandleFunc("/blocks/latest", publicAPI.cachedHandler("blocks-latest", publicAPIShortCacheTTL, publicAPI.handleLatestBlock)).Methods(http.MethodGet)
	router.HandleFunc("/blocks/{identifier}", publicAPI.cachedHandler("blocks-get", publicAPIShortCacheTTL, publicAPI.handleBlock)).Methods(http.MethodGet)
	router.HandleFunc("/blocks", publicAPI.cachedHandler("blocks-list", publicAPIExplorerCacheTTL, publicAPI.handleBlocksRange)).Methods(http.MethodGet)
	router.HandleFunc("/blocks/{height:[0-9]+}/txs", publicAPI.cachedHandler("blocks-txs", publicAPIShortCacheTTL, publicAPI.handleBlockTxs)).Methods(http.MethodGet)
	router.HandleFunc("/blocks/{height:[0-9]+}/validators", publicAPI.cachedHandler("blocks-validators", publicAPIShortCacheTTL, publicAPI.handleBlockValidators)).Methods(http.MethodGet)
	router.HandleFunc("/blocks/{height:[0-9]+}/proposer", publicAPI.cachedHandler("blocks-proposer", publicAPIShortCacheTTL, publicAPI.handleBlockProposer)).Methods(http.MethodGet)

	router.HandleFunc("/txs/recent", publicAPI.cachedHandler("txs-recent", publicAPIExplorerCacheTTL, publicAPI.handleRecentTxs)).Methods(http.MethodGet)
	router.HandleFunc("/txs/search", publicAPI.cachedHandler("txs-search", publicAPISearchCacheTTL, publicAPI.handleTxSearch)).Methods(http.MethodGet)
	router.HandleFunc("/txs/{hash}/events", publicAPI.cachedHandler("txs-events", publicAPIShortCacheTTL, publicAPI.handleTxEvents)).Methods(http.MethodGet)
	router.HandleFunc("/txs/{hash}/raw", publicAPI.cachedHandler("txs-raw", publicAPIShortCacheTTL, publicAPI.handleTxRaw)).Methods(http.MethodGet)
	router.HandleFunc("/txs/{hash}", publicAPI.cachedHandler("txs-get", publicAPIShortCacheTTL, publicAPI.handleTx)).Methods(http.MethodGet)
	router.HandleFunc("/txs", publicAPI.cachedHandler("txs-list", publicAPISearchCacheTTL, publicAPI.handleTxs)).Methods(http.MethodGet)
	router.HandleFunc("/txs/simulate", publicAPI.handleTxSimulate).Methods(http.MethodPost)
	router.HandleFunc("/txs/broadcast", publicAPI.handleTxBroadcast).Methods(http.MethodPost)

	router.HandleFunc("/accounts/{address}", publicAPI.cachedHandler("accounts-get", publicAPIShortCacheTTL, publicAPI.handleAccount)).Methods(http.MethodGet)
	router.HandleFunc("/accounts/{address}/balances", publicAPI.cachedHandler("accounts-balances", publicAPIShortCacheTTL, publicAPI.handleAccountBalances)).Methods(http.MethodGet)
	router.HandleFunc("/accounts/{address}/spendable", publicAPI.cachedHandler("accounts-spendable", publicAPIShortCacheTTL, publicAPI.handleAccountSpendable)).Methods(http.MethodGet)
	router.HandleFunc("/accounts/{address}/sequence", publicAPI.cachedHandler("accounts-sequence", publicAPIShortCacheTTL, publicAPI.handleAccountSequence)).Methods(http.MethodGet)
	router.HandleFunc("/accounts/{address}/txs", publicAPI.cachedHandler("accounts-txs", publicAPISearchCacheTTL, publicAPI.handleAccountTxs)).Methods(http.MethodGet)
	router.HandleFunc("/accounts/{address}/transfers", publicAPI.cachedHandler("accounts-transfers", publicAPISearchCacheTTL, publicAPI.handleAccountTransfers)).Methods(http.MethodGet)
	router.HandleFunc("/accounts/{address}/rewards", publicAPI.cachedHandler("accounts-rewards", publicAPIShortCacheTTL, publicAPI.handleAccountRewards)).Methods(http.MethodGet)

	router.HandleFunc("/validators", publicAPI.cachedHandler("validators-list", publicAPIValidatorCacheTTL, publicAPI.handleValidators)).Methods(http.MethodGet)
	router.HandleFunc("/validators/top", publicAPI.cachedHandler("validators-top", publicAPIValidatorCacheTTL, publicAPI.handleValidatorsTop)).Methods(http.MethodGet)
	router.HandleFunc("/validators/{valoper}", publicAPI.cachedHandler("validator-detail", publicAPIValidatorCacheTTL, publicAPI.handleValidator)).Methods(http.MethodGet)
	router.HandleFunc("/validators/{valoper}/status", publicAPI.cachedHandler("validator-status", publicAPIValidatorCacheTTL, publicAPI.handleValidatorStatus)).Methods(http.MethodGet)
	router.HandleFunc("/validators/{valoper}/delegations", publicAPI.cachedHandler("validator-delegations", publicAPIValidatorCacheTTL, publicAPI.handleValidatorDelegations)).Methods(http.MethodGet)
	router.HandleFunc("/validators/{valoper}/unbondings", publicAPI.cachedHandler("validator-unbondings", publicAPIValidatorCacheTTL, publicAPI.handleValidatorUnbondings)).Methods(http.MethodGet)
	router.HandleFunc("/validators/{valoper}/redelegations", publicAPI.cachedHandler("validator-redelegations", publicAPIValidatorCacheTTL, publicAPI.handleValidatorRedelegations)).Methods(http.MethodGet)
	router.HandleFunc("/validators/{valoper}/commission", publicAPI.cachedHandler("validator-commission", publicAPIValidatorCacheTTL, publicAPI.handleValidatorCommission)).Methods(http.MethodGet)
	router.HandleFunc("/validators/{valoper}/rewards", publicAPI.cachedHandler("validator-rewards", publicAPIValidatorCacheTTL, publicAPI.handleValidatorRewards)).Methods(http.MethodGet)
	router.HandleFunc("/validators/{valoper}/slashes", publicAPI.cachedHandler("validator-slashes", publicAPIValidatorCacheTTL, publicAPI.handleValidatorSlashes)).Methods(http.MethodGet)
	router.HandleFunc("/validators/{valoper}/signing-info", publicAPI.cachedHandler("validator-signing-info", publicAPIValidatorCacheTTL, publicAPI.handleValidatorSigningInfo)).Methods(http.MethodGet)
	router.HandleFunc("/validators/{valoper}/self-delegation", publicAPI.cachedHandler("validator-self-delegation", publicAPIValidatorCacheTTL, publicAPI.handleValidatorSelfDelegation)).Methods(http.MethodGet)
	router.HandleFunc("/validators/{valoper}/agent", publicAPI.cachedHandler("validator-agent", publicAPIValidatorCacheTTL, publicAPI.handleValidatorAgent)).Methods(http.MethodGet)
	router.HandleFunc("/validators/{valoper}/uptime", publicAPI.cachedHandler("validator-uptime", publicAPIValidatorCacheTTL, publicAPI.handleValidatorUptime)).Methods(http.MethodGet)

	router.HandleFunc("/delegators/{address}/delegations", publicAPI.cachedHandler("delegator-delegations", publicAPIShortCacheTTL, publicAPI.handleDelegatorDelegations)).Methods(http.MethodGet)
	router.HandleFunc("/delegators/{address}/unbondings", publicAPI.cachedHandler("delegator-unbondings", publicAPIShortCacheTTL, publicAPI.handleDelegatorUnbondings)).Methods(http.MethodGet)
	router.HandleFunc("/delegators/{address}/redelegations", publicAPI.cachedHandler("delegator-redelegations", publicAPIShortCacheTTL, publicAPI.handleDelegatorRedelegations)).Methods(http.MethodGet)
	router.HandleFunc("/delegators/{address}/rewards", publicAPI.cachedHandler("delegator-rewards", publicAPIShortCacheTTL, publicAPI.handleDelegatorRewards)).Methods(http.MethodGet)
	router.HandleFunc("/delegators/{address}/withdraw-address", publicAPI.cachedHandler("delegator-withdraw-address", publicAPIShortCacheTTL, publicAPI.handleDelegatorWithdrawAddress)).Methods(http.MethodGet)
	router.HandleFunc("/delegators/{address}/validators", publicAPI.cachedHandler("delegator-validators", publicAPIShortCacheTTL, publicAPI.handleDelegatorValidators)).Methods(http.MethodGet)

	router.HandleFunc("/gov/proposals", publicAPI.cachedHandler("gov-proposals", publicAPIShortCacheTTL, publicAPI.handleGovProposals)).Methods(http.MethodGet)
	router.HandleFunc("/gov/proposals/{id:[0-9]+}", publicAPI.cachedHandler("gov-proposal", publicAPIShortCacheTTL, publicAPI.handleGovProposal)).Methods(http.MethodGet)
	router.HandleFunc("/gov/proposals/{id:[0-9]+}/votes", publicAPI.cachedHandler("gov-proposal-votes", publicAPIShortCacheTTL, publicAPI.handleGovProposalVotes)).Methods(http.MethodGet)
	router.HandleFunc("/gov/proposals/{id:[0-9]+}/tally", publicAPI.cachedHandler("gov-proposal-tally", publicAPIShortCacheTTL, publicAPI.handleGovProposalTally)).Methods(http.MethodGet)
	router.HandleFunc("/gov/params", publicAPI.cachedHandler("gov-params", publicAPIChainParamsCacheTTL, publicAPI.handleGovParams)).Methods(http.MethodGet)

	router.HandleFunc("/explorer/overview", publicAPI.cachedHandler("explorer-overview", publicAPIExplorerCacheTTL, publicAPI.handleExplorerOverview)).Methods(http.MethodGet)
	router.HandleFunc("/explorer/stats", publicAPI.cachedHandler("explorer-stats", publicAPIExplorerCacheTTL, publicAPI.handleExplorerStats)).Methods(http.MethodGet)
	router.HandleFunc("/explorer/validators/top", publicAPI.cachedHandler("explorer-validators-top", publicAPIExplorerCacheTTL, publicAPI.handleValidatorsTop)).Methods(http.MethodGet)
	router.HandleFunc("/explorer/blocks/recent", publicAPI.cachedHandler("explorer-blocks-recent", publicAPIExplorerCacheTTL, publicAPI.handleRecentBlocks)).Methods(http.MethodGet)
	router.HandleFunc("/explorer/txs/recent", publicAPI.cachedHandler("explorer-txs-recent", publicAPIExplorerCacheTTL, publicAPI.handleRecentTxs)).Methods(http.MethodGet)

	router.HandleFunc("/agents/online-validators", publicAPI.cachedHandler("agents-online-validators", publicAPIAgentListCacheTTL, publicAPI.handleOnlineValidators)).Methods(http.MethodGet)
	router.HandleFunc("/agents/challenge/current", publicAPI.cachedHandler("agents-challenge-current", publicAPIShortCacheTTL, publicAPI.handleCurrentChallenge)).Methods(http.MethodGet)
	router.HandleFunc("/agents/{address}/heartbeat", publicAPI.cachedHandler("agent-heartbeat", publicAPIShortCacheTTL, publicAPI.handleAgentHeartbeat)).Methods(http.MethodGet)
	router.HandleFunc("/agents/{address}/reputation", publicAPI.cachedHandler("agent-reputation", publicAPIShortCacheTTL, publicAPI.handleAgentReputation)).Methods(http.MethodGet)
	router.HandleFunc("/agents/{address}/stake", publicAPI.cachedHandler("agent-stake", publicAPIShortCacheTTL, publicAPI.handleAgentStake)).Methods(http.MethodGet)
	router.HandleFunc("/agents/{address}", publicAPI.cachedHandler("agent-detail", publicAPIValidatorCacheTTL, publicAPI.handleAgent)).Methods(http.MethodGet)
	router.HandleFunc("/agents", publicAPI.cachedHandler("agents-list", publicAPIAgentListCacheTTL, publicAPI.handleAgents)).Methods(http.MethodGet)

	router.HandleFunc("/search", publicAPI.cachedHandler("search", publicAPISearchCacheTTL, publicAPI.handleSearch)).Methods(http.MethodGet)

	// Rewrite gRPC-gateway 501 (UNIMPLEMENTED) to 404 for unknown paths.
	apiSvr.Router.Use(rewriteUnimplementedMiddleware)
}

// rewriteUnimplementedMiddleware intercepts HTTP 501 responses produced by the
// gRPC-gateway for paths it cannot route and rewrites them to 404.
func rewriteUnimplementedMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &notImplementedInterceptor{ResponseWriter: w}
		next.ServeHTTP(rec, r)
		if rec.suppressed {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"endpoint not found"}`))
		}
	})
}

// notImplementedInterceptor wraps http.ResponseWriter to suppress 501 responses.
type notImplementedInterceptor struct {
	http.ResponseWriter
	suppressed bool
	committed  bool
}

func (w *notImplementedInterceptor) WriteHeader(code int) {
	if code == http.StatusNotImplemented {
		w.suppressed = true
		return
	}
	w.committed = true
	w.ResponseWriter.WriteHeader(code)
}

func (w *notImplementedInterceptor) Write(b []byte) (int, error) {
	if w.suppressed {
		return len(b), nil
	}
	if !w.committed {
		w.committed = true
		w.ResponseWriter.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

func newPublicAPI(clientCtx client.Context) *publicAPI {
	return &publicAPI{
		clientCtx:        clientCtx,
		tendermintClient: tendermintv1beta1.NewServiceClient(clientCtx),
		nodeClient:       nodev1beta1.NewServiceClient(clientCtx),
		txClient:         txtypes.NewServiceClient(clientCtx),
		authClient:       authtypes.NewQueryClient(clientCtx),
		bankClient:       banktypes.NewQueryClient(clientCtx),
		stakingClient:    stakingtypes.NewQueryClient(clientCtx),
		distribution:     distrtypes.NewQueryClient(clientCtx),
		slashingClient:   slashingtypes.NewQueryClient(clientCtx),
		govClient:        govv1.NewQueryClient(clientCtx),
		feeMarketClient:  feemarkettypes.NewQueryClient(clientCtx),
		agentClient:      agenttypes.NewQueryClient(clientCtx),
		cache:            make(map[string]publicAPICacheEntry),
	}
}

func (p *publicAPI) cachedHandler(namespace string, ttl time.Duration, loader func(context.Context, *http.Request) (any, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := namespace + ":" + r.URL.RequestURI()
		payload, source, generatedAt, err := p.loadCachedPayload(r.Context(), key, ttl, func(ctx context.Context) (any, error) {
			return loader(ctx, r)
		})
		if err != nil {
			p.writeError(w, err)
			return
		}

		p.writeSuccess(w, payload, source, generatedAt)
	}
}

func (p *publicAPI) loadCachedPayload(ctx context.Context, key string, ttl time.Duration, loader func(context.Context) (any, error)) (json.RawMessage, string, time.Time, error) {
	if ttl > 0 {
		if payload, generatedAt, ok := p.readCache(key); ok {
			return payload, "cache", generatedAt, nil
		}
	}

	value, err := loader(ctx)
	if err != nil {
		return nil, "", time.Time{}, err
	}

	payload, err := json.Marshal(value)
	if err != nil {
		return nil, "", time.Time{}, fmt.Errorf("failed to marshal response: %w", err)
	}

	generatedAt := time.Now().UTC()
	if ttl > 0 {
		p.writeCache(key, ttl, payload, generatedAt)
	}

	return payload, "node", generatedAt, nil
}

func (p *publicAPI) readCache(key string) (json.RawMessage, time.Time, bool) {
	now := time.Now()

	p.cacheMu.RLock()
	entry, ok := p.cache[key]
	p.cacheMu.RUnlock()
	if !ok || now.After(entry.expiresAt) {
		if ok {
			p.cacheMu.Lock()
			delete(p.cache, key)
			p.cacheMu.Unlock()
		}
		return nil, time.Time{}, false
	}

	return entry.payload, entry.generatedAt, true
}

func (p *publicAPI) writeCache(key string, ttl time.Duration, payload json.RawMessage, generatedAt time.Time) {
	p.cacheMu.Lock()
	now := time.Now()
	if len(p.cache) >= publicAPIMaxCacheEntries {
		for cacheKey, entry := range p.cache {
			if now.After(entry.expiresAt) {
				delete(p.cache, cacheKey)
			}
		}
	}
	if len(p.cache) >= publicAPIMaxCacheEntries {
		p.cacheMu.Unlock()
		return
	}
	p.cache[key] = publicAPICacheEntry{
		expiresAt:   now.Add(ttl),
		generatedAt: generatedAt,
		payload:     payload,
	}
	p.cacheMu.Unlock()
}

func (p *publicAPI) writeSuccess(w http.ResponseWriter, payload json.RawMessage, source string, generatedAt time.Time) {
	w.Header().Set("Content-Type", "application/json")
	response := publicAPISuccessResponse{
		Source:      source,
		GeneratedAt: generatedAt,
		Data:        payload,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (p *publicAPI) writeError(w http.ResponseWriter, err error) {
	statusCode := http.StatusInternalServerError
	code := "internal_error"
	message := err.Error()

	switch {
	case errors.Is(err, errBadRequest):
		statusCode = http.StatusBadRequest
		code = "bad_request"
	case errors.Is(err, errNotFound):
		statusCode = http.StatusNotFound
		code = "not_found"
	default:
		if grpcErr, ok := grpcstatus.FromError(err); ok {
			switch grpcErr.Code() {
			case codes.InvalidArgument:
				statusCode = http.StatusBadRequest
				code = "bad_request"
			case codes.NotFound:
				statusCode = http.StatusNotFound
				code = "not_found"
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(publicAPIErrorResponse{
		Code:    code,
		Message: message,
	})
}

var (
	errBadRequest = errors.New("bad request")
	errNotFound   = errors.New("not found")
)

func (p *publicAPI) handleChainInfo(ctx context.Context, _ *http.Request) (any, error) {
	nodeInfo, err := p.tendermintClient.GetNodeInfo(ctx, &tendermintv1beta1.GetNodeInfoRequest{})
	if err != nil {
		return nil, err
	}

	versionInfo := nodeInfo.GetApplicationVersion()
	if versionInfo == nil {
		versionInfo = &tendermintv1beta1.VersionInfo{}
	}
	chainID := p.clientCtx.ChainID
	if chainID == "" && nodeInfo.GetDefaultNodeInfo() != nil {
		chainID = nodeInfo.GetDefaultNodeInfo().Network
	}

	result := publicAPIChainInfo{
		ChainID:    chainID,
		Network:    inferNetworkName(chainID),
		AppName:    firstNonEmpty(versionInfo.GetAppName(), AppName),
		Version:    firstNonEmpty(versionInfo.GetVersion(), version.Version),
		CosmosSDK:  versionInfo.GetCosmosSdkVersion(),
		EVMChainID: axonconfig.EVMChainID,
		NativeToken: map[string]any{
			"base_denom":    axonconfig.AxonDenom,
			"display_denom": axonconfig.HumanDenom,
			"decimals":      18,
		},
		AddressPrefix: map[string]string{
			"account":   axonconfig.Bech32PrefixAccAddr,
			"validator": axonconfig.Bech32PrefixValAddr,
			"consensus": axonconfig.Bech32PrefixConsAddr,
		},
	}

	return result, nil
}

func (p *publicAPI) handleChainStatus(ctx context.Context, _ *http.Request) (any, error) {
	status, err := p.clientCtx.Client.Status(ctx)
	if err != nil {
		return nil, err
	}

	nodeInfoResp, err := p.tendermintClient.GetNodeInfo(ctx, &tendermintv1beta1.GetNodeInfoRequest{})
	if err != nil {
		return nil, err
	}
	appVersion := ""
	if nodeInfoResp.GetApplicationVersion() != nil {
		appVersion = nodeInfoResp.GetApplicationVersion().GetVersion()
	}

	peerCount := 0
	var peers []publicAPIPeerInfo
	if netClient, ok := p.clientCtx.Client.(interface {
		NetInfo(context.Context) (*cmttypes.ResultNetInfo, error)
	}); ok {
		netInfo, netErr := netClient.NetInfo(ctx)
		if netErr == nil {
			peerCount = netInfo.NPeers
			for _, peer := range netInfo.Peers {
				moniker, name := parseMonikerClientName(peer.NodeInfo.Moniker)
				peers = append(peers, publicAPIPeerInfo{
					NodeID:     string(peer.NodeInfo.DefaultNodeID),
					Name:       name,
					Moniker:    moniker,
					RemoteIP:   peer.RemoteIP,
					Network:    peer.NodeInfo.Network,
					IsOutbound: peer.IsOutbound,
				})
			}
		}
	}

	return publicAPIChainStatus{
		LatestBlockHeight: strconv.FormatInt(status.SyncInfo.LatestBlockHeight, 10),
		LatestBlockTime:   status.SyncInfo.LatestBlockTime,
		CatchingUp:        status.SyncInfo.CatchingUp,
		ClientName:        clientName(),
		PeerCount:         peerCount,
		Peers:             peers,
		NodeVersion:       status.NodeInfo.Version,
		AppVersion:        appVersion,
		TxIndexEnabled:    status.TxIndexEnabled(),
	}, nil
}

func (p *publicAPI) handleChainHealth(ctx context.Context, _ *http.Request) (any, error) {
	rpcAvailable := false
	peerCount := 0
	if _, err := p.clientCtx.Client.Status(ctx); err == nil {
		rpcAvailable = true
	}

	if netClient, ok := p.clientCtx.Client.(interface {
		NetInfo(context.Context) (*cmttypes.ResultNetInfo, error)
	}); ok {
		netInfo, err := netClient.NetInfo(ctx)
		if err == nil {
			peerCount = netInfo.NPeers
		}
	}

	healthy := rpcAvailable
	if healthClient, ok := p.clientCtx.Client.(interface {
		Health(context.Context) (*cmttypes.ResultHealth, error)
	}); ok {
		_, err := healthClient.Health(ctx)
		healthy = err == nil
	}

	return map[string]any{
		"healthy":             healthy,
		"rpc_available":       rpcAvailable,
		"rest_available":      true,
		"grpc_gateway_alive":  true,
		"peer_count":          peerCount,
		"checked_at":          time.Now().UTC(),
		"custom_public_api":   true,
		"standard_cosmos_api": true,
	}, nil
}

func (p *publicAPI) handleChainFees(ctx context.Context, _ *http.Request) (any, error) {
	paramsResp, err := p.feeMarketClient.Params(ctx, &feemarkettypes.QueryParamsRequest{})
	if err != nil {
		return nil, err
	}

	baseFeeResp, err := p.feeMarketClient.BaseFee(ctx, &feemarkettypes.QueryBaseFeeRequest{})
	if err != nil {
		return nil, err
	}

	baseFee := ""
	if baseFeeResp.BaseFee != nil {
		baseFee = baseFeeResp.BaseFee.String()
	}

	minGasPrice := paramsResp.Params.MinGasPrice.String()

	return map[string]any{
		"base_fee": map[string]any{
			"amount": baseFee,
			"denom":  axonconfig.AxonDenom,
		},
		"minimum_gas_price": map[string]any{
			"amount": minGasPrice,
			"denom":  axonconfig.AxonDenom,
		},
		"recommended_gas_price": map[string]any{
			"slow": map[string]any{
				"amount": minGasPrice,
				"denom":  axonconfig.AxonDenom,
			},
			"standard": map[string]any{
				"amount": preferredGasPrice(minGasPrice, baseFee),
				"denom":  axonconfig.AxonDenom,
			},
			"fast": map[string]any{
				"amount": preferredGasPrice(minGasPrice, baseFee),
				"denom":  axonconfig.AxonDenom,
			},
		},
		"estimation_note": "Use /cosmos/tx/v1beta1/simulate for transaction-specific gas estimation.",
	}, nil
}

func (p *publicAPI) handleChainParams(ctx context.Context, _ *http.Request) (any, error) {
	stakingParams, err := p.stakingClient.Params(ctx, &stakingtypes.QueryParamsRequest{})
	if err != nil {
		return nil, err
	}

	slashingParams, err := p.slashingClient.Params(ctx, &slashingtypes.QueryParamsRequest{})
	if err != nil {
		return nil, err
	}

	distributionParams, err := p.distribution.Params(ctx, &distrtypes.QueryParamsRequest{})
	if err != nil {
		return nil, err
	}

	agentParams, err := p.agentClient.Params(ctx, &agenttypes.QueryParamsRequest{})
	if err != nil {
		return nil, err
	}

	depositParams, err := p.govClient.Params(ctx, &govv1.QueryParamsRequest{ParamsType: "deposit"})
	if err != nil {
		return nil, err
	}

	votingParams, err := p.govClient.Params(ctx, &govv1.QueryParamsRequest{ParamsType: "voting"})
	if err != nil {
		return nil, err
	}

	tallyParams, err := p.govClient.Params(ctx, &govv1.QueryParamsRequest{ParamsType: "tallying"})
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"staking":      stakingParams.Params,
		"slashing":     slashingParams.Params,
		"distribution": distributionParams.Params,
		"agent":        agentParams.Params,
		"gov": map[string]any{
			"deposit": depositParams.GetParams(),
			"voting":  votingParams.GetParams(),
			"tally":   tallyParams.GetParams(),
		},
	}, nil
}

func (p *publicAPI) handleValidatorsTop(ctx context.Context, r *http.Request) (any, error) {
	limit, err := parseLimit(r, publicAPIDefaultLimit, publicAPIMaxLimit)
	if err != nil {
		return nil, err
	}

	validators, err := p.fetchAllValidators(ctx)
	if err != nil {
		return nil, err
	}

	bonded := make([]stakingtypes.Validator, 0, len(validators))
	for _, validator := range validators {
		if validator.IsBonded() {
			bonded = append(bonded, validator)
		}
	}

	sort.Slice(bonded, func(i, j int) bool {
		return bonded[i].Tokens.GT(bonded[j].Tokens)
	})

	if len(bonded) > limit {
		bonded = bonded[:limit]
	}

	items := make([]publicAPIValidatorSummary, 0, len(bonded))
	for _, validator := range bonded {
		items = append(items, summarizeValidator(validator))
	}

	return map[string]any{
		"validators": items,
		"pagination": publicAPIPagination{
			Offset: 0,
			Limit:  limit,
			Total:  len(items),
		},
	}, nil
}

func (p *publicAPI) handleValidatorStatus(ctx context.Context, r *http.Request) (any, error) {
	validator, err := p.fetchValidator(ctx, mux.Vars(r)["valoper"])
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"validator": summarizeValidator(validator),
		"status": map[string]any{
			"bonded_status": validator.Status.String(),
			"jailed":        validator.Jailed,
			"unbonding_time": func() *time.Time {
				if validator.UnbondingTime.IsZero() {
					return nil
				}
				t := validator.UnbondingTime.UTC()
				return &t
			}(),
		},
	}, nil
}

func (p *publicAPI) handleValidatorSelfDelegation(ctx context.Context, r *http.Request) (any, error) {
	valoper := mux.Vars(r)["valoper"]
	delegator, err := accAddressFromValoper(valoper)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid validator address", errBadRequest)
	}

	response, err := p.stakingClient.Delegation(ctx, &stakingtypes.QueryDelegationRequest{
		DelegatorAddr: delegator,
		ValidatorAddr: valoper,
	})
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"validator_address": valoper,
		"delegator_address": delegator,
		"self_delegation":   response.DelegationResponse,
	}, nil
}

func (p *publicAPI) handleValidatorAgent(ctx context.Context, r *http.Request) (any, error) {
	agent, err := p.fetchAgentByValidator(ctx, mux.Vars(r)["valoper"])
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"agent": agent,
	}, nil
}

func (p *publicAPI) handleValidatorUptime(ctx context.Context, r *http.Request) (any, error) {
	validator, err := p.fetchValidator(ctx, mux.Vars(r)["valoper"])
	if err != nil {
		return nil, err
	}

	signingInfo, slashingParams, err := p.fetchSigningInfo(ctx, validator)
	if err != nil {
		return nil, err
	}

	window := slashingParams.SignedBlocksWindow
	uptime := "1.000000000000000000"
	if window > 0 {
		signedRatio := sdkmath.LegacyOneDec().Sub(sdkmath.LegacyNewDec(signingInfo.MissedBlocksCounter).Quo(sdkmath.LegacyNewDec(window)))
		if signedRatio.IsNegative() {
			signedRatio = sdkmath.LegacyZeroDec()
		}
		uptime = signedRatio.String()
	}

	return map[string]any{
		"validator_address":    validator.OperatorAddress,
		"uptime":               uptime,
		"missed_blocks":        signingInfo.MissedBlocksCounter,
		"signed_blocks_window": slashingParams.SignedBlocksWindow,
		"start_height":         signingInfo.StartHeight,
		"index_offset":         signingInfo.IndexOffset,
		"tombstoned":           signingInfo.Tombstoned,
	}, nil
}

func (p *publicAPI) handleExplorerOverview(ctx context.Context, _ *http.Request) (any, error) {
	status, err := p.clientCtx.Client.Status(ctx)
	if err != nil {
		return nil, err
	}

	validators, err := p.fetchAllValidators(ctx)
	if err != nil {
		return nil, err
	}

	agentsResp, err := p.agentClient.Agents(ctx, &agenttypes.QueryAgentsRequest{})
	if err != nil {
		return nil, err
	}

	bondedValidators := 0
	onlineAgents := 0
	for _, validator := range validators {
		if validator.IsBonded() {
			bondedValidators++
		}
	}
	for _, agent := range agentsResp.Agents {
		if normalizeAgentStatus(agent.Status) == "online" {
			onlineAgents++
		}
	}

	return map[string]any{
		"chain_id":            p.clientCtx.ChainID,
		"network":             inferNetworkName(p.clientCtx.ChainID),
		"latest_block_height": strconv.FormatInt(status.SyncInfo.LatestBlockHeight, 10),
		"latest_block_time":   status.SyncInfo.LatestBlockTime,
		"catching_up":         status.SyncInfo.CatchingUp,
		"validator_count":     len(validators),
		"bonded_validators":   bondedValidators,
		"agent_count":         len(agentsResp.Agents),
		"online_agents":       onlineAgents,
		"native_token":        axonconfig.HumanDenom,
		"evm_chain_id":        axonconfig.EVMChainID,
	}, nil
}

func (p *publicAPI) handleExplorerStats(ctx context.Context, _ *http.Request) (any, error) {
	status, err := p.clientCtx.Client.Status(ctx)
	if err != nil {
		return nil, err
	}

	nodeStatus, err := p.nodeClient.Status(ctx, &nodev1beta1.StatusRequest{})
	if err != nil {
		return nil, err
	}

	latestHeight := status.SyncInfo.LatestBlockHeight
	minHeight := latestHeight - 19
	if minHeight < 1 {
		minHeight = 1
	}

	blockchainInfo, err := p.clientCtx.Client.BlockchainInfo(ctx, minHeight, latestHeight)
	if err != nil {
		return nil, err
	}

	totalRecentTxs := 0
	for _, meta := range blockchainInfo.BlockMetas {
		totalRecentTxs += meta.NumTxs
	}

	peerCount := 0
	if netClient, ok := p.clientCtx.Client.(interface {
		NetInfo(context.Context) (*cmttypes.ResultNetInfo, error)
	}); ok {
		netInfo, netErr := netClient.NetInfo(ctx)
		if netErr == nil {
			peerCount = netInfo.NPeers
		}
	}

	return map[string]any{
		"latest_block_height":   strconv.FormatInt(status.SyncInfo.LatestBlockHeight, 10),
		"latest_block_time":     status.SyncInfo.LatestBlockTime,
		"earliest_store_height": nodeStatus.EarliestStoreHeight,
		"peer_count":            peerCount,
		"recent_block_span":     len(blockchainInfo.BlockMetas),
		"recent_tx_count":       totalRecentTxs,
		"tx_index_enabled":      status.TxIndexEnabled(),
		"app_hash":              hex.EncodeToString(nodeStatus.AppHash),
		"validator_hash":        hex.EncodeToString(nodeStatus.ValidatorHash),
	}, nil
}

func (p *publicAPI) handleRecentBlocks(ctx context.Context, r *http.Request) (any, error) {
	limit, err := parseLimit(r, publicAPIDefaultLimit, publicAPIMaxLimit)
	if err != nil {
		return nil, err
	}

	status, err := p.clientCtx.Client.Status(ctx)
	if err != nil {
		return nil, err
	}

	latestHeight := status.SyncInfo.LatestBlockHeight
	minHeight := latestHeight - int64(limit) + 1
	if minHeight < 1 {
		minHeight = 1
	}

	blockchainInfo, err := p.clientCtx.Client.BlockchainInfo(ctx, minHeight, latestHeight)
	if err != nil {
		return nil, err
	}

	items := make([]map[string]any, 0, len(blockchainInfo.BlockMetas))
	for _, meta := range blockchainInfo.BlockMetas {
		items = append(items, summarizeBlockMeta(meta))
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i]["height"].(int64) > items[j]["height"].(int64)
	})

	return map[string]any{
		"blocks": items,
		"pagination": publicAPIPagination{
			Offset: 0,
			Limit:  limit,
			Total:  len(items),
		},
	}, nil
}

func (p *publicAPI) handleRecentTxs(ctx context.Context, r *http.Request) (any, error) {
	limit, err := parseLimit(r, publicAPIDefaultLimit, publicAPIMaxLimit)
	if err != nil {
		return nil, err
	}

	page := 1
	orderBy := "desc"
	results, err := p.clientCtx.Client.TxSearch(ctx, "tx.height > 0", false, &page, &limit, orderBy)
	if err != nil {
		return nil, err
	}

	items := make([]map[string]any, 0, len(results.Txs))
	for _, txResult := range results.Txs {
		items = append(items, summarizeResultTx(txResult))
	}

	return map[string]any{
		"txs": items,
		"pagination": publicAPIPagination{
			Offset: 0,
			Limit:  limit,
			Total:  results.TotalCount,
		},
	}, nil
}

func (p *publicAPI) handleAgents(ctx context.Context, r *http.Request) (any, error) {
	limit, err := parseLimit(r, publicAPIDefaultLimit, publicAPIMaxLimit)
	if err != nil {
		return nil, err
	}

	offset, err := parseOffset(r)
	if err != nil {
		return nil, err
	}

	statusFilter := normalizeAgentStatusFilter(r.URL.Query().Get("status"))
	if statusFilter == "invalid" {
		return nil, fmt.Errorf("%w: invalid status filter", errBadRequest)
	}

	agentsResp, err := p.agentClient.Agents(ctx, &agenttypes.QueryAgentsRequest{})
	if err != nil {
		return nil, err
	}

	latestHeight, agentParams, err := p.fetchChainHeightAndAgentParams(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]publicAPIAgentSummary, 0, len(agentsResp.Agents))
	for _, agent := range agentsResp.Agents {
		if statusFilter != "" && normalizeAgentStatus(agent.Status) != statusFilter {
			continue
		}
		items = append(items, summarizeAgent(agent, latestHeight, agentParams))
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].ReputationScore == items[j].ReputationScore {
			return items[i].AgentAddress < items[j].AgentAddress
		}
		return items[i].ReputationScore > items[j].ReputationScore
	})

	total := len(items)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}

	return map[string]any{
		"agents": items[offset:end],
		"pagination": publicAPIPagination{
			Offset: offset,
			Limit:  limit,
			Total:  total,
		},
		"filters": map[string]any{
			"status": statusFilter,
		},
	}, nil
}

func (p *publicAPI) handleAgent(ctx context.Context, r *http.Request) (any, error) {
	agentResp, err := p.agentClient.Agent(ctx, &agenttypes.QueryAgentRequest{Address: mux.Vars(r)["address"]})
	if err != nil {
		return nil, fmt.Errorf("%w: agent not found", errNotFound)
	}

	latestHeight, agentParams, err := p.fetchChainHeightAndAgentParams(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"agent": summarizeAgent(*agentResp.Agent, latestHeight, agentParams),
	}, nil
}

func (p *publicAPI) handleAgentHeartbeat(ctx context.Context, r *http.Request) (any, error) {
	address := mux.Vars(r)["address"]
	agentResp, err := p.agentClient.Agent(ctx, &agenttypes.QueryAgentRequest{Address: address})
	if err != nil {
		return nil, fmt.Errorf("%w: agent not found", errNotFound)
	}

	latestHeight, agentParams, err := p.fetchChainHeightAndAgentParams(ctx)
	if err != nil {
		return nil, err
	}

	lastHeartbeatTime := (*time.Time)(nil)
	if agentResp.Agent.LastHeartbeat > 0 {
		height := agentResp.Agent.LastHeartbeat
		blockResp, blockErr := p.clientCtx.Client.Block(ctx, &height)
		if blockErr == nil && blockResp != nil && blockResp.Block != nil {
			t := blockResp.Block.Time.UTC()
			lastHeartbeatTime = &t
		}
	}

	remaining := agentParams.HeartbeatTimeout - (latestHeight - agentResp.Agent.LastHeartbeat)
	if remaining < 0 {
		remaining = 0
	}

	return map[string]any{
		"agent_address":                  address,
		"validator_address":              valoperFromAcc(address),
		"status":                         normalizeAgentStatus(agentResp.Agent.Status),
		"last_heartbeat_height":          agentResp.Agent.LastHeartbeat,
		"last_heartbeat_time":            lastHeartbeatTime,
		"latest_block_height":            latestHeight,
		"remaining_blocks_until_offline": remaining,
		"heartbeat_timeout_blocks":       agentParams.HeartbeatTimeout,
	}, nil
}

func (p *publicAPI) handleOnlineValidators(ctx context.Context, _ *http.Request) (any, error) {
	agentsResp, err := p.agentClient.Agents(ctx, &agenttypes.QueryAgentsRequest{})
	if err != nil {
		return nil, err
	}

	latestHeight, agentParams, err := p.fetchChainHeightAndAgentParams(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]publicAPIAgentSummary, 0, len(agentsResp.Agents))
	for _, agent := range agentsResp.Agents {
		if normalizeAgentStatus(agent.Status) != "online" {
			continue
		}
		items = append(items, summarizeAgent(agent, latestHeight, agentParams))
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].ReputationScore == items[j].ReputationScore {
			return items[i].ValidatorAddress < items[j].ValidatorAddress
		}
		return items[i].ReputationScore > items[j].ReputationScore
	})

	return map[string]any{
		"validators": items,
		"pagination": publicAPIPagination{
			Offset: 0,
			Limit:  len(items),
			Total:  len(items),
		},
	}, nil
}

func (p *publicAPI) handleCurrentChallenge(ctx context.Context, _ *http.Request) (any, error) {
	resp, err := p.agentClient.CurrentChallenge(ctx, &agenttypes.QueryCurrentChallengeRequest{})
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"challenge": resp.Challenge,
	}, nil
}

func (p *publicAPI) handleSearch(ctx context.Context, r *http.Request) (any, error) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		return nil, fmt.Errorf("%w: missing q", errBadRequest)
	}

	if hash, ok := parseHexHash(query); ok {
		txResult, err := p.clientCtx.Client.Tx(ctx, hash, false)
		if err == nil {
			return map[string]any{
				"query": query,
				"type":  "tx",
				"match": summarizeResultTx(txResult),
			}, nil
		}
	}

	if strings.HasPrefix(query, axonconfig.Bech32PrefixValAddr) {
		validator, err := p.fetchValidator(ctx, query)
		if err == nil {
			return map[string]any{
				"query": query,
				"type":  "validator",
				"match": summarizeValidator(validator),
			}, nil
		}
	}

	if strings.HasPrefix(query, axonconfig.Bech32PrefixAccAddr) {
		accountInfo, accountErr := p.authClient.AccountInfo(ctx, &authtypes.QueryAccountInfoRequest{Address: query})
		agentResp, agentErr := p.agentClient.Agent(ctx, &agenttypes.QueryAgentRequest{Address: query})
		latestHeight, agentParams, _ := p.fetchChainHeightAndAgentParams(ctx)

		var accountData any
		if accountErr == nil && accountInfo != nil {
			accountData = summarizeBaseAccount(accountInfo.GetInfo())
		}

		result := map[string]any{
			"query": query,
			"type":  "account",
			"match": map[string]any{
				"account": accountData,
			},
		}
		if agentErr == nil {
			result["type"] = "agent"
			result["match"] = summarizeAgent(*agentResp.Agent, latestHeight, agentParams)
		} else if accountErr != nil {
			return nil, fmt.Errorf("%w: no result for query", errNotFound)
		}
		return result, nil
	}

	validators, err := p.fetchAllValidators(ctx)
	if err == nil {
		for _, validator := range validators {
			if strings.EqualFold(validator.Description.Moniker, query) {
				return map[string]any{
					"query": query,
					"type":  "validator",
					"match": summarizeValidator(validator),
				}, nil
			}
		}
	}

	agentsResp, err := p.agentClient.Agents(ctx, &agenttypes.QueryAgentsRequest{})
	if err == nil {
		latestHeight, agentParams, _ := p.fetchChainHeightAndAgentParams(ctx)
		for _, agent := range agentsResp.Agents {
			if strings.EqualFold(agent.AgentId, query) {
				return map[string]any{
					"query": query,
					"type":  "agent",
					"match": summarizeAgent(agent, latestHeight, agentParams),
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("%w: no result for query", errNotFound)
}

func (p *publicAPI) fetchAllValidators(ctx context.Context) ([]stakingtypes.Validator, error) {
	validatorsResp, err := p.stakingClient.Validators(ctx, &stakingtypes.QueryValidatorsRequest{
		Pagination: &query.PageRequest{
			Limit:      publicAPIAllValidatorFetchMax,
			CountTotal: true,
		},
	})
	if err != nil {
		return nil, err
	}

	return validatorsResp.Validators, nil
}

func (p *publicAPI) fetchValidator(ctx context.Context, valoper string) (stakingtypes.Validator, error) {
	resp, err := p.stakingClient.Validator(ctx, &stakingtypes.QueryValidatorRequest{ValidatorAddr: valoper})
	if err != nil {
		return stakingtypes.Validator{}, fmt.Errorf("%w: validator not found", errNotFound)
	}
	return resp.Validator, nil
}

func (p *publicAPI) fetchSigningInfo(ctx context.Context, validator stakingtypes.Validator) (slashingtypes.ValidatorSigningInfo, slashingtypes.Params, error) {
	consAddr := ""
	if p.clientCtx.InterfaceRegistry != nil && validator.ConsensusPubkey != nil {
		var pubKey cryptotypes.PubKey
		if err := p.clientCtx.InterfaceRegistry.UnpackAny(validator.ConsensusPubkey, &pubKey); err != nil {
			return slashingtypes.ValidatorSigningInfo{}, slashingtypes.Params{}, err
		}
		consAddr = sdk.ConsAddress(pubKey.Address()).String()
	} else {
		consAddrBytes, err := validator.GetConsAddr()
		if err != nil {
			return slashingtypes.ValidatorSigningInfo{}, slashingtypes.Params{}, err
		}

		consAddr = sdk.ConsAddress(consAddrBytes).String()
	}

	signingInfoResp, err := p.slashingClient.SigningInfo(ctx, &slashingtypes.QuerySigningInfoRequest{
		ConsAddress: consAddr,
	})
	if err != nil {
		return slashingtypes.ValidatorSigningInfo{}, slashingtypes.Params{}, err
	}

	slashingParamsResp, err := p.slashingClient.Params(ctx, &slashingtypes.QueryParamsRequest{})
	if err != nil {
		return slashingtypes.ValidatorSigningInfo{}, slashingtypes.Params{}, err
	}

	return signingInfoResp.ValSigningInfo, slashingParamsResp.Params, nil
}

func (p *publicAPI) fetchAgentByValidator(ctx context.Context, valoper string) (publicAPIAgentSummary, error) {
	accAddress, err := accAddressFromValoper(valoper)
	if err != nil {
		return publicAPIAgentSummary{}, fmt.Errorf("%w: invalid validator address", errBadRequest)
	}

	resp, err := p.agentClient.Agent(ctx, &agenttypes.QueryAgentRequest{Address: accAddress})
	if err != nil {
		return publicAPIAgentSummary{}, fmt.Errorf("%w: agent not found for validator", errNotFound)
	}

	latestHeight, agentParams, err := p.fetchChainHeightAndAgentParams(ctx)
	if err != nil {
		return publicAPIAgentSummary{}, err
	}

	return summarizeAgent(*resp.Agent, latestHeight, agentParams), nil
}

func (p *publicAPI) fetchChainHeightAndAgentParams(ctx context.Context) (int64, agenttypes.Params, error) {
	status, err := p.clientCtx.Client.Status(ctx)
	if err != nil {
		return 0, agenttypes.Params{}, err
	}

	paramsResp, err := p.agentClient.Params(ctx, &agenttypes.QueryParamsRequest{})
	if err != nil {
		return 0, agenttypes.Params{}, err
	}

	return status.SyncInfo.LatestBlockHeight, paramsResp.Params, nil
}

func summarizeValidator(validator stakingtypes.Validator) publicAPIValidatorSummary {
	return publicAPIValidatorSummary{
		Moniker:         validator.Description.Moniker,
		OperatorAddress: validator.OperatorAddress,
		BondedStatus:    validator.Status.String(),
		Jailed:          validator.Jailed,
		Tokens:          validator.Tokens,
		VotingPower:     validator.GetConsensusPower(sdk.DefaultPowerReduction),
		CommissionRate:  validator.Commission.Rate.String(),
	}
}

func summarizeBaseAccount(account *authtypes.BaseAccount) map[string]any {
	if account == nil {
		return nil
	}

	return map[string]any{
		"address":        account.Address,
		"pub_key":        summarizeConsensusPubkey(account.PubKey),
		"account_number": account.AccountNumber,
		"sequence":       account.Sequence,
	}
}

func summarizeAgent(agent agenttypes.Agent, latestHeight int64, params agenttypes.Params) publicAPIAgentSummary {
	remaining := int64(0)
	if params.HeartbeatTimeout > 0 {
		remaining = params.HeartbeatTimeout - (latestHeight - agent.LastHeartbeat)
		if remaining < 0 {
			remaining = 0
		}
	}

	return publicAPIAgentSummary{
		AgentAddress:         agent.Address,
		ValidatorAddress:     valoperFromAcc(agent.Address),
		AgentID:              agent.AgentId,
		Status:               normalizeAgentStatus(agent.Status),
		Model:                agent.Model,
		Capabilities:         agent.Capabilities,
		ReputationScore:      agent.Reputation,
		AgentStake:           agent.StakeAmount,
		RegisteredAt:         agent.RegisteredAt,
		LastHeartbeatHeight:  agent.LastHeartbeat,
		RemainingOfflineBars: remaining,
	}
}

func summarizeBlockMeta(meta *cmttmtypes.BlockMeta) map[string]any {
	return map[string]any{
		"height":     meta.Header.Height,
		"hash":       meta.BlockID.Hash.String(),
		"time":       meta.Header.Time.UTC(),
		"num_txs":    meta.NumTxs,
		"proposer":   sdk.ConsAddress(meta.Header.ProposerAddress).String(),
		"block_size": meta.BlockSize,
	}
}

func summarizeResultTx(txResult *cmttypes.ResultTx) map[string]any {
	return map[string]any{
		"hash":       strings.ToUpper(hex.EncodeToString(txResult.Hash)),
		"height":     txResult.Height,
		"index":      txResult.Index,
		"code":       txResult.TxResult.Code,
		"codespace":  txResult.TxResult.Codespace,
		"gas_wanted": txResult.TxResult.GasWanted,
		"gas_used":   txResult.TxResult.GasUsed,
	}
}

func parseLimit(r *http.Request, defaultValue, maxValue int) (int, error) {
	raw := strings.TrimSpace(r.URL.Query().Get("limit"))
	if raw == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%w: invalid limit", errBadRequest)
	}
	if value > maxValue {
		value = maxValue
	}
	return value, nil
}

func parseOffset(r *http.Request) (int, error) {
	raw := strings.TrimSpace(r.URL.Query().Get("offset"))
	if raw == "" {
		return 0, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("%w: invalid offset", errBadRequest)
	}
	return value, nil
}

func inferNetworkName(chainID string) string {
	lower := strings.ToLower(chainID)
	switch {
	case strings.Contains(lower, "mainnet"):
		return "mainnet"
	case strings.Contains(lower, "testnet"):
		return "testnet"
	case strings.Contains(lower, "devnet"):
		return "devnet"
	default:
		if lower == "" {
			return "unknown"
		}
		return lower
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func preferredGasPrice(minGasPrice, baseFee string) string {
	if baseFee == "" {
		return minGasPrice
	}

	minGasDec, minErr := sdkmath.LegacyNewDecFromStr(minGasPrice)
	baseFeeDec, baseErr := sdkmath.LegacyNewDecFromStr(baseFee)
	if minErr != nil || baseErr != nil {
		return minGasPrice
	}
	if baseFeeDec.GT(minGasDec) {
		return baseFee
	}
	return minGasPrice
}

func accAddressFromValoper(valoper string) (string, error) {
	valAddr, err := sdk.ValAddressFromBech32(valoper)
	if err != nil {
		return "", err
	}
	return sdk.AccAddress(valAddr).String(), nil
}

func valoperFromAcc(address string) string {
	accAddr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return ""
	}
	return sdk.ValAddress(accAddr).String()
}

func normalizeAgentStatus(status agenttypes.AgentStatus) string {
	switch status {
	case agenttypes.AgentStatus_AGENT_STATUS_ONLINE:
		return "online"
	case agenttypes.AgentStatus_AGENT_STATUS_OFFLINE:
		return "offline"
	case agenttypes.AgentStatus_AGENT_STATUS_SUSPENDED:
		return "suspended"
	default:
		return "unspecified"
	}
}

func normalizeAgentStatusFilter(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "":
		return ""
	case "online":
		return "online"
	case "offline":
		return "offline"
	case "suspended":
		return "suspended"
	default:
		return "invalid"
	}
}

func parseHexHash(input string) ([]byte, bool) {
	value := strings.TrimPrefix(strings.TrimSpace(input), "0x")
	if len(value) != 64 {
		return nil, false
	}

	hash, err := hex.DecodeString(value)
	if err != nil {
		return nil, false
	}

	return hash, true
}
