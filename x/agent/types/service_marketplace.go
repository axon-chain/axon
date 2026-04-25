package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ServiceStatus int32

const (
	ServiceStatus_SERVICE_STATUS_UNSPECIFIED ServiceStatus = 0
	ServiceStatus_SERVICE_STATUS_ACTIVE     ServiceStatus = 1
	ServiceStatus_SERVICE_STATUS_PAUSED   ServiceStatus = 2
	ServiceStatus_SERVICE_STATUS_DISABLED   ServiceStatus = 3
)

func (x ServiceStatus) String() string {
	return [...]string{
		"SERVICE_STATUS_UNSPECIFIED",
		"SERVICE_STATUS_ACTIVE",
		"SERVICE_STATUS_PAUSED",
		"SERVICE_STATUS_DISABLED",
	}[x]
}

func ServiceStatusFromString(s string) (ServiceStatus, error) {
	s = strings.ToUpper(s)
	switch s {
	case "ACTIVE":
		return ServiceStatus_SERVICE_STATUS_ACTIVE, nil
	case "PAUSED":
		return ServiceStatus_SERVICE_STATUS_PAUSED, nil
	case "DISABLED":
		return ServiceStatus_SERVICE_STATUS_DISABLED, nil
	default:
		return ServiceStatus_SERVICE_STATUS_UNSPECIFIED, fmt.Errorf("unknown service status: %s", s)
	}
}

type AgentService struct {
	Id              string    `json:"id"`
	AgentAddress    string    `json:"agent_address"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Capabilities   []string  `json:"capabilities"`
	InputTypes     []string  `json:"input_types"`
	OutputTypes    []string `json:"output_types"`
	PricePerCall   sdk.Coin  `json:"price_per_call"`
	Endpoint       string    `json:"endpoint"`
	Status         ServiceStatus `json:"status"`
	CreatedAt      int64     `json:"created_at"`
	TotalCalls     uint64    `json:"total_calls"`
	SuccessfulCalls uint64   `json:"successful_calls"`
	Reputation     uint64    `json:"reputation"`
}

func (s AgentService) Validate() error {
	if s.Id == "" {
		return fmt.Errorf("service id is required")
	}
	if s.AgentAddress == "" {
		return fmt.Errorf("agent address is required")
	}
	if s.Name == "" {
		return fmt.Errorf("service name is required")
	}
	if len(s.Name) > MaxServiceNameLength {
		return fmt.Errorf("service name too long (max %d chars)", MaxServiceNameLength)
	}
	if len(s.Description) > MaxServiceDescriptionLen {
		return fmt.Errorf("service description too long (max %d chars)", MaxServiceDescriptionLen)
	}
	if !s.PricePerCall.IsPositive() {
		return fmt.Errorf("price must be positive")
	}
	if len(s.Capabilities) == 0 {
		return fmt.Errorf("at least one capability required")
	}
	if len(s.InputTypes) == 0 {
		return fmt.Errorf("at least one input type required")
	}
	if len(s.OutputTypes) == 0 {
		return fmt.Errorf("at least one output type required")
	}
	return nil
}

func (s AgentService) IsActive() bool {
	return s.Status == ServiceStatus_SERVICE_STATUS_ACTIVE
}

type ServiceCall struct {
	Id          string    `json:"id"`
	ServiceId   string    `json:"service_id"`
	Caller      string    `json:"caller"`
	Requester   string    `json:"requester"`
	InputData   []byte    `json:"input_data"`
	OutputData  []byte    `json:"output_data"`
	CompletedAt int64     `json:"completed_at"`
	Success     bool      `json:"success"`
	Payment     sdk.Coin  `json:"payment"`
}

type TaskStatus int32

const (
	TaskStatus_TASK_STATUS_UNSPECIFIED  TaskStatus = 0
	TaskStatus_TASK_STATUS_OPEN         TaskStatus = 1
	TaskStatus_TASK_STATUS_BIDDING     TaskStatus = 2
	TaskStatus_TASK_STATUS_IN_PROGRESS    TaskStatus = 3
	TaskStatus_TASK_STATUS_COMPLETED     TaskStatus = 4
	TaskStatus_TASK_STATUS_CANCELLED    TaskStatus = 5
	TaskStatus_TASK_STATUS_FAILED      TaskStatus = 6
)

func (x TaskStatus) String() string {
	return [...]string{
		"TASK_STATUS_UNSPECIFIED",
		"TASK_STATUS_OPEN",
		"TASK_STATUS_BIDDING",
		"TASK_STATUS_IN_PROGRESS",
		"TASK_STATUS_COMPLETED",
		"TASK_STATUS_CANCELLED",
		"TASK_STATUS_FAILED",
	}[x]
}

func TaskStatusFromString(s string) (TaskStatus, error) {
	s = strings.ToUpper(s)
	switch s {
	case "OPEN":
		return TaskStatus_TASK_STATUS_OPEN, nil
	case "BIDDING":
		return TaskStatus_TASK_STATUS_BIDDING, nil
	case "IN_PROGRESS":
		return TaskStatus_TASK_STATUS_IN_PROGRESS, nil
	case "COMPLETED":
		return TaskStatus_TASK_STATUS_COMPLETED, nil
	case "CANCELLED":
		return TaskStatus_TASK_STATUS_CANCELLED, nil
	case "FAILED":
		return TaskStatus_TASK_STATUS_FAILED, nil
	default:
		return TaskStatus_TASK_STATUS_UNSPECIFIED, fmt.Errorf("unknown task status: %s", s)
	}
}

type TaskRequest struct {
	Id                   string     `json:"id"`
	Requester            string     `json:"requester"`
	Title                string     `json:"title"`
	Description         string     `json:"description"`
	RequiredCapabilities []string   `json:"required_capabilities"`
	Budget               sdk.Coin   `json:"budget"`
	DeadlineBlock       int64      `json:"deadline_block"`
	Status               TaskStatus `json:"status"`
	SelectedAgent        string     `json:"selected_agent"`
	CreatedAt            int64      `json:"created_at"`
}

func (t TaskRequest) Validate() error {
	if t.Id == "" {
		return fmt.Errorf("task id is required")
	}
	if t.Requester == "" {
		return fmt.Errorf("requester is required")
	}
	if t.Title == "" {
		return fmt.Errorf("title is required")
	}
	if len(t.Title) > MaxTaskTitleLength {
		return fmt.Errorf("task title too long (max %d chars)", MaxTaskTitleLength)
	}
	if len(t.Description) > MaxTaskDescriptionLen {
		return fmt.Errorf("task description too long (max %d chars)", MaxTaskDescriptionLen)
	}
	if !t.Budget.IsPositive() {
		return fmt.Errorf("budget must be positive")
	}
	if t.DeadlineBlock <= 0 {
		return fmt.Errorf("deadline must be positive")
	}
	if len(t.RequiredCapabilities) == 0 {
		return fmt.Errorf("at least one required capability")
	}
	return nil
}

func (t TaskRequest) IsOpen() bool {
	return t.Status == TaskStatus_TASK_STATUS_OPEN || t.Status == TaskStatus_TASK_STATUS_BIDDING
}

type TaskBid struct {
	TaskId        string    `json:"task_id"`
	AgentAddress string    `json:"agent_address"`
	Proposal     string    `json:"proposal"`
	Price        sdk.Coin  `json:"price"`
	SubmittedAt  int64     `json:"submitted_at"`
	Accepted     bool      `json:"accepted"`
}

func (b TaskBid) Validate() error {
	if b.TaskId == "" {
		return fmt.Errorf("task id is required")
	}
	if b.AgentAddress == "" {
		return fmt.Errorf("agent address is required")
	}
	if b.Proposal == "" {
		return fmt.Errorf("proposal is required")
	}
	if len(b.Proposal) > MaxProposalLength {
		return fmt.Errorf("proposal too long (max %d chars)", MaxProposalLength)
	}
	if !b.Price.IsPositive() {
		return fmt.Errorf("price must be positive")
	}
	return nil
}

type TaskCompletion struct {
	TaskId         string `json:"task_id"`
	CompletionData string `json:"completion_data"`
	CompletedAt   int64  `json:"completed_at"`
}

type ToolStatus int32

const (
	ToolStatus_TOOL_STATUS_UNSPECIFIED ToolStatus = 0
	ToolStatus_TOOL_STATUS_ACTIVE       ToolStatus = 1
	ToolStatus_TOOL_STATUS_DISABLED   ToolStatus = 2
)

func (x ToolStatus) String() string {
	return [...]string{
		"TOOL_STATUS_UNSPECIFIED",
		"TOOL_STATUS_ACTIVE",
		"TOOL_STATUS_DISABLED",
	}[x]
}

func ToolStatusFromString(s string) (ToolStatus, error) {
	s = strings.ToUpper(s)
	switch s {
	case "ACTIVE":
		return ToolStatus_TOOL_STATUS_ACTIVE, nil
	case "DISABLED":
		return ToolStatus_TOOL_STATUS_DISABLED, nil
	default:
		return ToolStatus_TOOL_STATUS_UNSPECIFIED, fmt.Errorf("unknown tool status: %s", s)
	}
}

type ToolDefinition struct {
	Id           string     `json:"id"`
	AgentAddress string     `json:"agent_address"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	InputSchema string     `json:"input_schema"`
	OutputSchema string    `json:"output_schema"`
	Price       sdk.Coin   `json:"price"`
	IsPublic    bool       `json:"is_public"`
	Status     ToolStatus `json:"status"`
	CreatedAt  int64      `json:"created_at"`
}

func (t ToolDefinition) Validate() error {
	if t.Id == "" {
		return fmt.Errorf("tool id is required")
	}
	if t.AgentAddress == "" {
		return fmt.Errorf("agent address is required")
	}
	if t.Name == "" {
		return fmt.Errorf("tool name is required")
	}
	if len(t.Name) > MaxServiceNameLength {
		return fmt.Errorf("tool name too long (max %d chars)", MaxServiceNameLength)
	}
	if len(t.Description) > MaxServiceDescriptionLen {
		return fmt.Errorf("tool description too long (max %d chars)", MaxServiceDescriptionLen)
	}
	if t.InputSchema == "" {
		return fmt.Errorf("input schema is required")
	}
	if t.OutputSchema == "" {
		return fmt.Errorf("output schema is required")
	}
	if !t.Price.IsPositive() {
		return fmt.Errorf("price must be positive")
	}
	return nil
}

func (t ToolDefinition) IsActive() bool {
	return t.Status == ToolStatus_TOOL_STATUS_ACTIVE
}

type ToolCall struct {
	CallId     string    `json:"call_id"`
	Caller     string    `json:"caller"`
	ToolId     string    `json:"tool_id"`
	InputData  []byte    `json:"input_data"`
	OutputData []byte    `json:"output_data"`
	Payment   sdk.Coin   `json:"payment"`
	ExecutedAt int64     `json:"executed_at"`
	Success    bool      `json:"success"`
}

func ValidateServiceName(name string) error {
	if name == "" {
		return ErrServiceNameInvalid
	}
	if len(name) > MaxServiceNameLength {
		return fmt.Errorf("name too long (max %d chars)", MaxServiceNameLength)
	}
	validNameChars := "abcdefghijklmnopqrstuvwxyz0123456789-_"
	lowercaseName := strings.ToLower(name)
	for _, c := range lowercaseName {
		if !strings.Contains(validNameChars, string(c)) {
			return fmt.Errorf("name can only contain lowercase letters, numbers, - and _")
		}
	}
	return nil
}

func ValidateProposal(proposal string) error {
	if proposal == "" {
		return fmt.Errorf("proposal is required")
	}
	if len(proposal) > MaxProposalLength {
		return fmt.Errorf("proposal too long (max %d chars)", MaxProposalLength)
	}
	return nil
}