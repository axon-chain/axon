package app

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	queryv1beta1 "cosmossdk.io/api/cosmos/base/query/v1beta1"
	tendermintv1beta1 "cosmossdk.io/api/cosmos/base/tendermint/v1beta1"
	"github.com/cosmos/gogoproto/proto"

	cmtrpctypes "github.com/cometbft/cometbft/rpc/core/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"

	agenttypes "github.com/axon-chain/axon/x/agent/types"
)

const maxRequestBodySize = 2 * 1024 * 1024

var actionTypePattern = regexp.MustCompile(`^[a-zA-Z0-9_./]+$`)

type publicAPITxRequest struct {
	TxBytes string `json:"tx_bytes"`
	Mode    string `json:"mode,omitempty"`
}

func (p *publicAPI) handleStakingParams(ctx context.Context, _ *http.Request) (any, error) {
	resp, err := p.stakingClient.Params(ctx, &stakingtypes.QueryParamsRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Params, nil
}

func (p *publicAPI) handleSlashingParams(ctx context.Context, _ *http.Request) (any, error) {
	resp, err := p.slashingClient.Params(ctx, &slashingtypes.QueryParamsRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Params, nil
}

func (p *publicAPI) handleDistributionParams(ctx context.Context, _ *http.Request) (any, error) {
	resp, err := p.distribution.Params(ctx, &distrtypes.QueryParamsRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Params, nil
}

func (p *publicAPI) handleAgentParams(ctx context.Context, _ *http.Request) (any, error) {
	resp, err := p.agentClient.Params(ctx, &agenttypes.QueryParamsRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Params, nil
}

func (p *publicAPI) handleGovParams(ctx context.Context, _ *http.Request) (any, error) {
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
		"deposit": depositParams.GetParams(),
		"voting":  votingParams.GetParams(),
		"tally":   tallyParams.GetParams(),
	}, nil
}

func (p *publicAPI) handleLatestBlock(ctx context.Context, _ *http.Request) (any, error) {
	block, err := p.clientCtx.Client.Block(ctx, nil)
	if err != nil {
		return nil, err
	}
	return summarizeBlockResult(block), nil
}

func (p *publicAPI) handleBlock(ctx context.Context, r *http.Request) (any, error) {
	identifier := mux.Vars(r)["identifier"]
	if height, err := strconv.ParseInt(identifier, 10, 64); err == nil {
		block, blockErr := p.clientCtx.Client.Block(ctx, &height)
		if blockErr != nil {
			return nil, blockErr
		}
		return summarizeBlockResult(block), nil
	}

	hash, ok := parseHexHash(identifier)
	if !ok {
		return nil, fmt.Errorf("%w: invalid block identifier", errBadRequest)
	}

	block, err := p.clientCtx.Client.BlockByHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	return summarizeBlockResult(block), nil
}

func (p *publicAPI) handleBlocksRange(ctx context.Context, r *http.Request) (any, error) {
	limit, err := parseLimit(r, publicAPIDefaultLimit, publicAPIMaxLimit)
	if err != nil {
		return nil, err
	}

	status, err := p.clientCtx.Client.Status(ctx)
	if err != nil {
		return nil, err
	}

	toHeight := status.SyncInfo.LatestBlockHeight
	if raw := strings.TrimSpace(r.URL.Query().Get("to")); raw != "" {
		value, parseErr := strconv.ParseInt(raw, 10, 64)
		if parseErr != nil || value <= 0 {
			return nil, fmt.Errorf("%w: invalid to", errBadRequest)
		}
		toHeight = value
	}

	fromHeight := toHeight - int64(limit) + 1
	if raw := strings.TrimSpace(r.URL.Query().Get("from")); raw != "" {
		value, parseErr := strconv.ParseInt(raw, 10, 64)
		if parseErr != nil || value <= 0 {
			return nil, fmt.Errorf("%w: invalid from", errBadRequest)
		}
		fromHeight = value
	}
	if fromHeight < 1 {
		fromHeight = 1
	}
	if fromHeight > toHeight {
		return nil, fmt.Errorf("%w: from cannot exceed to", errBadRequest)
	}

	blockchainInfo, err := p.clientCtx.Client.BlockchainInfo(ctx, fromHeight, toHeight)
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
		"range": map[string]any{
			"from": fromHeight,
			"to":   toHeight,
		},
		"pagination": publicAPIPagination{
			Offset: 0,
			Limit:  limit,
			Total:  len(items),
		},
	}, nil
}

func (p *publicAPI) handleBlockTxs(ctx context.Context, r *http.Request) (any, error) {
	height, err := parseMuxInt64(r, "height")
	if err != nil {
		return nil, err
	}
	pageReq, limit, _, err := buildPageRequestFromQuery(r, publicAPIDefaultLimit, publicAPIMaxLimit)
	if err != nil {
		return nil, err
	}

	resp, err := p.txClient.GetBlockWithTxs(ctx, &txtypes.GetBlockWithTxsRequest{
		Height:     height,
		Pagination: pageReq,
	})
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"height":     height,
		"block_id":   resp.BlockId,
		"block":      resp.Block,
		"txs":        resp.Txs,
		"pagination": summarizePageResponse(limit, resp.Pagination),
	}, nil
}

func (p *publicAPI) handleBlockValidators(ctx context.Context, r *http.Request) (any, error) {
	height, err := parseMuxInt64(r, "height")
	if err != nil {
		return nil, err
	}
	pageReq, limit, _, err := buildPageRequestFromQuery(r, publicAPIDefaultLimit, publicAPIMaxLimit)
	if err != nil {
		return nil, err
	}

	resp, err := p.tendermintClient.GetValidatorSetByHeight(ctx, &tendermintv1beta1.GetValidatorSetByHeightRequest{
		Height:     height,
		Pagination: toProtoPageRequest(pageReq),
	})
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"height":     resp.BlockHeight,
		"validators": resp.Validators,
		"pagination": summarizeProtoPageResponse(limit, resp.Pagination),
	}, nil
}

func (p *publicAPI) handleBlockProposer(ctx context.Context, r *http.Request) (any, error) {
	height, err := parseMuxInt64(r, "height")
	if err != nil {
		return nil, err
	}

	blockResp, err := p.clientCtx.Client.Block(ctx, &height)
	if err != nil {
		return nil, err
	}
	if blockResp == nil || blockResp.Block == nil {
		return nil, fmt.Errorf("%w: block not found", errNotFound)
	}

	proposerConsAddr := sdk.ConsAddress(blockResp.Block.ProposerAddress).String()
	validatorSet, valErr := p.tendermintClient.GetValidatorSetByHeight(ctx, &tendermintv1beta1.GetValidatorSetByHeightRequest{
		Height: height,
		Pagination: &queryv1beta1.PageRequest{
			Limit: publicAPIAllValidatorFetchMax,
		},
	})

	result := map[string]any{
		"height":                     height,
		"proposer_consensus_address": proposerConsAddr,
	}
	if valErr == nil {
		for _, validator := range validatorSet.Validators {
			if validator.Address == proposerConsAddr {
				result["proposer"] = validator
				break
			}
		}
	}
	return result, nil
}

func (p *publicAPI) handleTx(ctx context.Context, r *http.Request) (any, error) {
	resp, err := p.txClient.GetTx(ctx, &txtypes.GetTxRequest{Hash: normalizeHashString(mux.Vars(r)["hash"])})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"tx":          summarizeProtoTx(resp.Tx),
		"tx_response": summarizeSDKTxResponse(resp.TxResponse),
	}, nil
}

func (p *publicAPI) handleTxEvents(ctx context.Context, r *http.Request) (any, error) {
	resp, err := p.txClient.GetTx(ctx, &txtypes.GetTxRequest{Hash: normalizeHashString(mux.Vars(r)["hash"])})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"txhash": normalizeHashString(mux.Vars(r)["hash"]),
		"events": resp.GetTxResponse().Events,
	}, nil
}

func (p *publicAPI) handleTxRaw(ctx context.Context, r *http.Request) (any, error) {
	resp, err := p.txClient.GetTx(ctx, &txtypes.GetTxRequest{Hash: normalizeHashString(mux.Vars(r)["hash"])})
	if err != nil {
		return nil, err
	}
	if resp.Tx == nil {
		return nil, fmt.Errorf("%w: tx not found", errNotFound)
	}
	raw, err := proto.Marshal(resp.Tx)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"txhash":       normalizeHashString(mux.Vars(r)["hash"]),
		"encoding":     "base64",
		"raw_tx_bytes": base64.StdEncoding.EncodeToString(raw),
	}, nil
}

func (p *publicAPI) handleTxSearch(ctx context.Context, r *http.Request) (any, error) {
	queryText := strings.TrimSpace(r.URL.Query().Get("q"))
	if queryText == "" {
		return nil, fmt.Errorf("%w: missing q", errBadRequest)
	}
	return p.queryTxs(ctx, queryText, r)
}

func (p *publicAPI) handleTxs(ctx context.Context, r *http.Request) (any, error) {
	queryText, err := buildTxFilterQuery(r)
	if err != nil {
		return nil, err
	}
	return p.queryTxs(ctx, queryText, r)
}

func (p *publicAPI) handleTxSimulate(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	var req publicAPITxRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		p.writeError(w, fmt.Errorf("%w: invalid JSON body", errBadRequest))
		return
	}

	txBytes, err := decodeRequestTxBytes(req.TxBytes)
	if err != nil {
		p.writeError(w, err)
		return
	}

	resp, err := p.txClient.Simulate(r.Context(), &txtypes.SimulateRequest{TxBytes: txBytes})
	if err != nil {
		p.writeError(w, err)
		return
	}

	payload, marshalErr := json.Marshal(summarizeSimulateResponse(resp))
	if marshalErr != nil {
		p.writeError(w, marshalErr)
		return
	}
	p.writeSuccess(w, payload, "node", time.Now().UTC())
}

func (p *publicAPI) handleTxBroadcast(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	var req publicAPITxRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		p.writeError(w, fmt.Errorf("%w: invalid JSON body", errBadRequest))
		return
	}

	txBytes, err := decodeRequestTxBytes(req.TxBytes)
	if err != nil {
		p.writeError(w, err)
		return
	}

	mode, err := parseBroadcastMode(req.Mode)
	if err != nil {
		p.writeError(w, err)
		return
	}

	resp, err := p.txClient.BroadcastTx(r.Context(), &txtypes.BroadcastTxRequest{
		TxBytes: txBytes,
		Mode:    mode,
	})
	if err != nil {
		p.writeError(w, err)
		return
	}

	payload, marshalErr := json.Marshal(map[string]any{
		"tx_response": summarizeSDKTxResponse(resp.TxResponse),
	})
	if marshalErr != nil {
		p.writeError(w, marshalErr)
		return
	}
	p.writeSuccess(w, payload, "node", time.Now().UTC())
}

func (p *publicAPI) handleAccount(ctx context.Context, r *http.Request) (any, error) {
	address := mux.Vars(r)["address"]
	accountInfo, err := p.authClient.AccountInfo(ctx, &authtypes.QueryAccountInfoRequest{Address: address})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"address": address,
		"account": summarizeBaseAccount(accountInfo.GetInfo()),
	}, nil
}

func (p *publicAPI) handleAccountBalances(ctx context.Context, r *http.Request) (any, error) {
	address := mux.Vars(r)["address"]
	pageReq, limit, _, err := buildPageRequestFromQuery(r, publicAPIDefaultLimit, publicAPIMaxLimit)
	if err != nil {
		return nil, err
	}
	resp, err := p.bankClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
		Address:    address,
		Pagination: pageReq,
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"address":    address,
		"balances":   resp.Balances,
		"pagination": summarizePageResponse(limit, resp.Pagination),
	}, nil
}

func (p *publicAPI) handleAccountSpendable(ctx context.Context, r *http.Request) (any, error) {
	address := mux.Vars(r)["address"]
	pageReq, limit, _, err := buildPageRequestFromQuery(r, publicAPIDefaultLimit, publicAPIMaxLimit)
	if err != nil {
		return nil, err
	}
	resp, err := p.bankClient.SpendableBalances(ctx, &banktypes.QuerySpendableBalancesRequest{
		Address:    address,
		Pagination: pageReq,
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"address":            address,
		"spendable_balances": resp.Balances,
		"pagination":         summarizePageResponse(limit, resp.Pagination),
	}, nil
}

func (p *publicAPI) handleAccountSequence(ctx context.Context, r *http.Request) (any, error) {
	address := mux.Vars(r)["address"]
	accountInfo, err := p.authClient.AccountInfo(ctx, &authtypes.QueryAccountInfoRequest{Address: address})
	if err != nil {
		return nil, err
	}
	info := accountInfo.GetInfo()
	return map[string]any{
		"address":        address,
		"account_number": info.GetAccountNumber(),
		"sequence":       info.GetSequence(),
	}, nil
}

func (p *publicAPI) handleAccountTxs(ctx context.Context, r *http.Request) (any, error) {
	address := mux.Vars(r)["address"]
	return p.queryTransactionsByAddress(ctx, address, false, r)
}

func (p *publicAPI) handleAccountTransfers(ctx context.Context, r *http.Request) (any, error) {
	address := mux.Vars(r)["address"]
	return p.queryTransactionsByAddress(ctx, address, true, r)
}

func (p *publicAPI) handleAccountRewards(ctx context.Context, r *http.Request) (any, error) {
	address := mux.Vars(r)["address"]
	resp, err := p.distribution.DelegationTotalRewards(ctx, &distrtypes.QueryDelegationTotalRewardsRequest{
		DelegatorAddress: address,
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"address":       address,
		"rewards":       resp.Rewards,
		"total_rewards": resp.Total,
	}, nil
}

func (p *publicAPI) handleValidators(ctx context.Context, r *http.Request) (any, error) {
	pageReq, limit, offset, err := buildPageRequestFromQuery(r, publicAPIDefaultLimit, publicAPIMaxLimit)
	if err != nil {
		return nil, err
	}
	resp, err := p.stakingClient.Validators(ctx, &stakingtypes.QueryValidatorsRequest{
		Status:     strings.TrimSpace(r.URL.Query().Get("status")),
		Pagination: pageReq,
	})
	if err != nil {
		return nil, err
	}
	items := make([]publicAPIValidatorSummary, 0, len(resp.Validators))
	for _, validator := range resp.Validators {
		items = append(items, summarizeValidator(validator))
	}
	return map[string]any{
		"validators": items,
		"pagination": publicAPIPagination{
			Offset: offset,
			Limit:  limit,
			Total:  int(resp.Pagination.GetTotal()),
		},
	}, nil
}

func (p *publicAPI) handleValidator(ctx context.Context, r *http.Request) (any, error) {
	validator, err := p.fetchValidator(ctx, mux.Vars(r)["valoper"])
	if err != nil {
		return nil, err
	}

	signingInfo, slashingParams, signingErr := p.fetchSigningInfo(ctx, validator)
	selfDelegation, _ := p.stakingClient.Delegation(ctx, &stakingtypes.QueryDelegationRequest{
		DelegatorAddr: valoperToAccNoErr(validator.OperatorAddress),
		ValidatorAddr: validator.OperatorAddress,
	})
	commission, _ := p.distribution.ValidatorCommission(ctx, &distrtypes.QueryValidatorCommissionRequest{ValidatorAddress: validator.OperatorAddress})
	agent, _ := p.fetchAgentByValidator(ctx, validator.OperatorAddress)

	result := map[string]any{
		"validator":        summarizeValidator(validator),
		"consensus_pubkey": summarizeConsensusPubkey(validator.ConsensusPubkey),
		"self_delegation": func() any {
			if selfDelegation == nil {
				return nil
			}
			return selfDelegation.DelegationResponse
		}(),
		"commission": func() any {
			if commission == nil {
				return nil
			}
			return commission.Commission
		}(),
	}
	if signingErr == nil {
		result["signing_info"] = signingInfo
		result["signed_blocks_window"] = slashingParams.SignedBlocksWindow
	}
	if agent.AgentAddress != "" {
		result["agent"] = agent
	}
	return result, nil
}

func (p *publicAPI) handleValidatorDelegations(ctx context.Context, r *http.Request) (any, error) {
	valoper := mux.Vars(r)["valoper"]
	pageReq, limit, _, err := buildPageRequestFromQuery(r, publicAPIDefaultLimit, publicAPIMaxLimit)
	if err != nil {
		return nil, err
	}
	resp, err := p.stakingClient.ValidatorDelegations(ctx, &stakingtypes.QueryValidatorDelegationsRequest{
		ValidatorAddr: valoper,
		Pagination:    pageReq,
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"validator_address": valoper,
		"delegations":       resp.DelegationResponses,
		"pagination":        summarizePageResponse(limit, resp.Pagination),
	}, nil
}

func (p *publicAPI) handleValidatorUnbondings(ctx context.Context, r *http.Request) (any, error) {
	valoper := mux.Vars(r)["valoper"]
	pageReq, limit, _, err := buildPageRequestFromQuery(r, publicAPIDefaultLimit, publicAPIMaxLimit)
	if err != nil {
		return nil, err
	}
	resp, err := p.stakingClient.ValidatorUnbondingDelegations(ctx, &stakingtypes.QueryValidatorUnbondingDelegationsRequest{
		ValidatorAddr: valoper,
		Pagination:    pageReq,
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"validator_address": valoper,
		"unbondings":        resp.UnbondingResponses,
		"pagination":        summarizePageResponse(limit, resp.Pagination),
	}, nil
}

func (p *publicAPI) handleValidatorRedelegations(ctx context.Context, r *http.Request) (any, error) {
	valoper := mux.Vars(r)["valoper"]
	pageReq, limit, _, err := buildPageRequestFromQuery(r, publicAPIDefaultLimit, publicAPIMaxLimit)
	if err != nil {
		return nil, err
	}
	srcResp, srcErr := p.stakingClient.Redelegations(ctx, &stakingtypes.QueryRedelegationsRequest{
		SrcValidatorAddr: valoper,
		Pagination:       pageReq,
	})
	dstResp, dstErr := p.stakingClient.Redelegations(ctx, &stakingtypes.QueryRedelegationsRequest{
		DstValidatorAddr: valoper,
		Pagination:       pageReq,
	})
	if srcErr != nil && dstErr != nil {
		return nil, srcErr
	}
	seen := make(map[string]bool)
	items := make([]any, 0)
	appendItems := func(resp *stakingtypes.QueryRedelegationsResponse) {
		if resp == nil {
			return
		}
		for _, item := range resp.RedelegationResponses {
			key := item.Redelegation.DelegatorAddress + "|" + item.Redelegation.ValidatorSrcAddress + "|" + item.Redelegation.ValidatorDstAddress
			if seen[key] {
				continue
			}
			seen[key] = true
			items = append(items, item)
		}
	}
	appendItems(srcResp)
	appendItems(dstResp)
	return map[string]any{
		"validator_address": valoper,
		"redelegations":     items,
		"pagination": publicAPIPagination{
			Offset: 0,
			Limit:  limit,
			Total:  len(items),
		},
	}, nil
}

func (p *publicAPI) handleValidatorCommission(ctx context.Context, r *http.Request) (any, error) {
	valoper := mux.Vars(r)["valoper"]
	resp, err := p.distribution.ValidatorCommission(ctx, &distrtypes.QueryValidatorCommissionRequest{ValidatorAddress: valoper})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"validator_address": valoper,
		"commission":        resp.Commission,
	}, nil
}

func (p *publicAPI) handleValidatorRewards(ctx context.Context, r *http.Request) (any, error) {
	valoper := mux.Vars(r)["valoper"]
	outstanding, err := p.distribution.ValidatorOutstandingRewards(ctx, &distrtypes.QueryValidatorOutstandingRewardsRequest{ValidatorAddress: valoper})
	if err != nil {
		return nil, err
	}
	info, _ := p.distribution.ValidatorDistributionInfo(ctx, &distrtypes.QueryValidatorDistributionInfoRequest{ValidatorAddress: valoper})
	return map[string]any{
		"validator_address":   valoper,
		"outstanding_rewards": outstanding.Rewards,
		"distribution_info":   info,
	}, nil
}

func (p *publicAPI) handleValidatorSlashes(ctx context.Context, r *http.Request) (any, error) {
	valoper := mux.Vars(r)["valoper"]
	pageReq, limit, _, err := buildPageRequestFromQuery(r, publicAPIDefaultLimit, publicAPIMaxLimit)
	if err != nil {
		return nil, err
	}
	startHeight, _ := parseOptionalUint64(r.URL.Query().Get("from_height"))
	endHeight, _ := parseOptionalUint64(r.URL.Query().Get("to_height"))
	resp, err := p.distribution.ValidatorSlashes(ctx, &distrtypes.QueryValidatorSlashesRequest{
		ValidatorAddress: valoper,
		StartingHeight:   startHeight,
		EndingHeight:     endHeight,
		Pagination:       pageReq,
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"validator_address": valoper,
		"slashes":           resp.Slashes,
		"pagination":        summarizePageResponse(limit, resp.Pagination),
	}, nil
}

func (p *publicAPI) handleValidatorSigningInfo(ctx context.Context, r *http.Request) (any, error) {
	validator, err := p.fetchValidator(ctx, mux.Vars(r)["valoper"])
	if err != nil {
		return nil, err
	}
	signingInfo, slashingParams, err := p.fetchSigningInfo(ctx, validator)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"validator_address":     validator.OperatorAddress,
		"signing_info":          signingInfo,
		"signed_blocks_window":  slashingParams.SignedBlocksWindow,
		"min_signed_per_window": slashingParams.MinSignedPerWindow,
	}, nil
}

func (p *publicAPI) handleDelegatorDelegations(ctx context.Context, r *http.Request) (any, error) {
	address := mux.Vars(r)["address"]
	pageReq, limit, _, err := buildPageRequestFromQuery(r, publicAPIDefaultLimit, publicAPIMaxLimit)
	if err != nil {
		return nil, err
	}
	resp, err := p.stakingClient.DelegatorDelegations(ctx, &stakingtypes.QueryDelegatorDelegationsRequest{
		DelegatorAddr: address,
		Pagination:    pageReq,
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"delegator_address": address,
		"delegations":       resp.DelegationResponses,
		"pagination":        summarizePageResponse(limit, resp.Pagination),
	}, nil
}

func (p *publicAPI) handleDelegatorUnbondings(ctx context.Context, r *http.Request) (any, error) {
	address := mux.Vars(r)["address"]
	pageReq, limit, _, err := buildPageRequestFromQuery(r, publicAPIDefaultLimit, publicAPIMaxLimit)
	if err != nil {
		return nil, err
	}
	resp, err := p.stakingClient.DelegatorUnbondingDelegations(ctx, &stakingtypes.QueryDelegatorUnbondingDelegationsRequest{
		DelegatorAddr: address,
		Pagination:    pageReq,
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"delegator_address": address,
		"unbondings":        resp.UnbondingResponses,
		"pagination":        summarizePageResponse(limit, resp.Pagination),
	}, nil
}

func (p *publicAPI) handleDelegatorRedelegations(ctx context.Context, r *http.Request) (any, error) {
	address := mux.Vars(r)["address"]
	pageReq, limit, _, err := buildPageRequestFromQuery(r, publicAPIDefaultLimit, publicAPIMaxLimit)
	if err != nil {
		return nil, err
	}
	resp, err := p.stakingClient.Redelegations(ctx, &stakingtypes.QueryRedelegationsRequest{
		DelegatorAddr: address,
		Pagination:    pageReq,
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"delegator_address": address,
		"redelegations":     resp.RedelegationResponses,
		"pagination":        summarizePageResponse(limit, resp.Pagination),
	}, nil
}

func (p *publicAPI) handleDelegatorRewards(ctx context.Context, r *http.Request) (any, error) {
	address := mux.Vars(r)["address"]
	resp, err := p.distribution.DelegationTotalRewards(ctx, &distrtypes.QueryDelegationTotalRewardsRequest{
		DelegatorAddress: address,
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"delegator_address": address,
		"rewards":           resp.Rewards,
		"total_rewards":     resp.Total,
	}, nil
}

func (p *publicAPI) handleDelegatorWithdrawAddress(ctx context.Context, r *http.Request) (any, error) {
	address := mux.Vars(r)["address"]
	resp, err := p.distribution.DelegatorWithdrawAddress(ctx, &distrtypes.QueryDelegatorWithdrawAddressRequest{
		DelegatorAddress: address,
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"delegator_address": address,
		"withdraw_address":  resp.WithdrawAddress,
	}, nil
}

func (p *publicAPI) handleDelegatorValidators(ctx context.Context, r *http.Request) (any, error) {
	address := mux.Vars(r)["address"]
	pageReq, limit, _, err := buildPageRequestFromQuery(r, publicAPIDefaultLimit, publicAPIMaxLimit)
	if err != nil {
		return nil, err
	}
	resp, err := p.stakingClient.DelegatorValidators(ctx, &stakingtypes.QueryDelegatorValidatorsRequest{
		DelegatorAddr: address,
		Pagination:    pageReq,
	})
	if err != nil {
		return nil, err
	}
	items := make([]publicAPIValidatorSummary, 0, len(resp.Validators))
	for _, validator := range resp.Validators {
		items = append(items, summarizeValidator(validator))
	}
	return map[string]any{
		"delegator_address": address,
		"validators":        items,
		"pagination": publicAPIPagination{
			Offset: 0,
			Limit:  limit,
			Total:  len(items),
		},
	}, nil
}

func (p *publicAPI) handleGovProposals(ctx context.Context, r *http.Request) (any, error) {
	pageReq, limit, _, err := buildPageRequestFromQuery(r, publicAPIDefaultLimit, publicAPIMaxLimit)
	if err != nil {
		return nil, err
	}
	status := govv1.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED
	if raw := strings.TrimSpace(r.URL.Query().Get("status")); raw != "" {
		parsed, parseErr := govv1.ProposalStatusFromString(raw)
		if parseErr != nil {
			return nil, fmt.Errorf("%w: invalid proposal status", errBadRequest)
		}
		status = parsed
	}
	resp, err := p.govClient.Proposals(ctx, &govv1.QueryProposalsRequest{
		ProposalStatus: status,
		Voter:          strings.TrimSpace(r.URL.Query().Get("voter")),
		Depositor:      strings.TrimSpace(r.URL.Query().Get("depositor")),
		Pagination:     pageReq,
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"proposals":  resp.Proposals,
		"pagination": summarizePageResponse(limit, resp.Pagination),
	}, nil
}

func (p *publicAPI) handleGovProposal(ctx context.Context, r *http.Request) (any, error) {
	id, err := parseMuxUint64(r, "id")
	if err != nil {
		return nil, err
	}
	resp, err := p.fetchGovProposal(ctx, id)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"proposal": resp.Proposal,
	}, nil
}

func (p *publicAPI) handleGovProposalVotes(ctx context.Context, r *http.Request) (any, error) {
	id, err := parseMuxUint64(r, "id")
	if err != nil {
		return nil, err
	}
	if _, err := p.fetchGovProposal(ctx, id); err != nil {
		return nil, err
	}
	pageReq, limit, _, err := buildPageRequestFromQuery(r, publicAPIDefaultLimit, publicAPIMaxLimit)
	if err != nil {
		return nil, err
	}
	resp, err := p.govClient.Votes(ctx, &govv1.QueryVotesRequest{ProposalId: id, Pagination: pageReq})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"proposal_id": id,
		"votes":       resp.Votes,
		"pagination":  summarizePageResponse(limit, resp.Pagination),
	}, nil
}

func (p *publicAPI) handleGovProposalTally(ctx context.Context, r *http.Request) (any, error) {
	id, err := parseMuxUint64(r, "id")
	if err != nil {
		return nil, err
	}
	if _, err := p.fetchGovProposal(ctx, id); err != nil {
		return nil, err
	}
	resp, err := p.govClient.TallyResult(ctx, &govv1.QueryTallyResultRequest{ProposalId: id})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"proposal_id": id,
		"tally":       resp.Tally,
	}, nil
}

func (p *publicAPI) handleAgentReputation(ctx context.Context, r *http.Request) (any, error) {
	address := mux.Vars(r)["address"]
	resp, err := p.agentClient.Reputation(ctx, &agenttypes.QueryReputationRequest{Address: address})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"agent_address": address,
		"reputation":    resp.Reputation,
	}, nil
}

func (p *publicAPI) fetchGovProposal(ctx context.Context, id uint64) (*govv1.QueryProposalResponse, error) {
	resp, err := p.govClient.Proposal(ctx, &govv1.QueryProposalRequest{ProposalId: id})
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Proposal == nil {
		return nil, fmt.Errorf("%w: proposal not found", errNotFound)
	}
	return resp, nil
}

func (p *publicAPI) handleAgentStake(ctx context.Context, r *http.Request) (any, error) {
	address := mux.Vars(r)["address"]
	resp, err := p.agentClient.Agent(ctx, &agenttypes.QueryAgentRequest{Address: address})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"agent_address": address,
		"stake":         resp.Agent.StakeAmount,
	}, nil
}

func (p *publicAPI) queryTxs(ctx context.Context, queryText string, r *http.Request) (any, error) {
	limit, err := parseLimit(r, publicAPIDefaultLimit, publicAPIMaxLimit)
	if err != nil {
		return nil, err
	}
	offset, err := parseOffset(r)
	if err != nil {
		return nil, err
	}
	page := offset/limit + 1
	orderBy := "desc"
	resp, err := p.clientCtx.Client.TxSearch(ctx, queryText, false, &page, &limit, orderBy)
	if err != nil {
		return nil, err
	}

	statusFilter := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("status")))
	items := make([]map[string]any, 0, len(resp.Txs))
	if statusFilter != "" {
		for _, txResp := range resp.Txs {
			switch statusFilter {
			case "success":
				if txResp.TxResult.Code == 0 {
					items = append(items, summarizeResultTx(txResp))
				}
			case "failed":
				if txResp.TxResult.Code != 0 {
					items = append(items, summarizeResultTx(txResp))
				}
			default:
				return nil, fmt.Errorf("%w: invalid status filter", errBadRequest)
			}
		}
	} else {
		for _, txResp := range resp.Txs {
			items = append(items, summarizeResultTx(txResp))
		}
	}

	return map[string]any{
		"query": queryText,
		"txs":   items,
		"pagination": publicAPIPagination{
			Offset: offset,
			Limit:  limit,
			Total:  len(items),
		},
	}, nil
}

func (p *publicAPI) queryTransactionsByAddress(ctx context.Context, address string, transfersOnly bool, r *http.Request) (any, error) {
	limit, err := parseLimit(r, publicAPIDefaultLimit, publicAPIMaxLimit)
	if err != nil {
		return nil, err
	}
	queries := []string{
		fmt.Sprintf("transfer.sender='%s'", address),
		fmt.Sprintf("transfer.recipient='%s'", address),
	}
	if !transfersOnly {
		queries = append(queries, fmt.Sprintf("message.sender='%s'", address))
	}
	results, err := p.searchAndMergeTxs(ctx, queries, limit)
	if err != nil {
		return nil, err
	}
	items := make([]map[string]any, 0, len(results))
	for _, txResult := range results {
		items = append(items, summarizeResultTx(txResult))
	}
	return map[string]any{
		"address":    address,
		"txs":        items,
		"pagination": publicAPIPagination{Offset: 0, Limit: limit, Total: len(items)},
	}, nil
}

func (p *publicAPI) searchAndMergeTxs(ctx context.Context, queries []string, limit int) ([]*cmtrpctypes.ResultTx, error) {
	type wrapped struct {
		hash string
		tx   *cmtrpctypes.ResultTx
	}

	seen := make(map[string]*cmtrpctypes.ResultTx)
	page := 1
	orderBy := "desc"
	for _, q := range queries {
		resp, err := p.clientCtx.Client.TxSearch(ctx, q, false, &page, &limit, orderBy)
		if err != nil {
			continue
		}
		for _, txResult := range resp.Txs {
			key := strings.ToUpper(hex.EncodeToString(txResult.Hash))
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = txResult
		}
	}

	items := make([]wrapped, 0, len(seen))
	for hash, txResult := range seen {
		items = append(items, wrapped{hash: hash, tx: txResult})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].tx.Height == items[j].tx.Height {
			return items[i].tx.Index > items[j].tx.Index
		}
		return items[i].tx.Height > items[j].tx.Height
	})
	if len(items) > limit {
		items = items[:limit]
	}
	result := make([]*cmtrpctypes.ResultTx, 0, len(items))
	for _, item := range items {
		result = append(result, item.tx)
	}
	return result, nil
}

func summarizeBlockResult(blockResp *cmtrpctypes.ResultBlock) map[string]any {
	if blockResp == nil || blockResp.Block == nil {
		return map[string]any{}
	}
	return map[string]any{
		"height":               blockResp.Block.Height,
		"hash":                 blockResp.BlockID.Hash.String(),
		"time":                 blockResp.Block.Time.UTC(),
		"chain_id":             blockResp.Block.ChainID,
		"num_txs":              len(blockResp.Block.Txs),
		"proposer":             sdk.ConsAddress(blockResp.Block.ProposerAddress).String(),
		"last_commit_hash":     strings.ToUpper(hex.EncodeToString(blockResp.Block.LastCommitHash)),
		"data_hash":            strings.ToUpper(hex.EncodeToString(blockResp.Block.DataHash)),
		"validators_hash":      strings.ToUpper(hex.EncodeToString(blockResp.Block.ValidatorsHash)),
		"next_validators_hash": strings.ToUpper(hex.EncodeToString(blockResp.Block.NextValidatorsHash)),
		"app_hash":             strings.ToUpper(hex.EncodeToString(blockResp.Block.AppHash)),
		"evidence_hash":        strings.ToUpper(hex.EncodeToString(blockResp.Block.EvidenceHash)),
	}
}

func buildPageRequestFromQuery(r *http.Request, defaultLimit, maxLimit int) (*query.PageRequest, int, int, error) {
	limit, err := parseLimit(r, defaultLimit, maxLimit)
	if err != nil {
		return nil, 0, 0, err
	}
	offset, err := parseOffset(r)
	if err != nil {
		return nil, 0, 0, err
	}
	return &query.PageRequest{
		Offset:     uint64(offset),
		Limit:      uint64(limit),
		CountTotal: true,
	}, limit, offset, nil
}

func summarizePageResponse(limit int, pageResp *query.PageResponse) publicAPIPagination {
	total := 0
	if pageResp != nil {
		total = int(pageResp.Total)
	}
	return publicAPIPagination{
		Offset: 0,
		Limit:  limit,
		Total:  total,
	}
}

func parseMuxInt64(r *http.Request, key string) (int64, error) {
	value, err := strconv.ParseInt(mux.Vars(r)[key], 10, 64)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("%w: invalid %s", errBadRequest, key)
	}
	return value, nil
}

func parseMuxUint64(r *http.Request, key string) (uint64, error) {
	value, err := strconv.ParseUint(mux.Vars(r)[key], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: invalid %s", errBadRequest, key)
	}
	return value, nil
}

func parseOptionalUint64(value string) (uint64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: invalid uint64 value", errBadRequest)
	}
	return parsed, nil
}

func decodeRequestTxBytes(raw string) ([]byte, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, fmt.Errorf("%w: missing tx_bytes", errBadRequest)
	}
	if decoded, err := base64.StdEncoding.DecodeString(value); err == nil {
		return decoded, nil
	}
	value = strings.TrimPrefix(value, "0x")
	decoded, err := hex.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("%w: tx_bytes must be base64 or hex", errBadRequest)
	}
	return decoded, nil
}

func parseBroadcastMode(mode string) (txtypes.BroadcastMode, error) {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", "sync":
		return txtypes.BroadcastMode_BROADCAST_MODE_SYNC, nil
	case "async":
		return txtypes.BroadcastMode_BROADCAST_MODE_ASYNC, nil
	case "block":
		return txtypes.BroadcastMode_BROADCAST_MODE_BLOCK, nil
	default:
		return txtypes.BroadcastMode_BROADCAST_MODE_UNSPECIFIED, fmt.Errorf("%w: invalid broadcast mode", errBadRequest)
	}
}

func normalizeHashString(hash string) string {
	return strings.ToUpper(strings.TrimPrefix(strings.TrimSpace(hash), "0x"))
}

func buildTxFilterQuery(r *http.Request) (string, error) {
	var clauses []string

	if sender := strings.TrimSpace(r.URL.Query().Get("sender")); sender != "" {
		if !isValidAddress(sender) {
			return "", fmt.Errorf("%w: invalid sender", errBadRequest)
		}
		clauses = append(clauses, fmt.Sprintf("message.sender='%s'", sanitizeQueryStringValue(sender)))
	}
	if recipient := strings.TrimSpace(r.URL.Query().Get("recipient")); recipient != "" {
		if !isValidAddress(recipient) {
			return "", fmt.Errorf("%w: invalid recipient", errBadRequest)
		}
		clauses = append(clauses, fmt.Sprintf("transfer.recipient='%s'", sanitizeQueryStringValue(recipient)))
	}
	if txType := strings.TrimSpace(r.URL.Query().Get("type")); txType != "" {
		if !isValidActionType(txType) {
			return "", fmt.Errorf("%w: invalid type", errBadRequest)
		}
		clauses = append(clauses, fmt.Sprintf("message.action='%s'", sanitizeQueryStringValue(txType)))
	}
	if fromHeight := strings.TrimSpace(r.URL.Query().Get("from_height")); fromHeight != "" {
		if _, err := strconv.ParseInt(fromHeight, 10, 64); err != nil {
			return "", fmt.Errorf("%w: invalid from_height", errBadRequest)
		}
		clauses = append(clauses, fmt.Sprintf("tx.height >= %s", fromHeight))
	}
	if toHeight := strings.TrimSpace(r.URL.Query().Get("to_height")); toHeight != "" {
		if _, err := strconv.ParseInt(toHeight, 10, 64); err != nil {
			return "", fmt.Errorf("%w: invalid to_height", errBadRequest)
		}
		clauses = append(clauses, fmt.Sprintf("tx.height <= %s", toHeight))
	}
	if len(clauses) == 0 {
		return "tx.height > 0", nil
	}
	return strings.Join(clauses, " AND "), nil
}

func isValidAddress(value string) bool {
	if _, err := sdk.AccAddressFromBech32(value); err == nil {
		return true
	}
	return common.IsHexAddress(value)
}

func isValidActionType(value string) bool {
	return actionTypePattern.MatchString(value)
}

func sanitizeQueryStringValue(value string) string {
	return strings.ReplaceAll(value, "'", "")
}

func valoperToAccNoErr(valoper string) string {
	address, err := accAddressFromValoper(valoper)
	if err != nil {
		return ""
	}
	return address
}

func summarizeConsensusPubkey(pubkey *codectypes.Any) map[string]any {
	if pubkey == nil {
		return nil
	}

	return map[string]any{
		"type_url":     pubkey.TypeUrl,
		"value_base64": base64.StdEncoding.EncodeToString(pubkey.Value),
	}
}

func summarizeProtoTx(tx *txtypes.Tx) map[string]any {
	if tx == nil {
		return nil
	}

	body := map[string]any{
		"message_types":                        []string{},
		"memo":                                 "",
		"timeout_height":                       uint64(0),
		"extension_options_count":              0,
		"non_critical_extension_options_count": 0,
	}
	if tx.Body != nil {
		messageTypes := make([]string, 0, len(tx.Body.Messages))
		for _, msg := range tx.Body.Messages {
			messageTypes = append(messageTypes, msg.TypeUrl)
		}
		body = map[string]any{
			"message_types":                        messageTypes,
			"memo":                                 tx.Body.Memo,
			"timeout_height":                       tx.Body.TimeoutHeight,
			"extension_options_count":              len(tx.Body.ExtensionOptions),
			"non_critical_extension_options_count": len(tx.Body.NonCriticalExtensionOptions),
		}
	}

	authInfo := map[string]any{
		"signer_count": 0,
		"fee":          nil,
		"tip":          nil,
	}
	if tx.AuthInfo != nil {
		authInfo = map[string]any{
			"signer_count": len(tx.AuthInfo.SignerInfos),
			"fee":          tx.AuthInfo.Fee,
			"tip":          tx.AuthInfo.Tip,
		}
	}

	return map[string]any{
		"body":             body,
		"auth_info":        authInfo,
		"signatures_count": len(tx.Signatures),
	}
}

func summarizeSDKTxResponse(resp *sdk.TxResponse) map[string]any {
	if resp == nil {
		return nil
	}

	return map[string]any{
		"height":     resp.Height,
		"txhash":     resp.TxHash,
		"codespace":  resp.Codespace,
		"code":       resp.Code,
		"data":       resp.Data,
		"raw_log":    resp.RawLog,
		"logs":       resp.Logs,
		"info":       resp.Info,
		"gas_wanted": resp.GasWanted,
		"gas_used":   resp.GasUsed,
		"timestamp":  resp.Timestamp,
		"events":     resp.Events,
	}
}

func summarizeSimulateResponse(resp *txtypes.SimulateResponse) map[string]any {
	if resp == nil {
		return nil
	}

	result := map[string]any{
		"gas_info": nil,
		"result":   nil,
	}
	if resp.GasInfo != nil {
		result["gas_info"] = resp.GasInfo
	}
	if resp.Result != nil {
		msgResponseTypes := make([]string, 0, len(resp.Result.MsgResponses))
		for _, item := range resp.Result.MsgResponses {
			msgResponseTypes = append(msgResponseTypes, item.TypeUrl)
		}
		result["result"] = map[string]any{
			"data_base64":        base64.StdEncoding.EncodeToString(resp.Result.Data),
			"log":                resp.Result.Log,
			"events":             resp.Result.Events,
			"msg_response_types": msgResponseTypes,
		}
	}
	return result
}

func toProtoPageRequest(pageReq *query.PageRequest) *queryv1beta1.PageRequest {
	if pageReq == nil {
		return nil
	}
	return &queryv1beta1.PageRequest{
		Key:        pageReq.Key,
		Offset:     pageReq.Offset,
		Limit:      pageReq.Limit,
		CountTotal: pageReq.CountTotal,
		Reverse:    pageReq.Reverse,
	}
}

func summarizeProtoPageResponse(limit int, pageResp *queryv1beta1.PageResponse) publicAPIPagination {
	total := 0
	if pageResp != nil {
		total = int(pageResp.Total)
	}
	return publicAPIPagination{
		Offset: 0,
		Limit:  limit,
		Total:  total,
	}
}
