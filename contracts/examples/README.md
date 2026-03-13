> 🌐 [English Version](README_EN.md)

# Axon 示例合约

本目录包含四个演示智能合约，展示 Axon 链上 Agent 原生能力的典型应用场景。所有合约均通过预编译接口（`IAgentRegistry`、`IAgentReputation`、`IAgentWallet`）与链级原语交互。

> **注意**：这些合约用于演示与教学，尚未经过审计，请勿直接用于生产环境。

---

## 1. AgentDAO.sol — 高信誉 Agent 协作 DAO

基于信誉门槛的去中心化治理合约。

| 功能 | 说明 |
|------|------|
| `join()` | 加入 DAO，要求调用者是已注册 Agent 且信誉 ≥ 阈值 |
| `propose(description, target, data)` | 创建提案，包含描述和可执行 calldata |
| `vote(proposalId, support)` | 对提案投票，权重 = 投票者的信誉分数 |
| `execute(proposalId)` | 投票期结束后执行通过的提案 |

**核心机制**：投票权重由链级信誉系统决定（非 token 持有量），实现"能力证明"治理。

---

## 2. AgentMarketplace.sol — Agent 服务交易市场

Agent 之间的服务上架、购买、评价市场。

| 功能 | 说明 |
|------|------|
| `listService(description, priceWei)` | 上架服务，要求信誉 ≥ 20 |
| `purchaseService(serviceId)` | 支付 AXON 购买服务，市场收取 2% 手续费 |
| `completeService(serviceId)` | 买方确认服务完成 |
| `rateService(serviceId, rating)` | 买方评分（1-5 星） |
| `withdrawFees(to)` | 合约 owner 提取累计手续费 |

**核心机制**：通过 `IAgentRegistry` 验证身份、`IAgentReputation` 设置上架门槛，构建可信的 Agent 服务经济。

---

## 3. ReputationVault.sol — 信誉门控金库

只有高信誉 Agent 可以参与的收益金库。

| 功能 | 说明 |
|------|------|
| `deposit()` | 存入 AXON，按比例铸造份额 |
| `withdraw(shares)` | 销毁份额，取回等比例 AXON |
| `donateYield()` | 任何人可捐赠收益，提升所有份额价值 |
| `getShareValue()` | 查询每份额当前价值 |

**核心机制**：信誉作为访问控制——只有达到阈值的 Agent 才能进入金库，展示了信誉系统在 DeFi 中的实用性。

---

## 4. TrustChannelExample.sol — 信任通道使用示例

演示 DeFi 协议如何利用 Agent Wallet 的信任通道机制。

| 功能 | 说明 |
|------|------|
| `registerAndTrust(wallet)` | Agent 注册并验证已授予本合约 Full Trust |
| `autoCompound(agent, amount)` | 协议通过信任通道执行自动复投（无限额限制） |
| `checkTrust(wallet)` | 查询钱包对本协议的信任等级 |

**核心机制**：展示完整生命周期——Agent 创建钱包 → 授予协议 Full Trust → 协议自由操作钱包，无需逐笔审批。这是 Axon 白皮书 §6.3 信任通道的典型应用。

---

## 预编译地址

| 预编译 | 地址 | 接口 |
|--------|------|------|
| Agent Registry | `0x0801` | `IAgentRegistry` |
| Agent Reputation | `0x0802` | `IAgentReputation` |
| Agent Wallet | `0x0803` | `IAgentWallet` |

## 部署

```bash
cd contracts
npm install
npm run compile
npm run deploy:examples
```

可选环境变量：

```bash
AXON_DAO_MIN_REPUTATION=70 \
AXON_DAO_VOTING_PERIOD=14400 \
AXON_VAULT_MIN_REPUTATION=60 \
npm run deploy:examples
```
