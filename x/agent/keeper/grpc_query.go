package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/axon-chain/axon/x/agent/types"
)

type queryServer struct {
	types.UnimplementedQueryServer
	Keeper
}

var _ types.QueryServer = queryServer{}

func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &queryServer{Keeper: keeper}
}

func (k queryServer) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.QueryParamsResponse{Params: k.GetParams(ctx)}, nil
}

func (k queryServer) Agent(goCtx context.Context, req *types.QueryAgentRequest) (*types.QueryAgentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	agent, found := k.GetAgent(ctx, req.Address)
	if !found {
		return nil, status.Error(codes.NotFound, "agent not found")
	}

	return &types.QueryAgentResponse{Agent: &agent}, nil
}

const maxAgentsPerQuery = 200

func (k queryServer) Agents(goCtx context.Context, req *types.QueryAgentsRequest) (*types.QueryAgentsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	var agents []types.Agent
	k.IterateAgents(ctx, func(agent types.Agent) bool {
		agents = append(agents, agent)
		return len(agents) >= maxAgentsPerQuery
	})
	return &types.QueryAgentsResponse{Agents: agents}, nil
}

func (k queryServer) Reputation(goCtx context.Context, req *types.QueryReputationRequest) (*types.QueryReputationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	rep := k.GetReputation(ctx, req.Address)
	return &types.QueryReputationResponse{Reputation: rep}, nil
}

func (k queryServer) CurrentChallenge(goCtx context.Context, req *types.QueryCurrentChallengeRequest) (*types.QueryCurrentChallengeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	epoch := k.GetCurrentEpoch(ctx)
	challenge, found := k.GetChallenge(ctx, epoch)
	if !found {
		return nil, status.Errorf(codes.NotFound, "no active challenge for epoch %d", epoch)
	}

	return &types.QueryCurrentChallengeResponse{Challenge: &challenge}, nil
}

func (k queryServer) AgentStats(goCtx context.Context, req *types.QueryAgentStatsRequest) (*types.QueryAgentStatsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address is required")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	stats, found := k.GetAgentStats(ctx, req.Address)
	if !found {
		return nil, status.Error(codes.NotFound, "agent stats not found")
	}

	agent, found := k.GetAgent(ctx, req.Address)
	if found {
		stats.Reputation = agent.Reputation
		stats.Status = agent.Status.String()
	}

	return &types.QueryAgentStatsResponse{Stats: stats}, nil
}

func (k queryServer) ReputationHistory(goCtx context.Context, req *types.QueryReputationHistoryRequest) (*types.QueryReputationHistoryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address is required")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	limit := req.Limit
	if limit == 0 || limit > 100 {
		limit = 20
	}

	history := k.GetReputationHistory(ctx, req.Address, limit)
	return &types.QueryReputationHistoryResponse{History: history}, nil
}

func (k queryServer) TopAgents(goCtx context.Context, req *types.QueryTopAgentsRequest) (*types.QueryTopAgentsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	limit := req.Limit
	if limit == 0 || limit > 100 {
		limit = 50
	}

	var agents []types.AgentStatistics
	switch req.SortBy {
	case "reputation", "":
		agents = k.GetTopAgentsByReputation(ctx, limit)
	case "success_rate":
		agents = k.GetTopAgentsBySuccessRate(ctx, 5, limit)
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unknown sort_by: %s", req.SortBy)
	}

	return &types.QueryTopAgentsResponse{Agents: agents}, nil
}

func (k queryServer) AgentsByCapability(goCtx context.Context, req *types.QueryAgentsByCapabilityRequest) (*types.QueryAgentsByCapabilityResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if len(req.Capabilities) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one capability is required")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	agents := k.GetAgentsByCapabilities(ctx, req.Capabilities, req.MatchAll)
	return &types.QueryAgentsByCapabilityResponse{Agents: agents}, nil
}

func (k queryServer) ChallengeMetrics(goCtx context.Context, req *types.QueryChallengeMetricsRequest) (*types.QueryChallengeMetricsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	metrics, found := k.GetChallengeMetrics(ctx, req.Epoch)
	if !found {
		return nil, status.Errorf(codes.NotFound, "metrics not found for epoch %d", req.Epoch)
	}

	return &types.QueryChallengeMetricsResponse{Metrics: metrics}, nil
}

func (k queryServer) AgentChallengeHistory(goCtx context.Context, req *types.QueryAgentChallengeHistoryRequest) (*types.QueryAgentChallengeHistoryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address is required")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	limit := req.Limit
	if limit == 0 || limit > 100 {
		limit = 20
	}

	responses := k.GetAgentChallengeHistory(ctx, req.Address, limit)
	return &types.QueryAgentChallengeHistoryResponse{Responses: responses}, nil
}
