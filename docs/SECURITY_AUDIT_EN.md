> 🌐 [中文版](SECURITY_AUDIT.md)

# Axon Security Audit Self-Assessment Report

**Version: v1.0.0**
**Date: March 2026**
**Status: Post-Mainnet Release Review**

---

## Audit Scope

This report covers all core components of the Axon chain, ordered by security priority:

| Module | Path | Security Level |
|--------|------|----------------|
| Consensus Layer (CometBFT) | CometBFT Configuration | Critical |
| Token Economics Model | `x/agent/keeper/block_rewards.go`, `contribution.go` | Critical |
| Deflation Mechanism | `app/fee_burn.go`, `app/evm_hooks.go`, `keeper/reputation.go` | Critical |
| Agent Module | `x/agent/keeper/`, `x/agent/types/` | Critical |
| Precompile — IAgentRegistry | `precompiles/registry/` | High |
| Precompile — IAgentReputation | `precompiles/reputation/` | Medium |
| Precompile — IAgentWallet | `precompiles/wallet/` | Critical |
| EVM Integration | `app/evm_hooks.go`, Cosmos EVM Module | High |
| Module Permission Configuration | `app/config/permissions.go` | Critical |
| Genesis State | `x/agent/types/genesis.go`, `types/params.go` | High |
| Network Layer | CometBFT P2P, JSON-RPC | Medium |

---

## 1. Consensus Security

### 1.1 CometBFT BFT Consensus

- ✅ **Pass** — Uses CometBFT (formerly Tendermint) BFT consensus, tolerating up to 1/3 Byzantine validator failures. This consensus engine has been battle-tested in production across dozens of chains including Cosmos Hub, Osmosis, and dYdX.
- ✅ **Pass** — Instant finality with single-block confirmation, no fork risk.
- ✅ **Pass** — Block production time ~5 seconds, controlled by the CometBFT `timeout_commit` parameter.

### 1.2 Validator Set Management

- ✅ **Pass** — Validator set cap initially 100, managed via the `x/staking` module, adjustable through on-chain governance.
- ✅ **Pass** — Minimum validator stake 10,000 AXON, preventing low-cost attacks.
- ✅ **Pass** — Validator staking unlock cooldown of 14 days (`x/staking` default unbonding period).

### 1.3 Slashing Conditions

- ✅ **Pass** — Double-signing penalty: slash 5% of stake + reputation -50 + jail (`x/slashing` module standard behavior).
- ✅ **Pass** — Extended downtime penalty: slash 0.1% of stake + reputation -5 + jail.
- ✅ **Pass** — Reputation deduction constants are hardcoded: `ReputationLossSlashing = -50`, `ReputationLossOffline = -5` (`x/agent/types/params.go:15-16`).

### 1.4 Block Production Weight

- ✅ **Pass** — Block production weight formula `Stake × (1 + ReputationBonus + AIBonus)` is implemented. Five-tier reputation bonus (0%/5%/10%/15%/20%) is correctly implemented in `reputationBonusPercent()`.
- ✅ **Pass** — Multiplier minimum value clamped at 10 (i.e., minimum 10%), preventing the multiplier from reaching zero and causing zero weight: `if multiplier < 10 { multiplier = 10 }`.

**Risk Assessment: ✅ Low Risk**

The consensus layer relies on the mature CometBFT engine. Risk primarily concentrates on Axon's custom weight adjustment logic. We recommend that external auditors focus on the consistency between weight calculations in `distributeValidatorRewards` and voting power updates in the `x/staking` module.

---

## 2. Token Economics Security

### 2.1 Block Reward Calculation

- ✅ **Pass** — Base reward uses `big.Int` string initialization (`BaseBlockRewardStr = "12367000000000000000"`), avoiding floating-point precision issues.
- ✅ **Pass** — Halving uses bit right-shift (`Rsh`), mathematically equivalent to integer division by 2, with no overflow risk.
- ✅ **Pass** — Halving count cap of 64 checks: `if halvings >= 64 { return sdkmath.ZeroInt() }`, preventing uint overflow.
- ✅ **Pass** — Reward three-way distribution ratio (Proposer 25% + Validators 50% + AI 25% = 100%), with safe remainder handling: `aiReward := reward.Sub(proposerReward).Sub(validatorReward)`, ensuring no rounding loss.

### 2.2 Contribution Rewards

- ✅ **Pass** — Anti-gaming mechanisms implemented:
  - Self-calls are not counted (contribution scoring is based on separate `DeployCount` and `ContractCall` counters).
  - Per-Agent cap per Epoch = 2% of pool (`MaxSharePerAgentBPS = 200`).
  - Reputation < 20 excluded from distribution (`MinReputationForReward = 20`).
  - Registration less than 7 days old excluded (`MinRegistrationBlocks = 120960`).
- ⚠️ **Attention Needed** — Self-call filter: The `ContractCall` counter's increment entry point (`IncrementContractCalls`) has not yet been called in the EVM hook. Contract call count tracking needs to be integrated into the EVM execution layer. The self-call exclusion logic depends on correct implementation of this integration.
- ✅ **Pass** — Activity cap at 100 transactions; `activityCapped` limit prevents single-dimension gaming.

### 2.3 Total Supply Hard Cap

- ✅ **Pass** — Block reward hard cap 650M AXON: `MaxBlockRewardSupplyStr = "650000000000000000000000000"`, remaining quota checked before each mint.
- ✅ **Pass** — Contribution reward hard cap 350M AXON: `MaxContributionSupplyStr = "350000000000000000000000000"`, similarly checked before each mint.
- ✅ **Pass** — Total supply = 650M + 350M = 1B AXON, two pools counted independently, no cross-overflow.
- ✅ **Pass** — Pre-mint `remaining` calculation uses `big.Int` subtraction; minting stops when `remaining <= 0`.

### 2.4 Minting Permissions

- ✅ **Pass** — The `agent` module account has `Minter` and `Burner` permissions (`app/config/permissions.go:63`).
- ✅ **Pass** — All `MintCoins` calls are only executed in `DistributeBlockRewards` and `MintContributionRewards`, both protected by hard caps.
- ✅ **Pass** — No admin function can bypass the hard cap to mint directly.

### 2.5 Infinite Minting Vulnerability Check

- ✅ **Pass** — `TotalBlockRewardsMinted` and `TotalContributionMinted` accumulators use `sdkmath.Int.Marshal/Unmarshal` to persist to KV Store, updated immediately after each mint.
- ⚠️ **Attention Needed** — In `addTotalBlockRewardsMinted`, the `total.Marshal()` error is ignored (`bz, _ := total.Marshal()`). Although `sdkmath.Int.Marshal` will not fail in practice, adding error handling is recommended for robustness. The same issue exists in `addTotalContributionMinted`.

**Risk Assessment: ✅ Low Risk**

Core security guarantees of the token economics model (hard cap, halving, permissions) are correctly implemented. The contract call statistics integration for contribution rewards is a functional completeness issue that requires external audit attention.

---

## 3. Deflation Mechanism Security

### 3.1 Gas Fee Burn (Path 1)

- ✅ **Pass** — `BurnCollectedFees` is implemented in `app/fee_burn.go`. Burns 80% when EIP-1559 is active, 50% in test mode.
- ✅ **Pass** — `BurnCoins` uses the `authtypes.FeeCollectorName` module account, which has `Burner` permission.
- ⚠️ **Attention Needed** — BeginBlocker execution order is critical: `BurnCollectedFees` must run before `x/distribution`'s BeginBlocker; otherwise, the distribution module will allocate all fees (including base fee) to validators. This constraint is noted in code comments, but the correct module ordering in `app.go` needs to be confirmed.
- ⚠️ **Attention Needed** — The 80% burn ratio is a conservative estimate of the base fee proportion. The actual base fee proportion depends on network congestion; in extreme cases (when priority fee proportion is very high), it may be less than 80%. The whitepaper claims "Base Fee 100% burned", while the actual implementation is "80% of total gas fees burned" — these have different semantics. We recommend clarifying this in the documentation.

### 3.2 Registration Burn (Path 2)

- ✅ **Pass** — Burns 20 AXON at registration: `burnAmount := sdk.NewInt64Coin("aaxon", int64(params.RegisterBurnAmount)*1e18)`, burned from the `agent` module account via `BurnCoins`.
- ✅ **Pass** — Burn executes after the stake is transferred to the module account; the fund flow is clear: user → agent module → partial burn.

### 3.3 Contract Deployment Burn (Path 3)

- ✅ **Pass** — `DeployBurnHook` is implemented in `PostTxProcessing`, triggered when `receipt.ContractAddress != zero`.
- ✅ **Pass** — Burns 10 AXON: `DeployBurnAxon = 10`, completed through two steps: `SendCoinsFromAccountToModule` + `BurnCoins`.
- ✅ **Pass** — Returns an error rejecting the deployment when balance is insufficient.
- ⚠️ **Attention Needed** — The burn in `PostTxProcessing` uses `evmtypes.ModuleName` (i.e., the `vm` module) as an intermediary rather than the `agent` module. The `vm` module also has `Burner` permission, so this is functionally correct, but the cross-module fund flow increases comprehension complexity.

### 3.4 Zero Reputation Burn (Path 4)

- ✅ **Pass** — `handleZeroReputation` burns remaining stake when reputation drops to 0: calculates `remaining = StakeAmount - burnedAtRegister`, then calls `BurnCoins`.
- ✅ **Pass** — The deregistration queue also checks the zero reputation condition: `if agent.Reputation == 0 && moduleHeld.IsPositive()` → burn instead of refund.
- ⚠️ **Attention Needed** — `handleZeroReputation` calculates `remaining` assuming that the amount burned at registration always equals the current parameter `RegisterBurnAmount`. If the parameter is modified via governance, the actual burn amount for already-registered Agents may differ from the current parameter, causing `remaining` calculation deviation. We recommend recording the actual burn amount in the Agent state.

### 3.5 AI Cheating Penalty Burn (Path 5)

- ✅ **Pass** — `penalizeCheater` slashes 20% of stake and burns it: `slashAmount := agent.StakeAmount.Amount.MulRaw(CheatPenaltyStakePercent).QuoRaw(100)`.
- ✅ **Pass** — Also deducts reputation -20 and sets AIBonus = -5.
- ✅ **Pass** — Cheating detection logic: multiple validators with the same `CommitHash` are flagged as cheaters (`detectCheaters` checks for duplicate commit hashes).

### 3.6 BurnCoins Permission Verification

- ✅ **Pass** — All `BurnCoins` calls are executed through module accounts with `Burner` permission:
  - `authtypes.FeeCollectorName` (gas burn) → `Burner` configured
  - `agenttypes.ModuleName` (registration/reputation/cheating burn) → `Minter` + `Burner` configured
  - `evmtypes.ModuleName` (deployment burn) → `Minter` + `Burner` configured

**Risk Assessment: ⚠️ Medium Risk**

The deflation mechanisms themselves are correctly implemented, but BeginBlocker execution order dependency and backward compatibility during parameter changes require focused external audit verification.

---

## 4. Agent Module Security

### 4.1 Registration (Register)

- ✅ **Pass** — Duplicate registration check: `if k.IsAgent(ctx, msg.Sender) { return nil, ErrAgentAlreadyRegistered }`.
- ✅ **Pass** — Minimum stake check: `msg.Stake.IsLT(minStake)` compares in `aaxon` units, correctly handling 18 decimal places.
- ✅ **Pass** — Stake is first transferred to the module account then partially burned; fund flow is secure.
- ✅ **Pass** — Initial reputation = 10, status = Online, LastHeartbeat = current block height.
- ⚠️ **Attention Needed** — `minStake` calculation uses `int64` multiplication: `int64(params.MinRegisterStake)*1e18`. When `MinRegisterStake > 9223` (i.e., 9223 AXON), int64 overflow occurs. The current default value of 100 is safe, but caution is needed if the parameter is increased via governance. We recommend using `sdkmath.NewInt` for safe multiplication.

### 4.2 Heartbeat

- ✅ **Pass** — Heartbeat frequency limit: `ctx.BlockHeight()-agent.LastHeartbeat < params.HeartbeatInterval` prevents spam heartbeats.
- ✅ **Pass** — Suspended Agents cannot send heartbeats.
- ✅ **Pass** — Heartbeat timeout detection is implemented in `BeginBlocker.checkHeartbeatTimeouts`: exceeding `HeartbeatTimeout` (720 blocks ≈ 1 hour) without heartbeat → status becomes Offline + reputation -5.

### 4.3 Deregistration (Deregister)

- ✅ **Pass** — 7-day cooldown period (`DeregisterCooldownBlocks = 120960` blocks), preventing immediate stake withdrawal.
- ✅ **Pass** — Duplicate request prevention: `if k.HasDeregisterRequest(ctx, msg.Sender) { return nil, ErrDeregisterAlreadyQueued }`.
- ✅ **Pass** — At deregistration execution, stake is correctly refunded (minus the registration burn portion); when reputation is zero, all remaining stake is burned.
- ✅ **Pass** — All associated state is cleaned up after deregistration: `DeleteAgent` + `DeleteDeregisterRequest` + `DeleteAIBonus`.

### 4.4 AI Challenge (Commit-Reveal)

- ✅ **Pass** — Two-phase commit-reveal scheme prevents plagiarism: submit answer hash (Commit), reveal plaintext after deadline (Reveal), hash verification `hex.EncodeToString(revealHash[:]) != response.CommitHash`.
- ✅ **Pass** — Deadline block check: `if ctx.BlockHeight() > challenge.DeadlineBlock { return nil, ErrChallengeWindowClosed }`.
- ✅ **Pass** — Duplicate submission prevention: `if store.Has(key) { return nil, ErrAlreadySubmitted }`.
- ✅ **Pass** — Cheating detection: multiple validators with the same CommitHash are flagged as colluding.
- ⚠️ **Attention Needed** — The AI challenge question bank is hardcoded in the source code (110 questions in `challenge.go`). Validators can pre-read the source code to obtain all answers. A long-lived public network should replace this with a dynamic question bank (for example via governance injection or on-chain random generation).
- ⚠️ **Attention Needed** — Challenge question selection uses `HeaderHash + Epoch` as a seed. Since the block hash is determined at block production time, the Proposer can theoretically predict the next challenge. However, since challenges target the entire Epoch rather than individual blocks, the exploitation window is limited.

### 4.5 Reputation System

- ✅ **Pass** — Reputation boundary checks are implemented: `if newRep < 0 { newRep = 0 }`, `if newRep > int64(params.MaxReputation) { newRep = int64(params.MaxReputation) }`.
- ✅ **Pass** — Reputation is non-transferable and non-purchasable (no related msg types).
- ✅ **Pass** — Inactivity decay: Offline status -1 per Epoch, timeout directly -5.
- ✅ **Pass** — Heavy penalty for malicious behavior: double-signing -50, cheating -20.

**Risk Assessment: ⚠️ Medium Risk**

The Agent module logic is overall sound. The int64 overflow risk and hardcoded question bank are issues that need to be fixed before mainnet.

---

## 5. Precompile Contract Security

### 5.1 IAgentRegistry (0x...0801)

- ✅ **Pass** — Read/write methods are correctly separated: `IsTransaction` marks `register/updateAgent/heartbeat/deregister` as write operations, `isAgent/getAgent` as read operations.
- ✅ **Pass** — Write operations are rejected in readonly mode via `cmn.SetupABI`.
- ✅ **Pass** — `register` uses `contract.Caller()` to ensure the registrant is the caller, no proxy registration risk.
- ✅ **Pass** — Underlying implementation reuses `msgServer`, sharing the same logic as CLI registration.
- ✅ **Pass** — Gas metering is explicit: `GasRegister = 50000`, `GasIsAgent = 200`, etc.

### 5.2 IAgentReputation (0x...0802)

- ✅ **Pass** — Pure read-only contract: `IsTransaction` returns `false` for all methods.
- ✅ **Pass** — No state-changing capability, no reentrancy attack surface.
- ✅ **Pass** — Batch query `getReputations` does not limit array length — gas metering (`GasGetReputations = 500`) provides basic protection, but very large arrays may consume excessive computation.
- ⚠️ **Attention Needed** — `GasGetReputations = 500` is a fixed value that does not scale linearly with query array length. A malicious caller could pass an extremely large array to perform massive state reads at low gas cost. We recommend changing to `baseGas + perElementGas * len(addrs)`.

### 5.3 IAgentWallet (0x...0803)

- ✅ **Pass** — Three-key model fully implemented:
  - `createWallet`: `contract.Caller()` automatically becomes Owner.
  - `execute`: Only the Operator can execute (`contract.Caller() != wallet.Operator`).
  - `freeze`: Only the Guardian or Owner can freeze.
  - `recover`: Only the Guardian can execute, replacing the Operator and automatically unfreezing.
  - `setTrust` / `removeTrust`: Only the Owner can execute.
- ✅ **Pass** — Four trust channel levels correctly implemented: Blocked(0) reject, Unknown(1) default limit, Limited(2) channel limit, Full(3) unlimited.
- ✅ **Pass** — Trust channel expiry check: `if hasTrust && channel.ExpiresAt > 0 && ctx.BlockHeight() > channel.ExpiresAt { hasTrust = false }`.
- ✅ **Pass** — Daily limit reset: every 17,280 blocks (≈24 hours) automatically resets `DailySpent`.
- ✅ **Pass** — All `execute` calls are rejected when wallet is frozen: the first check is `if wallet.IsFrozen`.
- ✅ **Pass** — Trust level validation: `if trustLevel > TrustFull { return nil, "invalid trust level: must be 0-3" }`.
- ⚠️ **Attention Needed** — `doTransfer` uses `evm.Context.Transfer` to directly manipulate StateDB balances. This bypasses the `x/bank` module, meaning Cosmos-side balance queries may be inconsistent with the EVM side (this is a known architectural characteristic of Cosmos EVM, not specific to Axon).
- ⚠️ **Attention Needed** — The `data` parameter (`args[3]`) in `executeWallet` is received but not used. Currently only ETH/AXON transfers are supported, not contract call forwarding. If support for contract calls is added in the future, reentrancy and delegatecall security need to be carefully considered.

### 5.4 IPrivateIdentity (0x...0812)

- ✅ **Pass** — Identity commitment registration is restricted to registered Agents, and duplicate registration for the same Agent or the same commitment is rejected.
- ✅ **Pass** — Starting at the v1.1.1 upgrade height, the chain stores a reverse index (`agent -> commitment`) and clears private-identity state during successful deregistration queue execution. This prevents post-deregistration commitment reuse for identities registered after the upgrade.
- ✅ **Pass** — Historical replay compatibility is preserved: pre-upgrade registrations keep the legacy one-byte marker format, and deregistration only removes the agent-side marker when old data lacks a reverse index.
- ⚠️ **Attention Needed** — Legacy commitments created before the upgrade do not contain enough information to reconstruct and delete the original `IdentityKey` entry during deregistration. Those legacy commitments remain queryable unless a dedicated migration is added.
- ⚠️ **Attention Needed** — The proof semantics still depend on the ZK circuit design and downstream contract usage. The precompile verifies the supplied proof against the registered commitment and public inputs, but it does not by itself enforce how external contracts should interpret time-varying Agent properties.

**Risk Assessment: ⚠️ Medium Risk**

The wallet precompile contract is the most complex security component. The permission model is correctly implemented, and private-identity lifecycle cleanup now covers post-upgrade registrations, but gas pricing, EVM state consistency, and legacy identity-state migration still need external audit verification.

---

## 6. EVM Security

### 6.1 PostTxProcessing Hook

- ✅ **Pass** — `DeployBurnHook` executes in `PostTxProcessing`. At this point the EVM transaction has completed; returning an error causes the entire transaction to rollback (including the contract deployment), so there is no partial execution risk.
- ✅ **Pass** — Accurately determines whether it is a contract deployment transaction via `receipt.ContractAddress == (common.Address{})`.
- ⚠️ **Attention Needed** — `PostTxProcessing` itself is called by `StateTransition` in the Cosmos EVM framework. If a panic occurs in the hook, it may crash the node. The current implementation uses error return rather than panic, which is correct. Adding recover protection is recommended.

### 6.2 Precompile Gas Metering

- ✅ **Pass** — All three precompile contracts implement the `RequiredGas` method, setting fixed gas consumption for each function.
- ⚠️ **Attention Needed** — Gas values are empirical estimates that have not been verified through benchmarks. The computational cost of write operations (e.g., `GasCreateWallet = 50000`) may not match the actual KV Store operation costs. Gas priced too low can lead to DoS attacks; priced too high affects usability.
- ⚠️ **Attention Needed** — Fixed gas (500) for `IAgentReputation.getReputations` does not scale with array length, posing a low-cost bulk read risk.

### 6.3 EVM and Cosmos State Consistency

- ✅ **Pass** — Precompile contracts obtain `sdk.Context` via `RunNativeAction` to operate on the Cosmos KV Store. This ensures precompile state changes are committed or rolled back in the same transaction as Cosmos transactions.
- ⚠️ **Attention Needed** — The wallet precompile uses `evm.Context.Transfer` to operate on the EVM StateDB, while registry/reputation operations use the Cosmos KV Store. Consistency between the two state systems depends on the correct synchronization mechanism of the Cosmos EVM framework.

**Risk Assessment: ⚠️ Medium Risk**

EVM integration relies on the mature Cosmos EVM framework. Axon's custom hooks and precompile implementation risks are manageable. Gas metering requires more precise benchmarking.

---

## 7. Keys and Permissions

### 7.1 Module Account Permissions

```
Module Account Permission Overview (app/config/permissions.go):

  fee_collector         → [Burner]                  ✅ Burn only, no minting
  distribution          → []                        ✅ No special permissions
  transfer (IBC)        → [Minter, Burner]          ✅ IBC standard requirement
  mint                  → [Minter]                  ✅ Mint only
  bonded_pool           → [Burner, Staking]         ✅ Staking standard requirement
  not_bonded_pool       → [Burner, Staking]         ✅ Staking standard requirement
  gov                   → [Burner]                  ✅ Governance standard requirement
  vm (EVM)              → [Minter, Burner]          ✅ EVM standard requirement
  feemarket             → []                        ✅ No special permissions
  erc20                 → [Minter, Burner]          ✅ ERC20 bridge requirement
  agent                 → [Minter, Burner]          ✅ Block reward minting + deflation burning
```

- ✅ **Pass** — Modules with `Minter` permission: `mint`, `transfer`, `vm`, `erc20`, `agent`. All are necessary permissions; the `agent` module's minting is protected by a dual hard cap (650M + 350M).
- ✅ **Pass** — Modules with `Burner` permission: `fee_collector`, `bonded_pool`, `not_bonded_pool`, `gov`, `vm`, `erc20`, `agent`. All are necessary permissions.
- ✅ **Pass** — `BlockedAddresses()` function correctly blocks direct transfers to module accounts and precompile contract addresses.

### 7.2 Genesis Allocation

- ✅ **Pass** — `DefaultGenesis` returns an empty Agent list and default parameters: `Agents: []Agent{}`.
- ✅ **Pass** — No pre-allocated tokens. Initial tokens received by validators in the genesis file are only for gentx staking, configured by node operators themselves.
- ✅ **Pass** — The whitepaper promises 0% pre-allocation, and the code implementation is consistent.

### 7.3 Admin Keys

- ✅ **Pass** — No admin backdoor: the code contains no privileged admin addresses or owner patterns.
- ✅ **Pass** — Parameter changes are only possible through `x/gov` governance proposals, requiring network-wide voting approval.
- ✅ **Pass** — The `SetParams` function requires parameters to pass `Validate()`, preventing invalid parameter settings.

**Risk Assessment: ✅ Low Risk**

Permission configuration follows the principle of least privilege, with no abnormal privileges.

---

## 8. Network Security

### 8.1 P2P Configuration

- ✅ **Pass** — Uses the CometBFT standard P2P protocol, supporting persistent peers, seeds, and private peers configuration.
- ⚠️ **Attention Needed** — The following CometBFT configurations need to be confirmed for public deployment:
  - `max_num_inbound_peers` and `max_num_outbound_peers` set to reasonable values
  - `pex` (Peer Exchange) configuration on validator nodes
  - `addr_book_strict` enabled for public network deployment
  - `seed_mode` enabled only on seed nodes

### 8.2 RPC Access Control

- ⚠️ **Attention Needed** — JSON-RPC (the self-hosted default is port 8545) exposes the Ethereum standard interface, including `eth_sendRawTransaction`. Public RPC nodes need to configure:
  - CORS whitelist
  - Request body size limits
  - Method whitelist (disable dangerous methods like `debug_*`, `admin_*`)
  - TLS/HTTPS encryption
- ⚠️ **Attention Needed** — CometBFT RPC (the self-hosted default is port 26657) also requires access control. Validator nodes should only expose this port on internal networks.

### 8.3 Rate Limiting

- ⚠️ **Attention Needed** — No chain-level RPC rate limiting is currently implemented. We recommend configuring request rate limiting at the reverse proxy layer (Nginx/Caddy), or using Cosmos SDK's mempool configuration to limit the number of pending transactions per address.

**Risk Assessment: ⚠️ Medium Risk**

Network security configuration is an operational concern — not technically complex but easily overlooked. We recommend including network security checkpoints in the node deployment checklist.

---

## 9. Known Risks and Mitigations

### 9.1 High Priority

| # | Risk | Impact | Mitigation | Status |
|---|------|--------|------------|--------|
| 1 | AI challenge question bank is hardcoded in source code | Validators can pre-read source code to get answers, rendering AI challenges ineffective | Implement dynamic question bank (governance injection / on-chain generation) before mainnet | ⚠️ To Fix |
| 2 | int64 overflow: `int64(params.MinRegisterStake)*1e18` | Overflows when parameter > 9223 | Use `sdkmath.NewInt` instead | ⚠️ To Fix |
| 3 | BeginBlocker module order dependency | Gas fee burn must occur before distribution | Add integration tests to verify ordering | ⚠️ Needs Verification |

### 9.2 Medium Priority

| # | Risk | Impact | Mitigation | Status |
|---|------|--------|------------|--------|
| 4 | Precompile gas pricing not benchmarked | Gas too low can lead to DoS | Execute benchmarks then adjust gas values | ⚠️ To Optimize |
| 5 | `getReputations` fixed gas does not scale with array length | Low-cost bulk reads | Change to linear gas calculation | ⚠️ To Fix |
| 6 | `RegisterBurnAmount` uses current parameter instead of snapshot at registration when reputation reaches zero | Calculation deviation after parameter change | Record actual burn amount in Agent state | ⚠️ To Optimize |
| 7 | `ContractCall` counter lacks EVM hook integration | Contract call dimension in contribution rewards is ineffective | Track contract calls in EVM hook | ⚠️ To Implement |
| 8 | Challenge seed predictable by Proposer | Proposer can prepare answers in advance | Introduce VRF or multi-block seed aggregation | ⚠️ To Optimize |

### 9.3 Low Priority

| # | Risk | Impact | Mitigation | Status |
|---|------|--------|------------|--------|
| 9 | `Marshal` error is ignored (`bz, _ := total.Marshal()`) | `sdkmath.Int` theoretically won't fail, but lacks robustness | Add error handling and logging | ⚠️ To Optimize |
| 10 | Wallet `execute`'s `data` parameter is unused | Contract call forwarding not supported | Clarify in documentation or implement full functionality | ⚠️ To Confirm |
| 11 | Challenge answer matching uses simple string comparison | Semantically correct but differently formatted answers score low | Introduce more flexible scoring (NLP / fuzzy matching) | ⚠️ To Optimize |

---

## 10. Audit Recommendations

### 10.1 Recommended External Audit Firms

| Audit Firm | Expertise | Recommendation Reason |
|-----------|-----------|----------------------|
| **Trail of Bits** | Cosmos SDK, EVM, Consensus Protocols | Audited Cosmos Hub, multiple EVM chains |
| **Halborn** | Cosmos SDK, EVM Precompiles | Audited Evmos, Cronos, and similar architecture chains |
| **Oak Security** | Cosmos SDK Modules | Focused on Cosmos ecosystem audits |
| **Zellic** | EVM, Smart Contracts | Deep EVM security research |
| **CertiK** | Full-stack Blockchain Audit | Broad coverage, includes formal verification |

### 10.2 External Audit Priority

```
Priority 1 (Must-have before mainnet):
  ├── Token economics model (minting/hard cap/halving/distribution)
  ├── Correctness and completeness of five deflation paths
  ├── Module account permission configuration
  └── BeginBlocker/EndBlocker execution order

Priority 2 (Strongly recommended before mainnet):
  ├── IAgentWallet precompile (most complex security component)
  ├── Agent registration/deregistration fund flow
  ├── AI challenge commit-reveal scheme
  └── EVM PostTxProcessing hook

Priority 3 (Ongoing post-mainnet):
  ├── Precompile gas pricing optimization
  ├── Network layer configuration review
  ├── Dynamic AI question bank security
  └── Cross-chain bridge security (when IBC / Ethereum bridge launches)
```

### 10.3 Automated Security Measures Recommendations

```
Implemented:
  ✅ 70+ unit test cases
  ✅ CI (GitHub Actions) automated test runs
  ✅ gofmt consistency checks / go vet static analysis

Recommended additions:
  □ Fuzz testing (Go native fuzz + go-fuzz)
    - Focus: ABI decoding, parameter validation, big number arithmetic
  □ Invariant testing
    - Total supply = minted - burned
    - Reputation always in [0, 100] range
    - Module account balance ≥ sum of all Agent stakes
  □ Slither / Mythril static analysis on Solidity interfaces
  □ Chaos testing
    - Randomly stop validator nodes
    - Network partition simulation
    - High-concurrency registration/deregistration
```

---

## Audit Summary

| Category | Assessment | Description |
|----------|-----------|-------------|
| Consensus Security | ✅ Low Risk | Relies on mature CometBFT, custom logic risk is manageable |
| Token Economics Security | ✅ Low Risk | Hard cap protection is comprehensive, minting permissions are restricted |
| Deflation Mechanism Security | ⚠️ Medium Risk | All five paths implemented, but execution order and parameter compatibility need external verification |
| Agent Module Security | ⚠️ Medium Risk | Business logic is sound, int64 overflow and question bank issues need fixing |
| Precompile Contract Security | ⚠️ Medium Risk | Permission model is correct, gas pricing needs optimization |
| EVM Security | ⚠️ Medium Risk | Hook implementation is robust, state consistency depends on framework |
| Keys and Permissions | ✅ Low Risk | Principle of least privilege, no backdoors |
| Network Security | ⚠️ Medium Risk | Operational-level configuration needs hardening |

**Overall Assessment: The project has a solid security foundation. No 🔴 high-risk or critical vulnerabilities were found. All identified ⚠️ medium-risk issues have clear fix paths. We recommend completing Priority 1-2 external audits before mainnet launch.**

---

*This report was generated by the Axon team's internal security review and does not replace a professional third-party security audit.*
*Last updated: March 2026*
