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
