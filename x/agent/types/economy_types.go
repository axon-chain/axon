package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/cosmos/gogoproto/proto"
)

func (m *MsgRegister) XXX_Workaround()                {}
func (m *MsgAddStake) XXX_Workaround()               {}
func (m *MsgUpdateAgent) XXX_Workaround()            {}
func (m *MsgHeartbeat) XXX_Workaround()            {}
func (m *MsgDeregister) XXX_Workaround()          {}
func (m *MsgReduceStake) XXX_Workaround()          {}
func (m *MsgClaimReducedStake) XXX_Workaround()      {}
func (m *MsgSubmitAIChallengeResponse) XXX_Workaround() {}
func (m *MsgRevealAIChallengeResponse) XXX_Workaround() {}

type MsgRegisterService struct {
	Sender        string `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty"`
	Name         string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Description string `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"`
	Capabilities []string `protobuf:"bytes,4,rep,name=capabilities,proto3" json:"capabilities,omitempty"`
	InputTypes   []string `protobuf:"bytes,5,rep,name=input_types,json=inputTypes,proto3" json:"input_types,omitempty"`
	OutputTypes  []string `protobuf:"bytes,6,rep,name=output_types,json=outputTypes,proto3" json:"output_types,omitempty"`
	PricePerCall sdk.Coin `protobuf:"bytes,7,opt,name=price_per_call,json=pricePerCall,proto3" json:"price_per_call"`
	Endpoint    string `protobuf:"bytes,8,opt,name=endpoint,proto3" json:"endpoint,omitempty"`
}

func (m *MsgRegisterService) Reset()         { *m = MsgRegisterService{} }
func (m *MsgRegisterService) String() string { return proto.CompactTextString(m) }
func (*MsgRegisterService) ProtoMessage()    {}

func (m *MsgRegisterService) GetSigners() []sdk.AccAddress {
	return nil
}

type MsgRegisterServiceResponse struct {
	ServiceId string `protobuf:"bytes,1,opt,name=service_id,json=serviceId,proto3" json:"service_id,omitempty"`
}

func (m *MsgRegisterServiceResponse) Reset()         { *m = MsgRegisterServiceResponse{} }
func (m *MsgRegisterServiceResponse) String() string { return proto.CompactTextString(m) }
func (*MsgRegisterServiceResponse) ProtoMessage()     {}

type MsgUpdateService struct {
	Sender       string `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty"`
	ServiceId   string `protobuf:"bytes,2,opt,name=service_id,json=serviceId,proto3" json:"service_id,omitempty"`
	Name        string `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	Description string `protobuf:"bytes,4,opt,name=description,proto3" json:"description,omitempty"`
	PricePerCall sdk.Coin `protobuf:"bytes,5,opt,name=price_per_call,json=pricePerCall,proto3" json:"price_per_call"`
	Endpoint    string `protobuf:"bytes,6,opt,name=endpoint,proto3" json:"endpoint,omitempty"`
}

func (m *MsgUpdateService) Reset()         { *m = MsgUpdateService{} }
func (m *MsgUpdateService) String() string { return proto.CompactTextString(m) }
func (*MsgUpdateService) ProtoMessage()    {}

type MsgUpdateServiceResponse struct{}

func (m *MsgUpdateServiceResponse) Reset()         { *m = MsgUpdateServiceResponse{} }
func (m *MsgUpdateServiceResponse) String() string { return proto.CompactTextString(m) }
func (*MsgUpdateServiceResponse) ProtoMessage() {}

type MsgDisableService struct {
	Sender    string `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty"`
	ServiceId string `protobuf:"bytes,2,opt,name=service_id,json=serviceId,proto3" json:"service_id,omitempty"`
}

func (m *MsgDisableService) Reset()         { *m = MsgDisableService{} }
func (m *MsgDisableService) String() string { return proto.CompactTextString(m) }
func (*MsgDisableService) ProtoMessage()  {}

type MsgDisableServiceResponse struct{}

func (m *MsgDisableServiceResponse) Reset()         { *m = MsgDisableServiceResponse{} }
func (m *MsgDisableServiceResponse) String() string { return proto.CompactTextString(m) }
func (*MsgDisableServiceResponse) ProtoMessage() {}

type MsgCallService struct {
	Sender     string `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty"`
	ServiceId  string `protobuf:"bytes,2,opt,name=service_id,json=serviceId,proto3" json:"service_id,omitempty"`
	InputData []byte `protobuf:"bytes,3,opt,name=input_data,json=inputData,proto3" json:"input_data,omitempty"`
	Payment   sdk.Coin `protobuf:"bytes,4,opt,name=payment,proto3" json:"payment"`
}

func (m *MsgCallService) Reset()         { *m = MsgCallService{} }
func (m *MsgCallService) String() string { return proto.CompactTextString(m) }
func (*MsgCallService) ProtoMessage() {}

type MsgCallServiceResponse struct {
	OutputData []byte `protobuf:"bytes,1,opt,name=output_data,json=outputData,proto3" json:"output_data,omitempty"`
}

func (m *MsgCallServiceResponse) Reset()         { *m = MsgCallServiceResponse{} }
func (m *MsgCallServiceResponse) String() string { return proto.CompactTextString(m) }
func (*MsgCallServiceResponse) ProtoMessage()  {}

type MsgCreateTask struct {
	Sender             string `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty"`
	Title              string `protobuf:"bytes,2,opt,name=title,proto3" json:"title,omitempty"`
	Description       string `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"`
	RequiredCapabilities []string `protobuf:"bytes,4,rep,name=required_capabilities,json=requiredCapabilities,proto3" json:"required_capabilities,omitempty"`
	Budget             sdk.Coin `protobuf:"bytes,5,opt,name=budget,proto3" json:"budget"`
	DeadlineBlocks     int64  `protobuf:"varint,6,opt,name=deadline_blocks,json=deadlineBlocks,proto3" json:"deadline_blocks,omitempty"`
}

func (m *MsgCreateTask) Reset()         { *m = MsgCreateTask{} }
func (m *MsgCreateTask) String() string { return proto.CompactTextString(m) }
func (*MsgCreateTask) ProtoMessage()    {}

type MsgCreateTaskResponse struct {
	TaskId string `protobuf:"bytes,1,opt,name=task_id,json=taskId,proto3" json:"task_id,omitempty"`
}

func (m *MsgCreateTaskResponse) Reset()         { *m = MsgCreateTaskResponse{} }
func (m *MsgCreateTaskResponse) String() string { return proto.CompactTextString(m) }
func (*MsgCreateTaskResponse) ProtoMessage() {}

type MsgCancelTask struct {
	Sender string `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty"`
	TaskId string `protobuf:"bytes,2,opt,name=task_id,json=taskId,proto3" json:"task_id,omitempty"`
}

func (m *MsgCancelTask) Reset()         { *m = MsgCancelTask{} }
func (m *MsgCancelTask) String() string { return proto.CompactTextString(m) }
func (*MsgCancelTask) ProtoMessage()  {}

type MsgCancelTaskResponse struct{}

func (m *MsgCancelTaskResponse) Reset()         { *m = MsgCancelTaskResponse{} }
func (m *MsgCancelTaskResponse) String() string { return proto.CompactTextString(m) }
func (*MsgCancelTaskResponse) ProtoMessage() {}

type MsgSubmitBid struct {
	Sender   string  `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty"`
	TaskId  string  `protobuf:"bytes,2,opt,name=task_id,json=taskId,proto3" json:"task_id,omitempty"`
	Proposal string  `protobuf:"bytes,3,opt,name=proposal,proto3" json:"proposal,omitempty"`
	Price   sdk.Coin `protobuf:"bytes,4,opt,name=price,proto3" json:"price"`
}

func (m *MsgSubmitBid) Reset()         { *m = MsgSubmitBid{} }
func (m *MsgSubmitBid) String() string { return proto.CompactTextString(m) }
func (*MsgSubmitBid) ProtoMessage()   {}

type MsgSubmitBidResponse struct{}

func (m *MsgSubmitBidResponse) Reset()         { *m = MsgSubmitBidResponse{} }
func (m *MsgSubmitBidResponse) String() string { return proto.CompactTextString(m) }
func (*MsgSubmitBidResponse) ProtoMessage() {}

type MsgSelectBid struct {
	Sender       string `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty"`
	TaskId       string `protobuf:"bytes,2,opt,name=task_id,json=taskId,proto3" json:"task_id,omitempty"`
	AgentAddress string `protobuf:"bytes,3,opt,name=agent_address,json=agentAddress,proto3" json:"agent_address,omitempty"`
}

func (m *MsgSelectBid) Reset()         { *m = MsgSelectBid{} }
func (m *MsgSelectBid) String() string { return proto.CompactTextString(m) }
func (*MsgSelectBid) ProtoMessage()    {}

type MsgSelectBidResponse struct{}

func (m *MsgSelectBidResponse) Reset()         { *m = MsgSelectBidResponse{} }
func (m *MsgSelectBidResponse) String() string { return proto.CompactTextString(m) }
func (*MsgSelectBidResponse) ProtoMessage() {}

type MsgCompleteTask struct {
	Sender          string `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty"`
	TaskId         string `protobuf:"bytes,2,opt,name=task_id,json=taskId,proto3" json:"task_id,omitempty"`
	CompletionData string `protobuf:"bytes,3,opt,name=completion_data,json=completionData,proto3" json:"completion_data,omitempty"`
}

func (m *MsgCompleteTask) Reset()         { *m = MsgCompleteTask{} }
func (m *MsgCompleteTask) String() string { return proto.CompactTextString(m) }
func (*MsgCompleteTask) ProtoMessage()   {}

type MsgCompleteTaskResponse struct{}

func (m *MsgCompleteTaskResponse) Reset()         { *m = MsgCompleteTaskResponse{} }
func (m *MsgCompleteTaskResponse) String() string { return proto.CompactTextString(m) }
func (*MsgCompleteTaskResponse) ProtoMessage() {}

type MsgRegisterTool struct {
	Sender       string `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty"`
	Name         string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Description string `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"`
	InputSchema string `protobuf:"bytes,4,opt,name=input_schema,json=inputSchema,proto3" json:"input_schema,omitempty"`
	OutputSchema string `protobuf:"bytes,5,opt,name=output_schema,json=outputSchema,proto3" json:"output_schema,omitempty"`
	Price       sdk.Coin `protobuf:"bytes,6,opt,name=price,proto3" json:"price"`
	IsPublic    bool   `protobuf:"varint,7,opt,name=is_public,json=isPublic,proto3" json:"is_public,omitempty"`
}

func (m *MsgRegisterTool) Reset()         { *m = MsgRegisterTool{} }
func (m *MsgRegisterTool) String() string { return proto.CompactTextString(m) }
func (*MsgRegisterTool) ProtoMessage()    {}

type MsgRegisterToolResponse struct {
	ToolId string `protobuf:"bytes,1,opt,name=tool_id,json=toolId,proto3" json:"tool_id,omitempty"`
}

func (m *MsgRegisterToolResponse) Reset()         { *m = MsgRegisterToolResponse{} }
func (m *MsgRegisterToolResponse) String() string { return proto.CompactTextString(m) }
func (*MsgRegisterToolResponse) ProtoMessage() {}

type MsgCallTool struct {
	Sender    string `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty"`
	ToolId    string `protobuf:"bytes,2,opt,name=tool_id,json=toolId,proto3" json:"tool_id,omitempty"`
	InputData []byte `protobuf:"bytes,3,opt,name=input_data,json=inputData,proto3" json:"input_data,omitempty"`
	Payment  sdk.Coin `protobuf:"bytes,4,opt,name=payment,proto3" json:"payment"`
}

func (m *MsgCallTool) Reset()         { *m = MsgCallTool{} }
func (m *MsgCallTool) String() string { return proto.CompactTextString(m) }
func (*MsgCallTool) ProtoMessage()    {}

type MsgCallToolResponse struct {
	OutputData []byte `protobuf:"bytes,1,opt,name=output_data,json=outputData,proto3" json:"output_data,omitempty"`
}

func (m *MsgCallToolResponse) Reset()         { *m = MsgCallToolResponse{} }
func (m *MsgCallToolResponse) String() string { return proto.CompactTextString(m) }
func (*MsgCallToolResponse) ProtoMessage()    {}

type MsgSubmitL2Report struct {
	Sender   string `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty"`
	Target   string `protobuf:"bytes,2,opt,name=target,proto3" json:"target,omitempty"`
	Score    int32  `protobuf:"varint,3,opt,name=score,proto3" json:"score,omitempty"`
	Evidence string `protobuf:"bytes,4,opt,name=evidence,proto3" json:"evidence,omitempty"`
	Reason   string `protobuf:"bytes,5,opt,name=reason,proto3" json:"reason,omitempty"`
}

func (m *MsgSubmitL2Report) Reset()         { *m = MsgSubmitL2Report{} }
func (m *MsgSubmitL2Report) String() string { return proto.CompactTextString(m) }
func (*MsgSubmitL2Report) ProtoMessage()    {}

func (m *MsgSubmitL2Report) GetSigners() []sdk.AccAddress {
	if m.Sender == "" {
		return nil
	}
	addr, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return nil
	}
	return []sdk.AccAddress{addr}
}

type MsgSubmitL2ReportResponse struct{}

func (m *MsgSubmitL2ReportResponse) Reset()         { *m = MsgSubmitL2ReportResponse{} }
func (m *MsgSubmitL2ReportResponse) String() string { return proto.CompactTextString(m) }
func (*MsgSubmitL2ReportResponse) ProtoMessage() {}

type QueryServicesRequest struct {
	Capability   string `protobuf:"bytes,1,opt,name=capability,proto3" json:"capability,omitempty"`
	AgentAddress string `protobuf:"bytes,2,opt,name=agent_address,json=agentAddress,proto3" json:"agent_address,omitempty"`
	Limit       int64  `protobuf:"varint,3,opt,name=limit,proto3" json:"limit,omitempty"`
}

func (m *QueryServicesRequest) Reset()         { *m = QueryServicesRequest{} }
func (m *QueryServicesRequest) String() string { return proto.CompactTextString(m) }
func (*QueryServicesRequest) ProtoMessage() {}

type QueryServicesResponse struct {
	Services []AgentService `protobuf:"bytes,1,rep,name=services,proto3" json:"services"`
}

func (m *QueryServicesResponse) Reset()         { *m = QueryServicesResponse{} }
func (m *QueryServicesResponse) String() string { return proto.CompactTextString(m) }
func (*QueryServicesResponse) ProtoMessage() {}

type QueryServiceRequest struct {
	ServiceId string `protobuf:"bytes,1,opt,name=service_id,json=serviceId,proto3" json:"service_id,omitempty"`
}

func (m *QueryServiceRequest) Reset()         { *m = QueryServiceRequest{} }
func (m *QueryServiceRequest) String() string { return proto.CompactTextString(m) }
func (*QueryServiceRequest) ProtoMessage()   {}

type QueryServiceResponse struct {
	Service AgentService `protobuf:"bytes,1,opt,name=service,proto3" json:"service,omitempty"`
}

func (m *QueryServiceResponse) Reset()         { *m = QueryServiceResponse{} }
func (m *QueryServiceResponse) String() string { return proto.CompactTextString(m) }
func (*QueryServiceResponse) ProtoMessage()    {}

type QueryServiceCallsRequest struct {
	ServiceId string `protobuf:"bytes,1,opt,name=service_id,json=serviceId,proto3" json:"service_id,omitempty"`
	Limit    int64  `protobuf:"varint,2,opt,name=limit,proto3" json:"limit,omitempty"`
}

func (m *QueryServiceCallsRequest) Reset()         { *m = QueryServiceCallsRequest{} }
func (m *QueryServiceCallsRequest) String() string { return proto.CompactTextString(m) }
func (*QueryServiceCallsRequest) ProtoMessage()  {}

type QueryServiceCallsResponse struct {
	Calls []ServiceCall `protobuf:"bytes,1,rep,name=calls,proto3" json:"calls"`
}

func (m *QueryServiceCallsResponse) Reset()         { *m = QueryServiceCallsResponse{} }
func (m *QueryServiceCallsResponse) String() string { return proto.CompactTextString(m) }
func (*QueryServiceCallsResponse) ProtoMessage() {}

type QueryTasksRequest struct {
	Status    TaskStatus `protobuf:"varint,1,opt,name=status,proto3,enum=axon.agent.v1.TaskStatus" json:"status,omitempty"`
	Requester string    `protobuf:"bytes,2,opt,name=requester,proto3" json:"requester,omitempty"`
	Limit     int64    `protobuf:"varint,3,opt,name=limit,proto3" json:"limit,omitempty"`
}

func (m *QueryTasksRequest) Reset()         { *m = QueryTasksRequest{} }
func (m *QueryTasksRequest) String() string { return proto.CompactTextString(m) }
func (*QueryTasksRequest) ProtoMessage()   {}

type QueryTasksResponse struct {
	Tasks []TaskRequest `protobuf:"bytes,1,rep,name=tasks,proto3" json:"tasks"`
}

func (m *QueryTasksResponse) Reset()         { *m = QueryTasksResponse{} }
func (m *QueryTasksResponse) String() string { return proto.CompactTextString(m) }
func (*QueryTasksResponse) ProtoMessage()    {}

type QueryTaskRequest struct {
	TaskId string `protobuf:"bytes,1,opt,name=task_id,json=taskId,proto3" json:"task_id,omitempty"`
}

func (m *QueryTaskRequest) Reset()         { *m = QueryTaskRequest{} }
func (m *QueryTaskRequest) String() string { return proto.CompactTextString(m) }
func (*QueryTaskRequest) ProtoMessage()    {}

type QueryTaskResponse struct {
	Task TaskRequest `protobuf:"bytes,1,opt,name=task,proto3" json:"task,omitempty"`
}

func (m *QueryTaskResponse) Reset()         { *m = QueryTaskResponse{} }
func (m *QueryTaskResponse) String() string { return proto.CompactTextString(m) }
func (*QueryTaskResponse) ProtoMessage()    {}

type QueryTaskBidsRequest struct {
	TaskId string `protobuf:"bytes,1,opt,name=task_id,json=taskId,proto3" json:"task_id,omitempty"`
	Limit  int64  `protobuf:"varint,2,opt,name=limit,proto3" json:"limit,omitempty"`
}

func (m *QueryTaskBidsRequest) Reset()         { *m = QueryTaskBidsRequest{} }
func (m *QueryTaskBidsRequest) String() string { return proto.CompactTextString(m) }
func (*QueryTaskBidsRequest) ProtoMessage()     {}

type QueryTaskBidsResponse struct {
	Bids []TaskBid `protobuf:"bytes,1,rep,name=bids,proto3" json:"bids"`
}

func (m *QueryTaskBidsResponse) Reset()         { *m = QueryTaskBidsResponse{} }
func (m *QueryTaskBidsResponse) String() string { return proto.CompactTextString(m) }
func (*QueryTaskBidsResponse) ProtoMessage()    {}

type QueryToolsRequest struct {
	AgentAddress string `protobuf:"bytes,1,opt,name=agent_address,json=agentAddress,proto3" json:"agent_address,omitempty"`
	PublicOnly  bool    `protobuf:"varint,2,opt,name=public_only,json=publicOnly,proto3" json:"public_only,omitempty"`
	Limit      int64   `protobuf:"varint,3,opt,name=limit,proto3" json:"limit,omitempty"`
}

func (m *QueryToolsRequest) Reset()         { *m = QueryToolsRequest{} }
func (m *QueryToolsRequest) String() string { return proto.CompactTextString(m) }
func (*QueryToolsRequest) ProtoMessage()   {}

type QueryToolsResponse struct {
	Tools []ToolDefinition `protobuf:"bytes,1,rep,name=tools,proto3" json:"tools"`
}

func (m *QueryToolsResponse) Reset()         { *m = QueryToolsResponse{} }
func (m *QueryToolsResponse) String() string { return proto.CompactTextString(m) }
func (*QueryToolsResponse) ProtoMessage()    {}

type QueryToolRequest struct {
	ToolId string `protobuf:"bytes,1,opt,name=tool_id,json=toolId,proto3" json:"tool_id,omitempty"`
}

func (m *QueryToolRequest) Reset()         { *m = QueryToolRequest{} }
func (m *QueryToolRequest) String() string { return proto.CompactTextString(m) }
func (*QueryToolRequest) ProtoMessage()    {}

type QueryToolResponse struct {
	Tool ToolDefinition `protobuf:"bytes,1,opt,name=tool,proto3" json:"tool,omitempty"`
}

func (m *QueryToolResponse) Reset()         { *m = QueryToolResponse{} }
func (m *QueryToolResponse) String() string { return proto.CompactTextString(m) }
func (*QueryToolResponse) ProtoMessage()    {}