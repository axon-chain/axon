package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axon-chain/axon/x/agent/types"
)

func (k Keeper) generateServiceId(agentAddr, name string, blockTime int64) string {
	data := fmt.Sprintf("%s:%s:%d", agentAddr, name, blockTime)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16])
}

func (k Keeper) generateTaskId(requester, title string, blockTime int64) string {
	data := fmt.Sprintf("%s:%s:%d", requester, title, blockTime)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16])
}

func (k Keeper) generateToolId(agentAddr, name string, blockTime int64) string {
	data := fmt.Sprintf("%s:%s:%d", agentAddr, name, blockTime)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16])
}

func (k Keeper) RegisterService(
	ctx sdk.Context,
	agentAddr string,
	name string,
	description string,
	capabilities []string,
	inputTypes, outputTypes []string,
	pricePerCall sdk.Coin,
	endpoint string,
) (string, error) {
	if pricePerCall.IsNegative() {
		return "", fmt.Errorf("price cannot be negative")
	}

	agent, found := k.GetAgent(ctx, agentAddr)
	if !found {
		return "", types.ErrAgentNotFound
	}
	if agent.Status != types.AgentStatus_AGENT_STATUS_ONLINE {
		return "", fmt.Errorf("agent must be online to register services")
	}

	if err := types.ValidateServiceName(name); err != nil {
		return "", err
	}
	if len(description) > types.MaxServiceDescriptionLen {
		return "", fmt.Errorf("description too long")
	}

	exists := k.GetService(ctx, agentAddr, name)
	if exists.Id != "" {
		return "", fmt.Errorf("service with name '%s' already exists", name)
	}

	serviceId := k.generateServiceId(agentAddr, name, ctx.BlockHeight())
	service := types.AgentService{
		Id:               serviceId,
		AgentAddress:     agentAddr,
		Name:            name,
		Description:     description,
		Capabilities:    capabilities,
		InputTypes:      inputTypes,
		OutputTypes:     outputTypes,
		PricePerCall:    pricePerCall,
		Endpoint:       endpoint,
		Status:         types.ServiceStatus_SERVICE_STATUS_ACTIVE,
		CreatedAt:       ctx.BlockHeight(),
		TotalCalls:       0,
		SuccessfulCalls: 0,
		Reputation:      0,
	}

	k.SetService(ctx, service)

	k.IndexService(ctx, service)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"service_registered",
		sdk.NewAttribute("service_id", serviceId),
		sdk.NewAttribute("agent_address", agentAddr),
		sdk.NewAttribute("name", name),
		sdk.NewAttribute("price", pricePerCall.String()),
	))

	return serviceId, nil
}

func (k Keeper) UpdateService(
	ctx sdk.Context,
	agentAddr string,
	serviceId string,
	name string,
	description string,
	pricePerCall sdk.Coin,
	endpoint string,
) error {
	service, found := k.GetServiceById(ctx, serviceId)
	if !found {
		return types.ErrServiceNotFound
	}
	if service.AgentAddress != agentAddr {
		return types.ErrUnauthorized
	}

	if name != "" {
		service.Name = name
	}
	if description != "" {
		service.Description = description
	}
	if !pricePerCall.IsZero() {
		service.PricePerCall = pricePerCall
	}
	if endpoint != "" {
		service.Endpoint = endpoint
	}

	if err := service.Validate(); err != nil {
		return err
	}

	k.SetService(ctx, service)
	return nil
}

func (k Keeper) DisableService(ctx sdk.Context, agentAddr, serviceId string) error {
	service, found := k.GetServiceById(ctx, serviceId)
	if !found {
		return types.ErrServiceNotFound
	}
	if service.AgentAddress != agentAddr {
		return types.ErrUnauthorized
	}

	service.Status = types.ServiceStatus_SERVICE_STATUS_DISABLED
	k.SetService(ctx, service)

	return nil
}

func (k Keeper) SetService(ctx sdk.Context, service types.AgentService) {
	store := ctx.KVStore(k.storeKey)
	key := []byte(types.ServiceIdPrefix + service.Id)
	bz, _ := json.Marshal(&service)
	store.Set(key, bz)
}

func (k Keeper) GetServiceById(ctx sdk.Context, id string) (types.AgentService, bool) {
	store := ctx.KVStore(k.storeKey)
	key := []byte(types.ServiceIdPrefix + id)
	bz := store.Get(key)
	if bz == nil {
		return types.AgentService{}, false
	}
	var service types.AgentService
	if err := json.Unmarshal(bz, &service); err != nil {
		return types.AgentService{}, false
	}
	return service, true
}

func (k Keeper) GetService(ctx sdk.Context, agentAddr, name string) types.AgentService {
	store := ctx.KVStore(k.storeKey)
	prefix := []byte(types.ServiceIdPrefix + "Agent/")
	iterator := store.Iterator(prefix, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var service types.AgentService
		if err := json.Unmarshal(iterator.Value(), &service); err == nil {
			if service.AgentAddress == agentAddr && service.Name == name {
				return service
			}
		}
	}
	return types.AgentService{}
}

func (k Keeper) IndexService(ctx sdk.Context, service types.AgentService) {
	for _, cap := range service.Capabilities {
		store := ctx.KVStore(k.storeKey)
		key := []byte("ServiceCap/" + cap + "/" + service.Id)
		store.Set(key, []byte(service.Id))
	}
}

func (k Keeper) GetServicesByCapability(ctx sdk.Context, capability string) []types.AgentService {
	store := ctx.KVStore(k.storeKey)
	prefix := []byte("ServiceCap/" + capability + "/")
	var services []types.AgentService

	iterator := store.Iterator(prefix, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		serviceId := string(iterator.Value())
		if service, found := k.GetServiceById(ctx, serviceId); found {
			if service.IsActive() {
				services = append(services, service)
			}
		}
	}
	return services
}

func (k Keeper) GetAllServices(ctx sdk.Context) []types.AgentService {
	store := ctx.KVStore(k.storeKey)
	prefix := []byte(types.ServiceIdPrefix)

	var services []types.AgentService
	iterator := store.Iterator(prefix, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var service types.AgentService
		if err := json.Unmarshal(iterator.Value(), &service); err == nil {
			services = append(services, service)
		}
	}
	return services
}

func (k Keeper) GetServicesByAgent(ctx sdk.Context, agentAddr string) []types.AgentService {
	store := ctx.KVStore(k.storeKey)
	prefix := []byte(types.ServiceIdPrefix)

	var services []types.AgentService
	iterator := store.Iterator(prefix, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var service types.AgentService
		if err := json.Unmarshal(iterator.Value(), &service); err == nil {
			if service.AgentAddress == agentAddr {
				services = append(services, service)
			}
		}
	}
	return services
}

func (k Keeper) CallService(
	ctx sdk.Context,
	callerAddr string,
	serviceId string,
	inputData []byte,
	payment sdk.Coin,
) ([]byte, error) {
	service, found := k.GetServiceById(ctx, serviceId)
	if !found {
		return nil, types.ErrServiceNotFound
	}
	if !service.IsActive() {
		return nil, fmt.Errorf("service is not active")
	}
	if payment.IsLT(service.PricePerCall) {
		return nil, fmt.Errorf("insufficient payment: required %s, got %s", service.PricePerCall, payment)
	}

	agent, found := k.GetAgent(ctx, service.AgentAddress)
	if !found || agent.Status != types.AgentStatus_AGENT_STATUS_ONLINE {
		return nil, fmt.Errorf("service provider agent is offline")
	}

	if err := k.PayAgent(ctx, service.AgentAddress, payment); err != nil {
		return nil, fmt.Errorf("failed to pay agent: %v", err)
	}

	serviceCallId := k.generateCallId(serviceId, callerAddr, ctx.BlockHeight())
	call := types.ServiceCall{
		Id:          serviceCallId,
		ServiceId:   serviceId,
		Caller:      callerAddr,
		Requester:  callerAddr,
		InputData:   inputData,
		OutputData:  nil,
		CompletedAt: ctx.BlockHeight(),
		Success:     false,
		Payment:    payment,
	}

	k.SetServiceCall(ctx, call)

	outputData := []byte(fmt.Sprintf(`{"status":"executed","service_id":"%s","message":"Service call initiated - agent should process and respond via separate tx"}`, serviceId))

	call.OutputData = outputData
	call.Success = true
	k.SetServiceCall(ctx, call)

	service.TotalCalls++
	service.SuccessfulCalls++
	k.SetService(ctx, service)

	k.UpdateReputation(ctx, service.AgentAddress, 1)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"service_called",
		sdk.NewAttribute("call_id", serviceCallId),
		sdk.NewAttribute("service_id", serviceId),
		sdk.NewAttribute("caller", callerAddr),
		sdk.NewAttribute("payment", payment.String()),
	))

	return outputData, nil
}

func (k Keeper) SetServiceCall(ctx sdk.Context, call types.ServiceCall) {
	store := ctx.KVStore(k.storeKey)
	key := []byte(types.ServiceCallPrefix + call.Id)
	bz, _ := json.Marshal(&call)
	store.Set(key, bz)
}

func (k Keeper) GetServiceCall(ctx sdk.Context, id string) (types.ServiceCall, bool) {
	store := ctx.KVStore(k.storeKey)
	key := []byte(types.ServiceCallPrefix + id)
	bz := store.Get(key)
	if bz == nil {
		return types.ServiceCall{}, false
	}
	var call types.ServiceCall
	if err := json.Unmarshal(bz, &call); err != nil {
		return types.ServiceCall{}, false
	}
	return call, true
}

func (k Keeper) generateCallId(serviceId, caller string, blockTime int64) string {
	data := fmt.Sprintf("%s:%s:%d", serviceId, caller, blockTime)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16])
}

func (k Keeper) GetServicesByCapabilitySorted(ctx sdk.Context, capability string, limit int) []types.AgentService {
	services := k.GetServicesByCapability(ctx, capability)

	for i := 0; i < len(services)-1; i++ {
		for j := i + 1; j < len(services); j++ {
			if services[j].Reputation > services[i].Reputation ||
				(services[j].Reputation == services[i].Reputation &&
					services[j].SuccessfulCalls > services[i].SuccessfulCalls) {
				services[i], services[j] = services[j], services[i]
			}
		}
	}

	if len(services) > limit {
		services = services[:limit]
	}
	return services
}

func (k Keeper) GetTopServicesByReputation(ctx sdk.Context, limit int) []types.AgentService {
	allServices := k.GetAllServices(ctx)

	var activeServices []types.AgentService
	for _, s := range allServices {
		if s.IsActive() {
			activeServices = append(activeServices, s)
		}
	}

	for i := 0; i < len(activeServices)-1; i++ {
		for j := i + 1; j < len(activeServices); j++ {
			if activeServices[j].Reputation > activeServices[i].Reputation {
				activeServices[i], activeServices[j] = activeServices[j], activeServices[i]
			}
		}
	}

	if len(activeServices) > limit {
		activeServices = activeServices[:limit]
	}
	return activeServices
}