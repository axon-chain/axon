package keeper

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	storetypes "cosmossdk.io/store/types"

	"github.com/axon-chain/axon/x/agent/types"
)

func (k Keeper) CreateTask(
	ctx sdk.Context,
	requester string,
	title string,
	description string,
	requiredCapabilities []string,
	budget sdk.Coin,
	deadlineBlocks int64,
) (string, error) {
	if title == "" {
		return "", fmt.Errorf("title is required")
	}
	if len(description) > types.MaxTaskDescriptionLen {
		return "", fmt.Errorf("description too long")
	}
	if !budget.IsPositive() {
		return "", fmt.Errorf("budget must be positive")
	}
	if deadlineBlocks <= 0 {
		deadlineBlocks = types.DefaultTaskDeadlineBlocks
	}

	taskId := k.generateTaskId(requester, title, ctx.BlockHeight())
	deadlineBlock := ctx.BlockHeight() + deadlineBlocks

	task := types.TaskRequest{
		Id:                   taskId,
		Requester:            requester,
		Title:                title,
		Description:          description,
		RequiredCapabilities: requiredCapabilities,
		Budget:               budget,
		DeadlineBlock:       deadlineBlock,
		Status:               types.TaskStatus_TASK_STATUS_OPEN,
		SelectedAgent:        "",
		CreatedAt:            ctx.BlockHeight(),
	}

	k.SetTask(ctx, task)

	k.IndexTask(ctx, task)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"task_created",
		sdk.NewAttribute("task_id", taskId),
		sdk.NewAttribute("requester", requester),
		sdk.NewAttribute("title", title),
		sdk.NewAttribute("budget", budget.String()),
	))

	return taskId, nil
}

func (k Keeper) CancelTask(ctx sdk.Context, requester, taskId string) error {
	task, found := k.GetTaskById(ctx, taskId)
	if !found {
		return types.ErrTaskNotFound
	}
	if task.Requester != requester {
		return types.ErrUnauthorized
	}
	if !task.IsOpen() {
		return fmt.Errorf("task cannot be cancelled in current state: %s", task.Status)
	}

	task.Status = types.TaskStatus_TASK_STATUS_CANCELLED
	k.SetTask(ctx, task)

	return nil
}

func (k Keeper) SubmitBid(
	ctx sdk.Context,
	agentAddr string,
	taskId string,
	proposal string,
	price sdk.Coin,
) error {
	task, found := k.GetTaskById(ctx, taskId)
	if !found {
		return types.ErrTaskNotFound
	}
	if !task.IsOpen() {
		return types.ErrTaskNotOpen
	}
	if ctx.BlockHeight() > task.DeadlineBlock {
		return types.ErrBidTooLate
	}

	agent, found := k.GetAgent(ctx, agentAddr)
	if !found || agent.Status != types.AgentStatus_AGENT_STATUS_ONLINE {
		return types.ErrAgentOffline
	}

	hasCap := false
	for _, reqCap := range task.RequiredCapabilities {
		for _, agentCap := range agent.Capabilities {
			if reqCap == agentCap {
				hasCap = true
				break
			}
		}
	}
	if !hasCap {
		return fmt.Errorf("agent does not have required capabilities")
	}

	if price.IsGT(task.Budget) {
		return types.ErrInsufficientBudget
	}

	existingBid := k.GetBid(ctx, taskId, agentAddr)
	if existingBid.TaskId != "" {
		return types.ErrBidAlreadyExists
	}

	bid := types.TaskBid{
		TaskId:       taskId,
		AgentAddress: agentAddr,
		Proposal:    proposal,
		Price:       price,
		SubmittedAt: ctx.BlockHeight(),
		Accepted:    false,
	}

	k.SetBid(ctx, bid)

	if task.Status == types.TaskStatus_TASK_STATUS_OPEN {
		task.Status = types.TaskStatus_TASK_STATUS_BIDDING
		k.SetTask(ctx, task)
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"bid_submitted",
		sdk.NewAttribute("task_id", taskId),
		sdk.NewAttribute("agent_address", agentAddr),
		sdk.NewAttribute("price", price.String()),
	))

	return nil
}

func (k Keeper) SelectBid(
	ctx sdk.Context,
	requester string,
	taskId string,
	agentAddr string,
) error {
	task, found := k.GetTaskById(ctx, taskId)
	if !found {
		return types.ErrTaskNotFound
	}
	if task.Requester != requester {
		return types.ErrUnauthorized
	}
	if !task.IsOpen() {
		return types.ErrTaskNotOpen
	}

	bid := k.GetBid(ctx, taskId, agentAddr)
	if bid.TaskId == "" {
		return fmt.Errorf("no bid found for agent %s", agentAddr)
	}

	bid.Accepted = true
	k.SetBid(ctx, bid)

	task.Status = types.TaskStatus_TASK_STATUS_IN_PROGRESS
	task.SelectedAgent = agentAddr
	k.SetTask(ctx, task)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"bid_selected",
		sdk.NewAttribute("task_id", taskId),
		sdk.NewAttribute("agent_address", agentAddr),
	))

	return nil
}

func (k Keeper) CompleteTask(
	ctx sdk.Context,
	agentAddr string,
	taskId string,
	completionData string,
) error {
	task, found := k.GetTaskById(ctx, taskId)
	if !found {
		return types.ErrTaskNotFound
	}
	if task.SelectedAgent != agentAddr {
		return types.ErrUnauthorized
	}
	if task.Status != types.TaskStatus_TASK_STATUS_IN_PROGRESS {
		return types.ErrTaskNotOpen
	}

	bid := k.GetBid(ctx, taskId, agentAddr)
	if bid.Price.IsZero() {
		return fmt.Errorf("no accepted bid found")
	}

	task.Status = types.TaskStatus_TASK_STATUS_COMPLETED
	k.SetTask(ctx, task)

	if err := k.PayAgent(ctx, agentAddr, bid.Price); err != nil {
		return fmt.Errorf("failed to pay agent: %v", err)
	}

	k.UpdateReputation(ctx, agentAddr, 1)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"task_completed",
		sdk.NewAttribute("task_id", taskId),
		sdk.NewAttribute("agent_address", agentAddr),
		sdk.NewAttribute("payment", bid.Price.String()),
	))

	return nil
}

func (k Keeper) SetTask(ctx sdk.Context, task types.TaskRequest) {
	store := ctx.KVStore(k.storeKey)
	key := []byte(types.TaskPrefix + task.Id)
	bz, _ := json.Marshal(&task)
	store.Set(key, bz)
}

func (k Keeper) GetTaskById(ctx sdk.Context, id string) (types.TaskRequest, bool) {
	store := ctx.KVStore(k.storeKey)
	key := []byte(types.TaskPrefix + id)
	bz := store.Get(key)
	if bz == nil {
		return types.TaskRequest{}, false
	}
	var task types.TaskRequest
	if err := json.Unmarshal(bz, &task); err != nil {
		return types.TaskRequest{}, false
	}
	return task, true
}

func (k Keeper) IndexTask(ctx sdk.Context, task types.TaskRequest) {
	for _, cap := range task.RequiredCapabilities {
		store := ctx.KVStore(k.storeKey)
		key := []byte("TaskCap/" + cap + "/" + task.Id)
		store.Set(key, []byte(task.Id))
	}
}

func (k Keeper) GetTasksByCapability(ctx sdk.Context, capability string) []types.TaskRequest {
	store := ctx.KVStore(k.storeKey)
	prefix := []byte("TaskCap/" + capability + "/")
	var tasks []types.TaskRequest

	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		taskId := string(iterator.Value())
		if task, found := k.GetTaskById(ctx, taskId); found {
			if task.IsOpen() {
				tasks = append(tasks, task)
			}
		}
	}
	return tasks
}

func (k Keeper) GetAllTasks(ctx sdk.Context) []types.TaskRequest {
	store := ctx.KVStore(k.storeKey)
	prefix := []byte(types.TaskPrefix)

	var tasks []types.TaskRequest
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var task types.TaskRequest
		if err := json.Unmarshal(iterator.Value(), &task); err == nil {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

func (k Keeper) GetTasksByRequester(ctx sdk.Context, requester string) []types.TaskRequest {
	store := ctx.KVStore(k.storeKey)
	prefix := []byte(types.TaskPrefix)

	var tasks []types.TaskRequest
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var task types.TaskRequest
		if err := json.Unmarshal(iterator.Value(), &task); err == nil {
			if task.Requester == requester {
				tasks = append(tasks, task)
			}
		}
	}
	return tasks
}

func (k Keeper) SetBid(ctx sdk.Context, bid types.TaskBid) {
	store := ctx.KVStore(k.storeKey)
	key := []byte(types.TaskBidPrefix + bid.TaskId + "/" + bid.AgentAddress)
	bz, _ := json.Marshal(&bid)
	store.Set(key, bz)
}

func (k Keeper) GetBid(ctx sdk.Context, taskId, agentAddr string) types.TaskBid {
	store := ctx.KVStore(k.storeKey)
	key := []byte(types.TaskBidPrefix + taskId + "/" + agentAddr)
	bz := store.Get(key)
	if bz == nil {
		return types.TaskBid{}
	}
	var bid types.TaskBid
	if err := json.Unmarshal(bz, &bid); err != nil {
		return types.TaskBid{}
	}
	return bid
}

func (k Keeper) GetBidsForTask(ctx sdk.Context, taskId string) []types.TaskBid {
	store := ctx.KVStore(k.storeKey)
	prefix := []byte(types.TaskBidPrefix + taskId + "/")

	var bids []types.TaskBid
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var bid types.TaskBid
		if err := json.Unmarshal(iterator.Value(), &bid); err == nil {
			bids = append(bids, bid)
		}
	}
	return bids
}

func (k Keeper) GetTasksByStatus(ctx sdk.Context, status types.TaskStatus) []types.TaskRequest {
	allTasks := k.GetAllTasks(ctx)

	var filtered []types.TaskRequest
	for _, task := range allTasks {
		if task.Status == status {
			filtered = append(filtered, task)
		}
	}
	return filtered
}

func (k Keeper) GetOpenTasks(ctx sdk.Context) []types.TaskRequest {
	allTasks := k.GetAllTasks(ctx)

	var openTasks []types.TaskRequest
	for _, task := range allTasks {
		if task.IsOpen() {
			openTasks = append(openTasks, task)
		}
	}
	return openTasks
}