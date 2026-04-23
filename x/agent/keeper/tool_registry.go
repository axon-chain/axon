package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	storetypes "cosmossdk.io/store/types"

	"github.com/axon-chain/axon/x/agent/types"
)

func (k Keeper) RegisterTool(
	ctx sdk.Context,
	agentAddr string,
	name string,
	description string,
	inputSchema string,
	outputSchema string,
	price sdk.Coin,
	isPublic bool,
) (string, error) {
	agent, found := k.GetAgent(ctx, agentAddr)
	if !found {
		return "", types.ErrAgentNotFound
	}
	if agent.Status != types.AgentStatus_AGENT_STATUS_ONLINE {
		return "", fmt.Errorf("agent must be online to register tools")
	}

	exists := k.GetTool(ctx, agentAddr, name)
	if exists.Id != "" {
		return "", fmt.Errorf("tool with name '%s' already exists", name)
	}

	toolId := k.generateToolId(agentAddr, name, ctx.BlockHeight())
	tool := types.ToolDefinition{
		Id:          toolId,
		AgentAddress: agentAddr,
		Name:        name,
		Description: description,
		InputSchema: inputSchema,
		OutputSchema: outputSchema,
		Price:       price,
		IsPublic:    isPublic,
		Status:     types.ToolStatus_TOOL_STATUS_ACTIVE,
		CreatedAt:   ctx.BlockHeight(),
	}

	if err := tool.Validate(); err != nil {
		return "", err
	}

	k.SetTool(ctx, tool)

	k.IndexTool(ctx, tool)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"tool_registered",
		sdk.NewAttribute("tool_id", toolId),
		sdk.NewAttribute("agent_address", agentAddr),
		sdk.NewAttribute("name", name),
		sdk.NewAttribute("price", price.String()),
	))

	return toolId, nil
}

func (k Keeper) CallTool(
	ctx sdk.Context,
	callerAddr string,
	toolId string,
	inputData []byte,
	payment sdk.Coin,
) ([]byte, error) {
	tool, found := k.GetToolById(ctx, toolId)
	if !found {
		return nil, types.ErrToolNotFound
	}
	if !tool.IsActive() {
		return nil, types.ErrToolNotActive
	}

	callerAgent, found := k.GetAgent(ctx, callerAddr)
	if !found {
		callerAgent = types.Agent{Address: callerAddr}
	}

	isOwner := callerAddr == tool.AgentAddress
	isPublicCall := tool.IsPublic && callerAgent.Status == types.AgentStatus_AGENT_STATUS_ONLINE

	if !isOwner && !isPublicCall {
		return nil, types.ErrToolNotPublic
	}

	if payment.IsLT(tool.Price) {
		return nil, fmt.Errorf("insufficient payment: required %s, got %s", tool.Price, payment)
	}

	ownerAgent, found := k.GetAgent(ctx, tool.AgentAddress)
	if !found || (ownerAgent.Status != types.AgentStatus_AGENT_STATUS_ONLINE && !isOwner) {
		return nil, types.ErrAgentOffline
	}

	if !isOwner {
		if err := k.PayAgent(ctx, tool.AgentAddress, payment); err != nil {
			return nil, fmt.Errorf("failed to pay tool owner: %v", err)
		}

		k.UpdateReputation(ctx, tool.AgentAddress, 1)
	}

	callId := k.generateToolCallId(toolId, callerAddr, ctx.BlockHeight())
	call := types.ToolCall{
		CallId:     callId,
		Caller:     callerAddr,
		ToolId:     toolId,
		InputData:  inputData,
		OutputData:  nil,
		Payment:   payment,
		ExecutedAt: ctx.BlockHeight(),
		Success:   false,
	}

	k.SetToolCall(ctx, call)

	outputData := []byte(fmt.Sprintf(`{"status":"executed","tool_id":"%s","message":"Tool executed - actual result depends on tool implementation"}`, toolId))

	call.OutputData = outputData
	call.Success = true
	k.SetToolCall(ctx, call)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"tool_called",
		sdk.NewAttribute("call_id", callId),
		sdk.NewAttribute("tool_id", toolId),
		sdk.NewAttribute("caller", callerAddr),
		sdk.NewAttribute("payment", payment.String()),
	))

	return outputData, nil
}

func (k Keeper) SetTool(ctx sdk.Context, tool types.ToolDefinition) {
	store := ctx.KVStore(k.storeKey)
	key := []byte(types.ToolPrefix + tool.Id)
	bz, _ := json.Marshal(&tool)
	store.Set(key, bz)
}

func (k Keeper) GetToolById(ctx sdk.Context, id string) (types.ToolDefinition, bool) {
	store := ctx.KVStore(k.storeKey)
	key := []byte(types.ToolPrefix + id)
	bz := store.Get(key)
	if bz == nil {
		return types.ToolDefinition{}, false
	}
	var tool types.ToolDefinition
	if err := json.Unmarshal(bz, &tool); err != nil {
		return types.ToolDefinition{}, false
	}
	return tool, true
}

func (k Keeper) GetTool(ctx sdk.Context, agentAddr, name string) types.ToolDefinition {
	store := ctx.KVStore(k.storeKey)
	prefix := []byte(types.ToolPrefix)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var tool types.ToolDefinition
		if err := json.Unmarshal(iterator.Value(), &tool); err == nil {
			if tool.AgentAddress == agentAddr && tool.Name == name {
				return tool
			}
		}
	}
	return types.ToolDefinition{}
}

func (k Keeper) IndexTool(ctx sdk.Context, tool types.ToolDefinition) {
	store := ctx.KVStore(k.storeKey)
	key := []byte("ToolAgent/" + tool.AgentAddress + "/" + tool.Id)
	store.Set(key, []byte(tool.Id))

	if tool.IsPublic {
		key := []byte("ToolPublic/" + tool.Id)
		store.Set(key, []byte(tool.Id))
	}
}

func (k Keeper) GetToolsByAgent(ctx sdk.Context, agentAddr string) []types.ToolDefinition {
	store := ctx.KVStore(k.storeKey)
	prefix := []byte(types.ToolPrefix)

	var tools []types.ToolDefinition
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var tool types.ToolDefinition
		if err := json.Unmarshal(iterator.Value(), &tool); err == nil {
			if tool.AgentAddress == agentAddr {
				tools = append(tools, tool)
			}
		}
	}
	return tools
}

func (k Keeper) GetPublicTools(ctx sdk.Context) []types.ToolDefinition {
	store := ctx.KVStore(k.storeKey)
	prefix := []byte("ToolPublic/")

	var tools []types.ToolDefinition
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		toolId := string(iterator.Value())
		if tool, found := k.GetToolById(ctx, toolId); found {
			if tool.IsActive() {
				tools = append(tools, tool)
			}
		}
	}
	return tools
}

func (k Keeper) GetAllTools(ctx sdk.Context) []types.ToolDefinition {
	store := ctx.KVStore(k.storeKey)
	prefix := []byte(types.ToolPrefix)

	var tools []types.ToolDefinition
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var tool types.ToolDefinition
		if err := json.Unmarshal(iterator.Value(), &tool); err == nil {
			tools = append(tools, tool)
		}
	}
	return tools
}

func (k Keeper) SetToolCall(ctx sdk.Context, call types.ToolCall) {
	store := ctx.KVStore(k.storeKey)
	key := []byte(types.ToolCallPrefix + call.CallId)
	bz, _ := json.Marshal(&call)
	store.Set(key, bz)
}

func (k Keeper) GetToolCall(ctx sdk.Context, callId string) (types.ToolCall, bool) {
	store := ctx.KVStore(k.storeKey)
	key := []byte(types.ToolCallPrefix + callId)
	bz := store.Get(key)
	if bz == nil {
		return types.ToolCall{}, false
	}
	var call types.ToolCall
	if err := json.Unmarshal(bz, &call); err != nil {
		return types.ToolCall{}, false
	}
	return call, true
}

func (k Keeper) generateToolCallId(toolId, caller string, blockTime int64) string {
	data := fmt.Sprintf("%s:%s:%d", toolId, caller, blockTime)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16])
}