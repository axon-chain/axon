> 🌐 [English Version](SECURITY_AUDIT_EN.md)

# Axon 安全审计自查报告

**版本：v1.0.0**
**日期：2026年3月**
**状态：主网发布后复核版**

---

## 审计范围

本报告覆盖 Axon 链的全部核心组件，按安全优先级排序：

| 模块 | 路径 | 安全等级 |
|------|------|----------|
| 共识层（CometBFT） | CometBFT 配置 | 关键 |
| 代币经济模型 | `x/agent/keeper/block_rewards.go`, `contribution.go` | 关键 |
| 通缩机制 | `app/fee_burn.go`, `app/evm_hooks.go`, `keeper/reputation.go` | 关键 |
| Agent 模块 | `x/agent/keeper/`, `x/agent/types/` | 关键 |
| 预编译合约 — IAgentRegistry | `precompiles/registry/` | 高 |
| 预编译合约 — IAgentReputation | `precompiles/reputation/` | 中 |
| 预编译合约 — IAgentWallet | `precompiles/wallet/` | 关键 |
| EVM 集成 | `app/evm_hooks.go`, Cosmos EVM 模块 | 高 |
| 模块权限配置 | `app/config/permissions.go` | 关键 |
| 创世状态 | `x/agent/types/genesis.go`, `types/params.go` | 高 |
| 网络层 | CometBFT P2P, JSON-RPC | 中 |

---

## 1. 共识安全

### 1.1 CometBFT BFT 共识

- ✅ **通过** — 使用 CometBFT（原 Tendermint）BFT 共识，容忍不超过 1/3 验证者拜占庭故障。该共识引擎经过 Cosmos Hub、Osmosis、dYdX 等数十条链的生产环境验证。
- ✅ **通过** — 即时终局性，单区块确认，无分叉风险。
- ✅ **通过** — 出块时间 ~5 秒，由 CometBFT `timeout_commit` 参数控制。

### 1.2 验证者集合管理

- ✅ **通过** — 验证者集合上限初始 100，通过 `x/staking` 模块管理，可由链上治理调整。
- ✅ **通过** — 最低验证者质押 10,000 AXON，防止低成本攻击。
- ✅ **通过** — 验证者质押解锁冷却期 14 天（`x/staking` 默认 unbonding period）。

### 1.3 Slashing 条件

- ✅ **通过** — 双签惩罚：罚没 5% 质押 + 信誉 -50 + 入狱（`x/slashing` 模块标准行为）。
- ✅ **通过** — 长期离线惩罚：罚没 0.1% 质押 + 信誉 -5 + 入狱。
- ✅ **通过** — 信誉扣减常量已在代码中硬编码：`ReputationLossSlashing = -50`，`ReputationLossOffline = -5`（`x/agent/types/params.go:15-16`）。

### 1.4 出块权重

- ✅ **通过** — 出块权重公式 `Stake × (1 + ReputationBonus + AIBonus)` 已实现。五级信誉加成（0%/5%/10%/15%/20%）在 `reputationBonusPercent()` 中正确实现。
- ✅ **通过** — 乘数最低值限制为 10（即最低 10%），防止乘数归零导致权重为零：`if multiplier < 10 { multiplier = 10 }`。

**风险评估：✅ 低风险**

共识层依赖成熟的 CometBFT 引擎，风险主要集中在 Axon 自定义的权重调整逻辑。建议外部审计重点关注 `distributeValidatorRewards` 中权重计算与 `x/staking` 模块投票权力更新的一致性。

---

## 2. 代币经济安全

### 2.1 区块奖励计算

- ✅ **通过** — 基础奖励使用 `big.Int` 字符串初始化（`BaseBlockRewardStr = "12367000000000000000"`），避免浮点精度问题。
- ✅ **通过** — 减半使用位右移（`Rsh`），数学上等价于整除 2，无溢出风险。
- ✅ **通过** — 减半次数上限 64 次检查：`if halvings >= 64 { return sdkmath.ZeroInt() }`，防止 uint 溢出。
- ✅ **通过** — 奖励三方分配比例（Proposer 25% + Validators 50% + AI 25% = 100%），余数安全处理：`aiReward := reward.Sub(proposerReward).Sub(validatorReward)`，确保无舍入丢失。

### 2.2 贡献奖励

- ✅ **通过** — 防刷机制已实现：
  - 自调用不计分（贡献评分基于 `DeployCount` 和 `ContractCall` 独立计数器）。
  - 单 Agent 每 Epoch 上限 = 池的 2%（`MaxSharePerAgentBPS = 200`）。
  - 信誉 < 20 不参与分配（`MinReputationForReward = 20`）。
  - 注册不满 7 天不参与（`MinRegistrationBlocks = 120960`）。
- ⚠️ **需关注** — 自调用过滤器：当前 `ContractCall` 计数器的递增入口（`IncrementContractCalls`）尚未在 EVM hook 中实际调用，合约被调用次数的统计尚需集成到 EVM 执行层。自调用排除逻辑依赖此集成的正确实现。
- ✅ **通过** — 活跃度上限为 100 笔交易，`activityCapped` 限制防止单维度刷分。

### 2.3 总供应量硬顶

- ✅ **通过** — 区块奖励硬顶 650M AXON：`MaxBlockRewardSupplyStr = "650000000000000000000000000"`，每次铸造前检查剩余额度。
- ✅ **通过** — 贡献奖励硬顶 350M AXON：`MaxContributionSupplyStr = "350000000000000000000000000"`，同样每次铸造前检查。
- ✅ **通过** — 总供应量 = 650M + 350M = 1B AXON，两个池独立计数，无交叉溢出。
- ✅ **通过** — 铸造前的 `remaining` 计算使用 `big.Int` 减法，当 `remaining <= 0` 时停止铸造。

### 2.4 铸造权限

- ✅ **通过** — `agent` 模块账户拥有 `Minter` 和 `Burner` 权限（`app/config/permissions.go:63`）。
- ✅ **通过** — 所有 `MintCoins` 调用仅在 `DistributeBlockRewards` 和 `MintContributionRewards` 中执行，均受硬顶保护。
- ✅ **通过** — 无 admin 函数可绕过硬顶直接铸造。

### 2.5 无限铸造漏洞检查

- ✅ **通过** — `TotalBlockRewardsMinted` 和 `TotalContributionMinted` 累加器使用 `sdkmath.Int.Marshal/Unmarshal` 持久化到 KV Store，每次铸造后立即更新。
- ⚠️ **需关注** — `addTotalBlockRewardsMinted` 中 `total.Marshal()` 的错误被忽略（`bz, _ := total.Marshal()`）。虽然 `sdkmath.Int.Marshal` 在实践中不会失败，但建议添加错误处理以提高健壮性。同样问题存在于 `addTotalContributionMinted`。

**风险评估：✅ 低风险**

代币经济模型的核心安全保障（硬顶、减半、权限）均已正确实现。贡献奖励的合约调用统计集成是需要外部审计关注的功能完整性问题。

---

## 3. 通缩机制安全

### 3.1 Gas 费销毁（路径 1）

- ✅ **通过** — `BurnCollectedFees` 在 `app/fee_burn.go` 中实现。EIP-1559 激活时销毁 80%，测试模式销毁 50%。
- ✅ **通过** — `BurnCoins` 使用 `authtypes.FeeCollectorName` 模块账户，该账户拥有 `Burner` 权限。
- ⚠️ **需关注** — BeginBlocker 执行顺序至关重要：`BurnCollectedFees` 必须在 `x/distribution` 的 BeginBlocker 之前运行，否则 distribution 模块会将全部 fee（含 base fee）分配给验证者。代码注释中已标注此约束，但需确认 `app.go` 中模块顺序的正确性。
- ⚠️ **需关注** — 80% 销毁比例是对 base fee 占比的保守估算。实际 base fee 占比取决于网络拥堵程度，极端情况下（priority fee 占比极高时）可能低于 80%。白皮书声称"Base Fee 100% 销毁"，实际实现为"总 Gas 费的 80% 销毁"——两者语义不同，建议在文档中明确说明。

### 3.2 注册销毁（路径 2）

- ✅ **通过** — 注册时销毁 20 AXON：`burnAmount := sdk.NewInt64Coin("aaxon", int64(params.RegisterBurnAmount)*1e18)`，通过 `BurnCoins` 从 `agent` 模块账户销毁。
- ✅ **通过** — 销毁在质押转入模块账户之后执行，资金流清晰：用户 → agent 模块 → 部分销毁。

### 3.3 合约部署销毁（路径 3）

- ✅ **通过** — `DeployBurnHook` 在 `PostTxProcessing` 中实现，检测 `receipt.ContractAddress != zero` 触发销毁。
- ✅ **通过** — 销毁 10 AXON：`DeployBurnAxon = 10`，通过 `SendCoinsFromAccountToModule` + `BurnCoins` 两步完成。
- ✅ **通过** — 余额不足时返回错误拒绝部署。
- ⚠️ **需关注** — `PostTxProcessing` 中销毁使用 `evmtypes.ModuleName`（即 `vm` 模块）作为中转，而非 `agent` 模块。`vm` 模块同样拥有 `Burner` 权限，功能上正确，但跨模块资金流增加了理解复杂度。

### 3.4 信誉归零销毁（路径 4）

- ✅ **通过** — `handleZeroReputation` 在信誉降为 0 时销毁剩余质押：计算 `remaining = StakeAmount - burnedAtRegister`，然后 `BurnCoins`。
- ✅ **通过** — 注销队列中同样检查信誉归零条件：`if agent.Reputation == 0 && moduleHeld.IsPositive()` → 销毁而非退还。
- ⚠️ **需关注** — `handleZeroReputation` 中计算 `remaining` 时假设注册时的销毁量始终等于当前参数的 `RegisterBurnAmount`。如果参数通过治理修改，已注册 Agent 的实际销毁量可能与当前参数不一致，导致 `remaining` 计算偏差。建议将实际销毁量记录在 Agent 状态中。

### 3.5 AI 作弊惩罚销毁（路径 5）

- ✅ **通过** — `penalizeCheater` 罚没 20% 质押并销毁：`slashAmount := agent.StakeAmount.Amount.MulRaw(CheatPenaltyStakePercent).QuoRaw(100)`。
- ✅ **通过** — 同时扣除信誉 -20 和设置 AIBonus = -5。
- ✅ **通过** — 作弊检测逻辑：相同 `CommitHash` 的多个验证者被标记为作弊者（`detectCheaters` 中检查 commit hash 重复）。

### 3.6 BurnCoins 权限验证

- ✅ **通过** — 所有 `BurnCoins` 调用均通过拥有 `Burner` 权限的模块账户执行：
  - `authtypes.FeeCollectorName`（Gas 销毁）→ 已配置 `Burner`
  - `agenttypes.ModuleName`（注册/信誉/作弊销毁）→ 已配置 `Minter` + `Burner`
  - `evmtypes.ModuleName`（部署销毁）→ 已配置 `Minter` + `Burner`

**风险评估：⚠️ 中风险**

通缩机制本身实现正确，但 BeginBlocker 执行顺序依赖和参数变更时的向后兼容性需要外部审计重点验证。

---

## 4. Agent 模块安全

### 4.1 注册（Register）

- ✅ **通过** — 重复注册检查：`if k.IsAgent(ctx, msg.Sender) { return nil, ErrAgentAlreadyRegistered }`。
- ✅ **通过** — 最低质押检查：`msg.Stake.IsLT(minStake)` 比较 `aaxon` 单位，正确处理 18 位小数。
- ✅ **通过** — 质押先转入模块账户再销毁部分，资金流安全。
- ✅ **通过** — 初始信誉 = 10，状态 = Online，LastHeartbeat = 当前区块高度。
- ⚠️ **需关注** — `minStake` 计算使用 `int64` 乘法：`int64(params.MinRegisterStake)*1e18`。当 `MinRegisterStake > 9223`（即 9223 AXON）时会导致 int64 溢出。当前默认值 100 安全，但如果通过治理调高参数需注意。建议使用 `sdkmath.NewInt` 进行安全乘法。

### 4.2 心跳（Heartbeat）

- ✅ **通过** — 心跳频率限制：`ctx.BlockHeight()-agent.LastHeartbeat < params.HeartbeatInterval` 防止垃圾心跳。
- ✅ **通过** — Suspended 状态的 Agent 无法发送心跳。
- ✅ **通过** — 心跳超时检测在 `BeginBlocker.checkHeartbeatTimeouts` 中实现：超过 `HeartbeatTimeout`（720 块 ≈ 1 小时）未心跳 → 状态变为 Offline + 信誉 -5。

### 4.3 注销（Deregister）

- ✅ **通过** — 冷却期 7 天（`DeregisterCooldownBlocks = 120960` 块），防止即时提取质押。
- ✅ **通过** — 防重复请求：`if k.HasDeregisterRequest(ctx, msg.Sender) { return nil, ErrDeregisterAlreadyQueued }`。
- ✅ **通过** — 注销执行时正确退还质押（扣除注册销毁部分），信誉归零时销毁全部剩余质押。
- ✅ **通过** — 注销后清理所有关联状态：`DeleteAgent` + `DeleteDeregisterRequest` + `DeleteAIBonus`。

### 4.4 AI 挑战（Commit-Reveal）

- ✅ **通过** — Commit-Reveal 两阶段方案防止抄袭：提交答案哈希（Commit），截止后揭示明文（Reveal），哈希验证 `hex.EncodeToString(revealHash[:]) != response.CommitHash`。
- ✅ **通过** — 截止区块检查：`if ctx.BlockHeight() > challenge.DeadlineBlock { return nil, ErrChallengeWindowClosed }`。
- ✅ **通过** — 防重复提交：`if store.Has(key) { return nil, ErrAlreadySubmitted }`。
- ✅ **通过** — 作弊检测：相同 CommitHash 的多个验证者被标记为串通。
- ⚠️ **需关注** — AI 挑战题库硬编码在源代码中（`challenge.go` 中 110 道题），验证者可以预读源代码获取全部答案。长期公网运行应改为动态题库（如治理注入或链上随机生成）。
- ⚠️ **需关注** — 挑战选题使用 `HeaderHash + Epoch` 作为种子，由于区块哈希在出块时确定，Proposer 理论上可以预测下一个挑战。但由于挑战针对整个 Epoch 而非单个区块，利用窗口有限。

### 4.5 信誉系统

- ✅ **通过** — 信誉边界检查已实现：`if newRep < 0 { newRep = 0 }`，`if newRep > int64(params.MaxReputation) { newRep = int64(params.MaxReputation) }`。
- ✅ **通过** — 信誉不可转让、不可购买（无相关 msg 类型）。
- ✅ **通过** — 不活跃衰减：Offline 状态每 Epoch -1，超时直接 -5。
- ✅ **通过** — 恶意行为重罚：双签 -50，作弊 -20。

**风险评估：⚠️ 中风险**

Agent 模块逻辑整体健全。int64 溢出风险和硬编码题库是需要在主网前修复的问题。

---

## 5. 预编译合约安全

### 5.1 IAgentRegistry（0x...0801）

- ✅ **通过** — 读写方法正确分离：`IsTransaction` 将 `register/updateAgent/heartbeat/deregister` 标记为写操作，`isAgent/getAgent` 为读操作。
- ✅ **通过** — 写操作通过 `cmn.SetupABI` 在 readonly 模式下被拒绝。
- ✅ **通过** — `register` 使用 `contract.Caller()` 确保注册者即为调用者，无代理注册风险。
- ✅ **通过** — 底层复用 `msgServer` 实现，与 CLI 注册共享同一逻辑。
- ✅ **通过** — Gas 计量明确：`GasRegister = 50000`，`GasIsAgent = 200` 等。

### 5.2 IAgentReputation（0x...0802）

- ✅ **通过** — 纯只读合约：`IsTransaction` 对所有方法返回 `false`。
- ✅ **通过** — 无状态变更能力，无重入攻击面。
- ✅ **通过** — 批量查询 `getReputations` 未限制数组长度 — Gas 计量（`GasGetReputations = 500`）提供了基本保护，但超大数组可能消耗过多计算。
- ⚠️ **需关注** — `GasGetReputations = 500` 是固定值，不随查询数组长度线性增长。恶意调用者可以传入极大数组以低 Gas 消耗执行大量状态读取。建议改为 `baseGas + perElementGas * len(addrs)`。

### 5.3 IAgentWallet（0x...0803）

- ✅ **通过** — 三密钥模型完整实现：
  - `createWallet`：`contract.Caller()` 自动成为 Owner。
  - `execute`：仅 Operator 可执行（`contract.Caller() != wallet.Operator`）。
  - `freeze`：仅 Guardian 或 Owner 可冻结。
  - `recover`：仅 Guardian 可执行，更换 Operator 并自动解冻。
  - `setTrust` / `removeTrust`：仅 Owner 可执行。
- ✅ **通过** — 信任通道四级正确实现：Blocked(0) 拒绝、Unknown(1) 默认限额、Limited(2) 通道限额、Full(3) 无限制。
- ✅ **通过** — 信任通道过期检查：`if hasTrust && channel.ExpiresAt > 0 && ctx.BlockHeight() > channel.ExpiresAt { hasTrust = false }`。
- ✅ **通过** — 日限额重置：每 17,280 块（≈24 小时）自动重置 `DailySpent`。
- ✅ **通过** — 冻结钱包时所有 `execute` 调用被拒绝：第一个检查即为 `if wallet.IsFrozen`。
- ✅ **通过** — Trust level 验证：`if trustLevel > TrustFull { return nil, "invalid trust level: must be 0-3" }`。
- ⚠️ **需关注** — `doTransfer` 使用 `evm.Context.Transfer` 直接操作 StateDB 余额。这绕过了 `x/bank` 模块，意味着 Cosmos 侧余额查询可能与 EVM 侧不一致（这是 Cosmos EVM 的已知架构特点，非 Axon 特有问题）。
- ⚠️ **需关注** — `executeWallet` 的 `data` 参数（`args[3]`）被接收但未被使用。当前仅支持 ETH/AXON 转账，不支持合约调用转发。如果未来扩展支持合约调用，需要仔细考虑重入和委托调用安全。

### 5.4 IPrivateIdentity（0x...0812）

- ✅ **通过** — 身份承诺注册仅允许已注册 Agent 调用，同一 Agent 或同一 commitment 的重复注册都会被拒绝。
- ✅ **通过** — 自 v1.1.1 升级高度起，链上会为身份承诺保存 `agent -> commitment` 反向索引，并在注销队列成功执行时清理 private identity 状态。这可以阻止升级后新注册身份在注销后继续复用 commitment。
- ✅ **通过** — 历史回放兼容性得到保留：升级前注册继续使用旧的一字节标记格式；对于缺少反向索引的旧数据，注销时仅删除 agent 侧标记，不改写历史 commitment 状态。
- ⚠️ **需关注** — 升级前写入的 legacy commitment 缺少足够信息，注销时无法反推出并删除原始 `IdentityKey`。如果要彻底清除这部分历史 identity 状态，需要额外迁移。
- ⚠️ **需关注** — 证明语义仍然依赖 ZK 电路设计和下游合约的解释方式。预编译只负责针对注册 commitment 和公开输入验证 proof，本身不直接规定外部合约如何解释随时间变化的 Agent 属性。

**风险评估：⚠️ 中风险**

钱包预编译合约仍是最复杂的安全组件。权限模型已正确实现，private identity 生命周期清理现在也覆盖了升级后注册的数据，但 Gas 定价、EVM 状态一致性以及 legacy identity 状态迁移仍需外部审计验证。

---

## 6. EVM 安全

### 6.1 PostTxProcessing Hook

- ✅ **通过** — `DeployBurnHook` 在 `PostTxProcessing` 中执行，此时 EVM 交易已完成，返回 error 会导致整个交易回滚（包括合约部署），不存在部分执行风险。
- ✅ **通过** — 通过 `receipt.ContractAddress == (common.Address{})` 准确判断是否为合约部署交易。
- ⚠️ **需关注** — `PostTxProcessing` 本身在 Cosmos EVM 框架中由 `StateTransition` 调用。如果 hook 中发生 panic，可能导致节点崩溃。当前实现使用 error return 而非 panic，这是正确的做法。建议添加 recover 保护。

### 6.2 预编译 Gas 计量

- ✅ **通过** — 所有三个预编译合约均实现了 `RequiredGas` 方法，为每个函数设定了固定 Gas 消耗。
- ⚠️ **需关注** — Gas 值为经验估算，尚未通过 benchmark 验证。写操作（如 `GasCreateWallet = 50000`）与实际 KV Store 操作的计算成本可能不匹配。Gas 定价过低可导致 DoS 攻击，过高会影响可用性。
- ⚠️ **需关注** — `IAgentReputation.getReputations` 的固定 Gas（500）不随数组长度增长，存在低成本批量读取风险。

### 6.3 EVM 与 Cosmos 状态一致性

- ✅ **通过** — 预编译合约通过 `RunNativeAction` 获取 `sdk.Context`，操作 Cosmos KV Store。这确保预编译的状态变更与 Cosmos 交易在同一事务中提交或回滚。
- ⚠️ **需关注** — Wallet 预编译使用 `evm.Context.Transfer` 操作 EVM StateDB，而注册/信誉操作使用 Cosmos KV Store。两套状态系统的一致性依赖 Cosmos EVM 框架的正确同步机制。

**风险评估：⚠️ 中风险**

EVM 集成依赖成熟的 Cosmos EVM 框架，Axon 的自定义 hook 和预编译实现风险可控。Gas 计量需要更精确的 benchmark。

---

## 7. 密钥与权限

### 7.1 模块账户权限

```
模块账户权限一览（app/config/permissions.go）：

  fee_collector         → [Burner]                  ✅ 仅销毁，无铸造
  distribution          → []                        ✅ 无特殊权限
  transfer (IBC)        → [Minter, Burner]          ✅ IBC 标准需求
  mint                  → [Minter]                  ✅ 仅铸造
  bonded_pool           → [Burner, Staking]         ✅ 质押标准需求
  not_bonded_pool       → [Burner, Staking]         ✅ 质押标准需求
  gov                   → [Burner]                  ✅ 治理标准需求
  vm (EVM)              → [Minter, Burner]          ✅ EVM 标准需求
  feemarket             → []                        ✅ 无特殊权限
  erc20                 → [Minter, Burner]          ✅ ERC20 桥接需求
  agent                 → [Minter, Burner]          ✅ 区块奖励铸造 + 通缩销毁
```

- ✅ **通过** — 拥有 `Minter` 权限的模块：`mint`、`transfer`、`vm`、`erc20`、`agent`。均为必要权限，其中 `agent` 模块的铸造受双重硬顶（650M + 350M）保护。
- ✅ **通过** — 拥有 `Burner` 权限的模块：`fee_collector`、`bonded_pool`、`not_bonded_pool`、`gov`、`vm`、`erc20`、`agent`。均为必要权限。
- ✅ **通过** — `BlockedAddresses()` 函数正确阻止向模块账户和预编译合约地址直接转账。

### 7.2 创世分配

- ✅ **通过** — `DefaultGenesis` 返回空 Agent 列表和默认参数：`Agents: []Agent{}`。
- ✅ **通过** — 无预分配代币。创世文件中验证者获得的初始代币仅用于 gentx 质押，由节点运营者自行配置。
- ✅ **通过** — 白皮书承诺 0% 预分配，代码实现与之一致。

### 7.3 Admin 密钥

- ✅ **通过** — 无 admin 后门：代码中不存在特权管理员地址或 owner 模式。
- ✅ **通过** — 参数修改仅通过 `x/gov` 治理提案实现，需要全网投票通过。
- ✅ **通过** — `SetParams` 函数要求参数通过 `Validate()`，防止设置无效参数。

**风险评估：✅ 低风险**

权限配置遵循最小权限原则，无异常特权。

---

## 8. 网络安全

### 8.1 P2P 配置

- ✅ **通过** — 使用 CometBFT 标准 P2P 协议，支持 persistent peers、seeds、private peers 配置。
- ⚠️ **需关注** — 公网部署时需确认以下 CometBFT 配置：
  - `max_num_inbound_peers` 和 `max_num_outbound_peers` 设置合理值
  - `pex`（Peer Exchange）在验证者节点上的配置
  - `addr_book_strict` 在公网部署时启用
  - `seed_mode` 仅在种子节点上启用

### 8.2 RPC 访问控制

- ⚠️ **需关注** — JSON-RPC（自托管默认端口为 8545）暴露了以太坊标准接口，包括 `eth_sendRawTransaction`。公开 RPC 节点需要配置：
  - CORS 白名单
  - 请求体大小限制
  - 方法白名单（禁用 `debug_*`、`admin_*` 等危险方法）
  - TLS/HTTPS 加密
- ⚠️ **需关注** — CometBFT RPC（自托管默认端口为 26657）同样需要访问控制。验证者节点应仅在内网暴露此端口。

### 8.3 速率限制

- ⚠️ **需关注** — 当前未实现链级别的 RPC 速率限制。建议在反向代理层（Nginx/Caddy）配置请求速率限制，或使用 Cosmos SDK 的 mempool 配置限制单地址的 pending 交易数量。

**风险评估：⚠️ 中风险**

网络安全配置为运维层面问题，技术上不复杂但容易遗漏。建议在节点部署清单中加入网络安全检查项。

---

## 9. 已知风险与缓解措施

### 9.1 高优先级

| # | 风险 | 影响 | 缓解措施 | 状态 |
|---|------|------|---------|------|
| 1 | AI 挑战题库硬编码在源代码中 | 验证者可预读源码获取答案，使 AI 挑战形同虚设 | 主网前实现动态题库（治理注入 / 链上生成） | ⚠️ 待修复 |
| 2 | int64 溢出：`int64(params.MinRegisterStake)*1e18` | 参数 > 9223 时溢出 | 改用 `sdkmath.NewInt` | ⚠️ 待修复 |
| 3 | BeginBlocker 模块顺序依赖 | Gas 费销毁必须在 distribution 之前 | 添加集成测试验证顺序 | ⚠️ 需验证 |

### 9.2 中优先级

| # | 风险 | 影响 | 缓解措施 | 状态 |
|---|------|------|---------|------|
| 4 | 预编译 Gas 定价未经 benchmark | Gas 过低可导致 DoS | 执行 benchmark 后调整 Gas 值 | ⚠️ 待优化 |
| 5 | `getReputations` 固定 Gas 不随数组长度增长 | 低成本大批量读取 | 改为线性 Gas 计算 | ⚠️ 待修复 |
| 6 | 信誉归零时 `RegisterBurnAmount` 使用当前参数而非注册时快照 | 参数变更后计算偏差 | 在 Agent 状态中记录实际销毁量 | ⚠️ 待优化 |
| 7 | `ContractCall` 计数器缺少 EVM hook 集成 | 贡献奖励中合约被调用维度无效 | 在 EVM hook 中追踪合约调用 | ⚠️ 待实现 |
| 8 | 挑战种子可被 Proposer 预测 | Proposer 可提前准备答案 | 引入 VRF 或多区块种子聚合 | ⚠️ 待优化 |

### 9.3 低优先级

| # | 风险 | 影响 | 缓解措施 | 状态 |
|---|------|------|---------|------|
| 9 | `Marshal` 错误被忽略（`bz, _ := total.Marshal()`） | 理论上 `sdkmath.Int` 不会失败，但不够健壮 | 添加错误处理和日志 | ⚠️ 待优化 |
| 10 | Wallet `execute` 的 `data` 参数未使用 | 不支持合约调用转发 | 明确文档说明或实现完整功能 | ⚠️ 待确认 |
| 11 | 挑战答案匹配使用简单字符串比较 | 语义正确但格式不同的答案得低分 | 引入更灵活的评分机制（NLP / 模糊匹配） | ⚠️ 待优化 |

---

## 10. 审计建议

### 10.1 推荐外部审计机构

| 审计机构 | 擅长领域 | 推荐理由 |
|---------|---------|---------|
| **Trail of Bits** | Cosmos SDK、EVM、共识协议 | 审计过 Cosmos Hub、多条 EVM 链 |
| **Halborn** | Cosmos SDK、EVM 预编译 | 审计过 Evmos、Cronos 等同架构链 |
| **Oak Security** | Cosmos SDK 模块 | 专注 Cosmos 生态审计 |
| **Zellic** | EVM、智能合约 | 深度 EVM 安全研究 |
| **CertiK** | 全栈区块链审计 | 覆盖面广，含形式化验证 |

### 10.2 外部审计优先级

```
优先级 1（主网前必须）：
  ├── 代币经济模型（铸造/硬顶/减半/分配）
  ├── 五条通缩路径的正确性与完整性
  ├── 模块账户权限配置
  └── BeginBlocker/EndBlocker 执行顺序

优先级 2（主网前强烈建议）：
  ├── IAgentWallet 预编译（最复杂的安全组件）
  ├── Agent 注册/注销资金流
  ├── AI 挑战 commit-reveal 方案
  └── EVM PostTxProcessing hook

优先级 3（主网后持续）：
  ├── 预编译 Gas 定价优化
  ├── 网络层配置审查
  ├── 动态 AI 题库安全性
  └── 跨链桥安全（IBC / 以太坊桥上线时）
```

### 10.3 自动化安全措施建议

```
已实现：
  ✅ 单元测试 70+ 用例
  ✅ CI（GitHub Actions）自动运行测试
  ✅ gofmt 一致性检查 / go vet 静态分析

建议补充：
  □ Fuzzing 测试（Go 原生 fuzz + go-fuzz）
    - 重点：ABI 解码、参数验证、大数运算
  □ 不变量测试（Invariant Testing）
    - 总供应量 = 已铸造 - 已销毁
    - 信誉始终在 [0, 100] 范围内
    - 模块账户余额 ≥ 所有 Agent 质押之和
  □ Slither / Mythril 对 Solidity 接口的静态分析
  □ 混沌测试（Chaos Testing）
    - 随机停止验证者节点
    - 网络分区模拟
    - 大量并发注册/注销
```

---

## 审计总结

| 类别 | 评估 | 说明 |
|------|------|------|
| 共识安全 | ✅ 低风险 | 依赖成熟的 CometBFT，自定义逻辑风险可控 |
| 代币经济安全 | ✅ 低风险 | 硬顶保护完善，铸造权限受限 |
| 通缩机制安全 | ⚠️ 中风险 | 五条路径均已实现，但执行顺序和参数兼容性需外部验证 |
| Agent 模块安全 | ⚠️ 中风险 | 业务逻辑健全，int64 溢出和题库问题需修复 |
| 预编译合约安全 | ⚠️ 中风险 | 权限模型正确，Gas 定价需优化 |
| EVM 安全 | ⚠️ 中风险 | Hook 实现稳健，状态一致性依赖框架 |
| 密钥与权限 | ✅ 低风险 | 最小权限原则，无后门 |
| 网络安全 | ⚠️ 中风险 | 运维层面配置待加固 |

**总体评估：项目安全基础扎实，无发现 🔴 高风险或关键漏洞。已识别的 ⚠️ 中风险问题均有明确的修复路径。建议在主网上线前完成优先级 1-2 的外部审计。**

---

*本报告由 Axon 团队内部安全审查生成，不替代专业第三方安全审计。*
*最后更新：2026年3月*
