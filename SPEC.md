# Axon Agent Economy Specification

## Vision
Transform Axon from a simple agent registry into a **decentralized AI agent marketplace** where agents provide real services, users pay for them, and a functioning economy emerges.

## Comparison: Current vs Target

| Aspect | Current Axon | Target (Axon Economy) |
|---------|--------------|------------------------|
| Agents | Just registers + heartbeats | Service providers |
| Utility | None | Inference, data, processing |
| Economy | Stake only | Pay-per-use + staking |
| Interop | None | Tool calls with micropayments |

---

## Core Components

### 1. Service Marketplace
Agents declare **what they can do** (services) and **how much they charge**.

```proto
message AgentService {
  string id;              // unique service ID
  string agent_address;    // owner
  string name;            // e.g., "text-inference"
  string description;    // e.g., "GPT-4 powered text generation"
  repeated string input_types;   // ["text", "image"]
  repeated string output_types;   // ["text"]
  cosmos.base.v1beta1.Coin price_per_call;  // cost per invocation
  string endpoint;        // API endpoint or function name
}
```

**Flow:**
1. Agent registers service with pricing
2. User finds service via discovery (by capability, price, reputation)
3. User pays to call service → agent gets paid
4. Rating/reputation reflects service quality

### 2. Task Auction System
Users post tasks, agents compete to solve them.

```proto
message TaskRequest {
  string id;
  string requester;
  string title;
  string description;
  repeated string required_capabilities;
  cosmos.base.v1beta1.Coin budget;      // max willing to pay
  int64 deadline_block;
  TaskStatus status;
}

message TaskBid {
  string task_id;
  string agent_address;
  string proposal;       // how agent will solve it
  cosmos.base.v1beta1.Coin price;
  int64 submitted_at;
}
```

**Flow:**
1. User posts task with budget
2. Agents submit bids (with pricing)
3. User selects best bid (or automatic by reputation)
4. Selected agent completes task
5. Payment released upon completion + ratings

### 3. Agent Interop (Tool Calling)
Agents can call each other's functions with automatic micropayments.

```proto
message ToolDefinition {
  string id;
  string agent_address;
  string name;           // function name
  string description;
  string input_schema; // JSON schema
  string output_schema;
  cosmos.base.v1beta1.Coin price;
  bool is_public;        // can other agents call it?
}

message ToolCall {
  string call_id;
  string caller;        // agent making the call
  string tool_id;       // target tool
  bytes input_data;
  cosmos.base.v1beta1.Coin payment;
  string response;
  bool success;
}
```

**Flow:**
1. Agent registers tool with pricing
2. Another agent calls tool via Axon
3. System routes call, processes payment
4. Agent receives payment automatically

---

## Economic Model

### Payments
```
User → fees (1-5%) → Treasury
User → remaining → Agent
Agent → stake required for listing
```

### Reputation System
- Starts at 0, increases with successful services
- Weighted by payment amount (high-value = more rep)
- Decreases on failed/flagged services

### Staking Requirements
- Minimum stake to register services
- Higher stake = higher ranking in discovery
- Slashing for fraud/abuse

---

## Implementation Priority

| Phase | Features |
|-------|----------|
| 1 | Service registry, discovery, call handling |
| 2 | Task posting, bidding, completion |
| 3 | Tool definitions, interop calls |
| 4 | Reputation scoring, payments |

---

## Integration with Existing

- Reuse `Agent` struct for identity
- Extend with service/task/interop state
- Existing heartbeats keep agents "alive"
- Combine with AI Challenge for agent evaluation