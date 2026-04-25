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

func (k queryServer) Services(goCtx context.Context, req *types.QueryServicesRequest) (*types.QueryServicesResponse, error) {
	if goCtx == nil {
		goCtx = context.Background()
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	var services []types.AgentService
	if req.Capability != "" {
		limit := int(req.Limit)
		if limit <= 0 {
			limit = 50
		}
		services = k.GetServicesByCapabilitySorted(ctx, req.Capability, limit)
	} else if req.AgentAddress != "" {
		services = k.GetServicesByAgent(ctx, req.AgentAddress)
	} else {
		services = k.GetAllServices(ctx)
	}

	return &types.QueryServicesResponse{Services: services}, nil
}

func (k queryServer) Service(goCtx context.Context, req *types.QueryServiceRequest) (*types.QueryServiceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if req.ServiceId == "" {
		return nil, status.Error(codes.InvalidArgument, "service_id is required")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	service, found := k.GetServiceById(ctx, req.ServiceId)
	if !found {
		return nil, status.Errorf(codes.NotFound, "service %s not found", req.ServiceId)
	}

	return &types.QueryServiceResponse{Service: service}, nil
}

func (k queryServer) ServiceCalls(goCtx context.Context, req *types.QueryServiceCallsRequest) (*types.QueryServiceCallsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if req.ServiceId == "" {
		return nil, status.Error(codes.InvalidArgument, "service_id is required")
	}

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 20
	}

	return &types.QueryServiceCallsResponse{Calls: []types.ServiceCall{}}, nil
}

func (k queryServer) Tasks(goCtx context.Context, req *types.QueryTasksRequest) (*types.QueryTasksResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	var tasks []types.TaskRequest
	if req.Status != 0 {
		tasks = k.GetTasksByStatus(ctx, req.Status)
	} else if req.Requester != "" {
		tasks = k.GetTasksByRequester(ctx, req.Requester)
	} else {
		tasks = k.GetAllTasks(ctx)
	}

	limit := int(req.Limit)
	if limit > 0 && len(tasks) > limit {
		tasks = tasks[:limit]
	}

	return &types.QueryTasksResponse{Tasks: tasks}, nil
}

func (k queryServer) Task(goCtx context.Context, req *types.QueryTaskRequest) (*types.QueryTaskResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if req.TaskId == "" {
		return nil, status.Error(codes.InvalidArgument, "task_id is required")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	task, found := k.GetTaskById(ctx, req.TaskId)
	if !found {
		return nil, status.Errorf(codes.NotFound, "task %s not found", req.TaskId)
	}

	return &types.QueryTaskResponse{Task: task}, nil
}

func (k queryServer) TaskBids(goCtx context.Context, req *types.QueryTaskBidsRequest) (*types.QueryTaskBidsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if req.TaskId == "" {
		return nil, status.Error(codes.InvalidArgument, "task_id is required")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	bids := k.GetBidsForTask(ctx, req.TaskId)

	limit := int(req.Limit)
	if limit > 0 && len(bids) > limit {
		bids = bids[:limit]
	}

	return &types.QueryTaskBidsResponse{Bids: bids}, nil
}

func (k queryServer) Tools(goCtx context.Context, req *types.QueryToolsRequest) (*types.QueryToolsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	var tools []types.ToolDefinition
	if req.PublicOnly {
		tools = k.GetPublicTools(ctx)
	} else if req.AgentAddress != "" {
		tools = k.GetToolsByAgent(ctx, req.AgentAddress)
	} else {
		tools = k.GetAllTools(ctx)
	}

	limit := int(req.Limit)
	if limit > 0 && len(tools) > limit {
		tools = tools[:limit]
	}

	return &types.QueryToolsResponse{Tools: tools}, nil
}

func (k queryServer) Tool(goCtx context.Context, req *types.QueryToolRequest) (*types.QueryToolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if req.ToolId == "" {
		return nil, status.Error(codes.InvalidArgument, "tool_id is required")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	tool, found := k.GetToolById(ctx, req.ToolId)
	if !found {
		return nil, status.Errorf(codes.NotFound, "tool %s not found", req.ToolId)
	}

	return &types.QueryToolResponse{Tool: tool}, nil
}
