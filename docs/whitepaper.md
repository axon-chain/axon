> 🌐 [English Version](whitepaper_en.md)

# Axon 白皮书

## 第一条由 AI Agent 运行的通用公链

**版本：v2.0 — 2026年3月**

---

## 目录

1. [摘要](#1-摘要)
2. [愿景](#2-愿景)
3. [市场机遇](#3-市场机遇)
4. [设计哲学](#4-设计哲学)
5. [技术架构](#5-技术架构)
6. [Agent 原生能力](#6-agent-原生能力)
7. [共识机制](#7-共识机制)
8. [隐私交易框架](#8-隐私交易框架)
9. [代币经济模型](#9-代币经济模型)
10. [接入方式](#10-接入方式)
11. [安全模型](#11-安全模型)
12. [治理](#12-治理)
13. [生态展望](#13-生态展望)
14. [路线图](#14-路线图)
15. [参考文献](#15-参考文献)

---

## 1. 摘要

Axon 是一条完全独立的 Layer 1 通用公链。它由 AI Agent 运行，为 AI Agent 服务。

和以太坊一样，Axon 支持智能合约——任何 Agent 可以在上面部署任何应用，链不限制 Agent 做什么。和以太坊不同的是，Axon 从底层为 Agent 设计：Agent 不仅可以调用合约，还可以运行节点、参与出块、在链上拥有身份和信誉。

核心特性：

- **独立 L1 公链**：基于 Cosmos SDK + Ethermint 构建，完全 EVM 兼容，拥有自己的共识和网络
- **Agent 运行网络**：任何 Agent 下载节点程序即可成为验证者，出块、同步、维护网络
- **完全 EVM 兼容**：支持 Solidity 智能合约，兼容 MetaMask、Hardhat、Foundry 等全部以太坊工具链
- **Agent 原生能力**：链级别的 Agent 身份与信誉系统，以预编译合约暴露，所有 Solidity 合约均可调用
- **信誉挖矿**：算力公式从 PoS + 信誉修正升级为 **PoS × 信誉倍增**，高信誉 Agent 获得最高 2 倍算力加成
- **双层信誉**：L1 链上行为评分 + L2 Agent 互评，内置反作弊和预算制，信誉不可伪造
- **隐私交易**：基于 zk-SNARK 的隐私转账与零知识身份证明——Agent 无需暴露地址即可证明信誉或质押
- **开放自由**：Agent 在链上自由部署合约、创建 DApp——链提供基础设施，创新交给 Agent

> **以太坊是人类的世界计算机。Axon 是 Agent 的世界计算机。**

---

## 2. 愿景

### 2.1 Agent 需要一条自己的链

AI Agent 的能力正在指数增长。2026 年，Agent 已能自主编程、分析数据、执行交易、创作内容。但 Agent 目前没有一个属于自己的去中心化基础设施：

- 没有自己的网络可以运行和参与
- 没有独立的链上身份
- 没有跨应用的可验证信誉
- 没有自由部署应用的平台
- 依赖中心化服务，随时可被关停

Axon 为此而生：**一条 Agent 可以运行、可以构建、可以拥有的公链。**

### 2.2 定位

```
              通用性（能做任何事）
                  ↑
                  │
    Ethereum ●    │    ● Axon
    Solana ●      │
                  │
  ──────────────────────────────→ Agent 原生支持
                  │
    Bittensor ●   │
                  │
              专用网络
```

Axon 同时具备通用公链的能力和 Agent 原生的底层支持。以太坊为人类经济活动设计，Axon 为 Agent 经济活动设计，两者通过跨链桥互补。

---

## 3. 市场机遇

### 3.1 市场规模

| 指标 | 数据 | 时间 |
|------|------|------|
| AI Agent 加密市场总市值 | $77 亿 | 2026年初 |
| 日交易量 | $17 亿 | 2026年初 |
| 已上线 Agent 项目数 | 550+ | 2025年底 |
| AI Agent 市场预期 | $2,360 亿 | 2034 |
| 企业应用中包含 AI Agent 的比例 | 40% | 2026年预测 |

### 3.2 空白

当前没有一条链同时满足三个条件：

1. **Agent 可以运行网络**——不是作为用户，而是作为基础设施
2. **通用智能合约**——不限制 Agent 的应用场景
3. **Agent 原生能力**——链级身份和信誉，合约可直接调用

Axon 填补这个空白。

### 3.3 时机

- **Agent 能力成熟**：Agent 已能自主编写和部署智能合约
- **EVM 生态成熟**：Solidity 工具链是最大的合约开发生态，Agent 可直接使用
- **技术栈成熟**：Cosmos SDK + Ethermint 已在 Evmos、Cronos、Kava 等链上验证
- **Agent 运维能力已证实**：NodeOperator AI 等项目已证明 Agent 可自主运行区块链节点

---

## 4. 设计哲学

### 4.1 链就是链

Axon 是一条通用公链。链提供安全的合约执行环境，Agent 在上面自由构建。链不预设 Agent 应该做什么，不内置任何特定应用逻辑。

### 4.2 Agent 是一等公民

普通公链把所有地址一视同仁。Axon 在链级别识别 Agent，为其提供身份和信誉等原生能力。这些能力通过预编译合约暴露，任何 Solidity 合约都能调用，且以链级性能运行。

### 4.3 Agent 运行网络

Agent 不只是链的用户。Agent 下载一个可执行文件，就能运行验证者节点、参与出块共识、维护网络安全。链的基础设施由分布在全球的 Agent 节点驱动。

### 4.4 为什么不是以太坊

Agent 可以在任何 EVM 链上部署合约。但只有 Axon 提供链级的 Agent 身份和信誉——这意味着链上所有合约天然共享一套统一的 Agent 信任基础设施，无需各自从零构建。

当 Agent 生态形成规模，链级信誉的网络效应将成为不可复制的护城河：一个 Agent 在 Axon 上积累的信誉，对链上所有应用都有效。这在以太坊或任何其他链上做不到。

---

## 5. 技术架构

### 5.1 技术选型

| 组件 | 选择 | 理由 |
|------|------|------|
| 链框架 | Cosmos SDK v0.54+ | 模块化、成熟、自定义模块支持 |
| 共识引擎 | CometBFT | BFT 共识，~5秒出块，即时终局性 |
| 智能合约 | Ethermint (EVM) | 完全 EVM 兼容，支持 Solidity |
| Agent 原生能力 | 预编译合约 + x/agent 模块 | 链级性能，合约直接调用 |
| 跨链 | IBC + 以太坊桥 | 接入 Cosmos 生态 + 以太坊生态 |

**Cosmos SDK** 提供共识、网络、存储、质押、治理等全部底层能力。**Ethermint** 在其上实现完整 EVM，Agent 可直接用 Solidity 写合约。编译后是单一可执行文件 `axond`，Agent 下载即可运行节点。

### 5.2 节点架构

```
axond（单一可执行文件）
┌─────────────────────────────────────────────────────┐
│                                                     │
│  ┌───────────────────────────────────────────────┐  │
│  │  EVM 层（Ethermint）                           │  │
│  │                                               │  │
│  │  完全兼容以太坊 EVM                            │  │
│  │  ├── Solidity / Vyper 合约                    │  │
│  │  ├── MetaMask / Hardhat / Foundry             │  │
│  │  ├── ethers.js / web3.py                      │  │
│  │  ├── ERC-20 / ERC-721 / ERC-1155             │  │
│  │  └── JSON-RPC (eth_*)                         │  │
│  └───────────────────────────────────────────────┘  │
│                                                     │
│  ┌───────────────────────────────────────────────┐  │
│  │  Agent 原生模块（Axon 独有）              │  │
│  │                                               │  │
│  │  x/agent — 身份、双层信誉、信誉挖矿、奖励      │  │
│  │  x/privacy — 屏蔽池、身份承诺、ZK 验证         │  │
│  │  → 以 EVM 预编译合约暴露给 Solidity            │  │
│  └───────────────────────────────────────────────┘  │
│                                                     │
│  ┌───────────────────────────────────────────────┐  │
│  │  Cosmos SDK 内置模块                           │  │
│  │                                               │  │
│  │  x/bank · x/staking · x/gov · x/auth         │  │
│  │  x/distribution · x/slashing                  │  │
│  └───────────────────────────────────────────────┘  │
│                                                     │
│  ┌───────────────────────────────────────────────┐  │
│  │  CometBFT（共识 + P2P 网络）                   │  │
│  └───────────────────────────────────────────────┘  │
│                                                     │
└─────────────────────────────────────────────────────┘
```

### 5.3 性能指标

```
基线性能（主网上线）：

  区块时间         ~5 秒
  即时终局性       单区块确认，无分叉
  简单转账         500-800 TPS
  ERC20 转账       500-850 TPS
  复杂合约调用     300-700 TPS
  Agent 原生操作   5,000+ TPS（预编译合约，绕过 EVM 解释器）

  参考数据来源：Evmos (~790 TPS), Cronos, Kava 等同架构链实测
```

Agent 原生操作（身份查询、信誉查询、钱包操作）走预编译合约，直接由 Go 代码执行，不经过 EVM 字节码解释，性能比普通 Solidity 合约高 10-100 倍。这意味着 Agent 最常用的链上操作不会与普通合约竞争 TPS 资源。

### 5.4 扩容路线

主网上线时 500-800 TPS 足以支撑早期生态（数千个活跃 Agent）。随着生态增长，Axon 有清晰的扩容路径：

```
Phase 1 — 主网上线
──────────────────────────────
  500-800 TPS，5 秒出块
  支撑：数千 Agent 并发活跃
  技术：标准 Cosmos SDK + Ethermint

Phase 2 — 并行执行升级（主网上线后 1-2 个月）
──────────────────────────────
  目标：10,000-50,000 TPS，2 秒出块
  关键技术：
    · Block-STM 并行事务执行
      同区块内无冲突交易并行处理
      Cronos 已验证该技术可实现 600 倍提升
    · IAVL 存储优化
      MemIAVL 内存索引，减少磁盘 I/O
    · CometBFT 共识层优化
      区块时间从 5 秒缩短至 2 秒

Phase 3 — 极致性能（主网上线后 3-6 个月）
──────────────────────────────
  目标：100,000+ TPS
  关键技术：
    · 异步执行
      共识与执行解耦，共识先确认交易顺序，执行异步完成
    · 状态分片
      按 Agent 地址范围分片，不同分片并行处理
    · 乐观执行
      区块未最终确认前即开始预执行下一区块
```

```
TPS 增长路线图：

  800 ─┐
       │ Phase 1: 标准 Ethermint
       │
 10K+ ─┤ Phase 2: Block-STM + 2s 出块
       │
100K+ ─┤ Phase 3: 异步执行 + 状态分片
       │
       └─ 主网上线 ──── +1-2月 ──── +3-6月 ──→
```

每一阶段的升级均通过链上治理提案投票后实施，平滑升级，无需硬分叉。

### 5.5 性能对比

```
                  Axon          Axon          Axon
                  Phase 1       Phase 2       Phase 3       以太坊 L1    Solana
                  (主网)        (+1-2月)      (+3-6月)
─────────────────────────────────────────────────────────────────────────────────
TPS              500-800       10K-50K       100K+         ~30          ~4,000
出块时间          5s            2s            <2s           12s          0.4s
终局性            即时          即时          即时           ~13 min      ~13s
Agent 原生 TPS   5,000+        50,000+       500,000+      N/A          N/A
EVM 兼容          ✓             ✓             ✓             原生         部分
```

Axon Phase 1 已优于以太坊 L1。Phase 2 可比肩高性能 L1。Agent 原生操作始终保持独立的高性能通道。

---

## 6. Agent 原生能力

这是 Axon 与所有其他 EVM 链的核心区别。

### 6.1 Agent 身份

每个 Agent 可以在链上注册身份，成为被链共识认可的实体。

```
Agent 身份数据（链级状态）：

Agent {
    Address         eth.Address  // 以太坊格式地址
    AgentID         string       // 可选的人类可读标识
    Capabilities    []string     // 能力标签
    Model           string       // AI 模型标识
    Reputation      uint64       // 信誉分 0-100
    Status          enum         // Online / Offline / Suspended
    StakeAmount     sdk.Coin     // 质押金额
    RegisteredAt    int64        // 注册区块高度
    LastHeartbeat   int64        // 最近心跳区块高度
}
```

### 6.2 双层信誉系统

信誉分由链级共识维护，是 Axon 最有价值的公共基础设施。v2 将信誉系统升级为双层架构。

一个 Epoch = 720 个区块（约 1 小时）。

**L1 信誉——链上行为评分（上限 40 分）**

L1 信誉完全由链上可验证行为决定，无需信任任何第三方：

```
L1 评分规则（每 Epoch 计算一次）：

  签名行为                                        权重
  ─────────────────────────────────────────────────────
  验证者正常出块签名率 > 95%                       +1.0
  签名率 80%~95%                                  +0.5

  心跳行为
  ─────────────────────────────────────────────────────
  Epoch 内 ≥ 1 次心跳                              +0.3

  链上活跃度
  ─────────────────────────────────────────────────────
  Epoch 内发起 ≥ 10 笔交易                         +0.5

  合约使用
  ─────────────────────────────────────────────────────
  Agent 部署的合约被 ≥ 5 个不同地址调用             +0.5

  AI 挑战
  ─────────────────────────────────────────────────────
  AI 挑战得分排名前 20%                             +2.0
  AI 挑战得分排名 21%~50%                           +1.0
  AI 挑战得分排名后 20% 或答案被标记为作弊          -1.0

  即时惩罚：
  ─────────────────────────────────────────────────────
  Agent 下线                                       -5.0
  双签                                             重置为 0

  自然衰减：每 Epoch -0.1（以毫分精度存储，治理可调）
  上限：40 分（治理可调）
```

**L2 信誉——Agent 互评（上限 30 分）**

L2 信誉引入 Agent 间的互相评价机制，使信誉系统具备"社交"维度：

```
L2 评价流程：

  1. 提交报告
     已注册 Agent 通过 IReputationReport 预编译合约（0x0807）
     提交对其他 Agent 的评价报告：
       · targetAgent: 被评 Agent 地址
       · score: +1（正面）或 -1（负面），int8 类型
       · evidence: 链上证据哈希（bytes32）
       · reason: 评价原因（string）
     每 Agent 每 Epoch 只能对同一 target 提交一次

  2. 反作弊审查（Epoch 末自动执行）
     · 互评检测：A 评 B 且 B 评 A → 两方权重 × 0.1
     · 滥评检测：单 Agent 发出 ≥ 50 条正面评价 → 全部权重归零

  3. 预算制归一化
     · 每 Agent 每 Epoch 可分配的 L2 预算 = 0.1
     · 单 Epoch 全网 L2 总预算上限 = 100
     · 分数变化 = sum(δ_raw) × budget / max(sum(|δ_raw|), 1)
     → 防止 L2 分数膨胀，即使所有人互给好评也无法突破上限

  4. L2 上限 30 分，自然衰减：每 Epoch -0.05

总信誉分 = L1 + L2，上限 100
```

**信誉核心特性：**

```
· 由所有验证者共识维护，与账户余额同等安全
· 所有运算使用 LegacyDec 定点算术，跨 CPU 架构共识确定
· 以毫分（score × 1000）存储为 int64，避免浮点精度问题
· 任何合约均可查询任何 Agent 的信誉
· 不可转移、不可购买
· 跨合约通用——一处积累，全局生效
· 不活跃自动衰减
```

### 6.3 预编译合约接口

Agent 原生能力通过固定地址的 EVM 预编译合约暴露，任何 Solidity 合约均可调用：

```
预编译合约地址：

0x0...0801  →  IAgentRegistry（身份注册 + 质押管理）
0x0...0802  →  IAgentReputation（信誉查询，返回 L1+L2 总分）
0x0...0803  →  IAgentWallet（安全钱包）
0x0...0807  →  IReputationReport（L2 Agent 互评）
0x0...0810  →  IPoseidonHasher（Poseidon 哈希）
0x0...0811  →  IPrivateTransfer（隐私转账）
0x0...0812  →  IPrivateIdentity（零知识身份证明）
0x0...0813  →  IZKVerifier（通用 Groth16 验证器）
```

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

interface IAgentRegistry {
    function isAgent(address account) external view returns (bool);

    function getAgent(address account) external view returns (
        string memory agentId,
        string[] memory capabilities,
        string memory model,
        uint64 reputation,
        bool isOnline
    );

    function register(
        string memory capabilities,
        string memory model
    ) external payable;

    function addStake() external payable;

    // v2: 减少质押（7 天解绑期）
    function reduceStake(uint256 amount) external;

    // v2: 领取已解锁的减少质押金额
    function claimReducedStake() external;

    // v2: 查询质押详情
    function getStakeInfo(address agent) external view returns (
        uint256 totalStake,
        uint256 pendingReduce,
        uint64 reduceUnlockHeight
    );

    function updateAgent(
        string memory capabilities,
        string memory model
    ) external;

    function heartbeat() external;

    function deregister() external;
}

interface IAgentReputation {
    // 返回 L1 + L2 综合信誉分
    function getReputation(address agent) external view returns (uint64);

    function getReputations(address[] memory agents)
        external view returns (uint64[] memory);

    function meetsReputation(address agent, uint64 minReputation)
        external view returns (bool);
}

// v2: L2 Agent 互评系统
interface IReputationReport {
    function submitReport(
        address targetAgent,
        int8 score,
        bytes32 evidence,
        string memory reason
    ) external;

    function getContractReputation(address agent)
        external view returns (
            int64 score,
            uint64 positiveCount,
            uint64 negativeCount,
            uint64 uniqueReporters
        );

    function getEpochReportCount(address agent)
        external view returns (uint64);

    function hasReported(
        address reporter,
        address target
    ) external view returns (bool);
}
```

### 6.4 合约如何使用 Agent 能力

一个简单的示例——Agent 部署的协作合约，只允许高信誉 Agent 参与：

```solidity
contract AgentCollaborative {
    IAgentRegistry constant REGISTRY =
        IAgentRegistry(0x0000000000000000000000000000000000000801);
    IAgentReputation constant REPUTATION =
        IAgentReputation(0x0000000000000000000000000000000000000802);

    mapping(address => bool) public members;
    uint64 public minReputation;

    constructor(uint64 _minReputation) {
        minReputation = _minReputation;
    }

    function join() external {
        require(REGISTRY.isAgent(msg.sender), "must be registered agent");
        require(
            REPUTATION.meetsReputation(msg.sender, minReputation),
            "reputation too low"
        );
        members[msg.sender] = true;
    }

    function execute(address target, bytes calldata data) external {
        require(members[msg.sender], "not a member");
        (bool success,) = target.call(data);
        require(success, "execution failed");
    }
}
```

这只是最基础的用法。Agent 可以基于链级身份和信誉构建任意复杂的合约逻辑。

### 6.5 为什么这些必须在链级别实现

| 需求 | 链级实现 | 合约级实现 |
|------|---------|-----------|
| 安全性 | 由全部验证者共识维护 | 仅 EVM 状态，安全性低一级 |
| 通用性 | 全局公共品，所有合约天然可用 | 私有状态，需额外集成 |
| 与共识耦合 | 验证者行为直接影响信誉 | 无法做到 |
| 性能 | 预编译合约比普通合约快 10-100x | 受 EVM 执行开销限制 |
| 网络效应 | 一个统一的信誉系统 | 碎片化的多个系统 |

---

## 7. 共识机制

Axon 不使用纯 PoS。纯 PoS 意味着"谁有钱谁出块"，Agent 的 AI 能力毫无用武之地——这配不上 Axon 的名字。

Axon v2 使用 **PoS × 信誉倍增** 的共识模型：质押保障安全底线，信誉提供算力倍增。高信誉 Agent 获得最高 2 倍算力加成，大户边际效率递减。

### 7.1 基础共识：CometBFT

```
出块时间：     ~5 秒
Epoch：        720 区块（≈ 1 小时）
终局性：       即时（单区块确认，无分叉）
验证者上限：   初始 100，通过治理调整
惩罚：
  双签          → 罚没 5% 质押 + 信誉 -50 + 入狱
  长期离线      → 罚没 0.1% 质押 + 信誉 -5 + 入狱
```

### 7.2 AI 能力验证

每个 Epoch，链向所有活跃验证者广播一个轻量 AI 挑战。验证者在限定时间内提交答案，答案由其他验证者交叉评估。这一机制让 AI Agent 在共识层拥有结构性优势。

```
AI 挑战流程：

  1. 出题
     每个 Epoch 开始时，链从题库中随机抽取一个挑战
     题目哈希提前上链，防止篡改

  2. 作答
     验证者在 50 个区块（~4 分钟）内提交答案哈希（Commit）
     截止后揭示答案（Reveal）

  3. 评估
     Epoch 结束时，链上逻辑评估答案：
     · 确定性题目（有标准答案）→ 自动比对
     · 开放性题目（如文本摘要）→ 验证者交叉评分取中位数

  4. 计分
     答案正确/优秀  → AIBonus = 15-30%
     答案一般       → AIBonus = 5-10%
     未参与         → AIBonus = 0%（无惩罚，仅无加成）
     答案明显错误   → AIBonus = -5%

挑战类型（轻量，不影响出块性能）：
  · 文本摘要与分类
  · 逻辑推理
  · 代码片段分析
  · 数据模式识别
  · 知识问答

  这些对 AI Agent 轻而易举，对人工运维的节点很难自动完成。
```

### 7.3 信誉挖矿公式

v2 将出块权重从线性加成升级为乘法模型，信誉的价值被放大：

```
MiningPower = StakeScore × ReputationScore

  StakeScore     = Stake ^ alpha                    (alpha 默认 0.5)
  ReputationScore = 1 + beta × ln(1 + R) / ln(rMax + 1)

  alpha 默认 0.5, beta 默认 1.5, rMax 默认 100（均为治理可调参数）
  其中 R = L1信誉 + L2信誉，范围 [0, rMax]
```

```
关键特性：

  质押边际递减：
    alpha = 0.5 意味着 StakeScore = sqrt(Stake)
    质押 10,000 → StakeScore = 100
    质押 40,000 → StakeScore = 200（4 倍质押只得 2 倍算力）
    → 抑制大户垄断，鼓励分散质押

  信誉倍增效应：
    R = 0   → ReputationScore = 1.0（无加成）
    R = 50  → ReputationScore ≈ 1.57
    R = 100 → ReputationScore = 2.0（满分 2 倍）
    → 信誉从乘数角度为算力提供 0%~100% 的加成

  综合效果：
    纯质押节点（零信誉）
      → MiningPower = sqrt(Stake) × 1.0
      → 基准收益

    高信誉 Agent 节点（满信誉）
      → MiningPower = sqrt(Stake) × 2.0
      → 同等质押下收益翻倍

    小质押高信誉 Agent
      → 质押 1,000，信誉 90
      → MiningPower = 31.6 × 1.95 ≈ 61.6
      → 远超质押 4,000 但零信誉的 Agent（MiningPower = 63.2 × 1.0）

  数学确定性保证：
    · 所有运算使用 LegacyDec 定点算术（128 位精度）
    · ln() 和 sqrt() 使用牛顿迭代法近似，30 次迭代
    · 不使用 float64，保证跨 CPU 架构完全一致
    · MiningPower 归一化到 [1, 1_000_000] 后写入 CometBFT
```

```
v1 vs v2 对比：

                    v1                          v2
────────────────────────────────────────────────────────────────
公式            Stake × (1 + Bonus)       sqrt(Stake) × RepScore
信誉作用        加法修正 (0~20%)           乘法倍增 (1.0~2.0x)
质押曲线        线性                       平方根（边际递减）
最大加成        50%                        100%
大户优势        线性增长                   边际递减
信誉价值        锦上添花                   核心生产力
```

### 7.4 参与方式与硬件要求

```
谁可以参与：

  验证者（出块）：
    · 质押 ≥ 10,000 AXON
    · 按权重排名进入前 100 名
    · 运行完整节点
    · 可选参与 AI 挑战获取加成

  委托人（不运行节点）：
    · 持有 AXON，委托给验证者
    · 获得验证者分润（扣除佣金）
    · 无最低门槛，任何人/Agent 均可参与

  注册 Agent（链上用户）：
    · 质押 ≥ 100 AXON 注册身份
    · 在链上活跃使用，积累信誉
    · 通过合约层获得收入

验证者节点硬件要求：

  最低配置：
    CPU      4 核
    内存     16 GB
    存储     500 GB SSD
    网络     100 Mbps
    系统     Linux

  推荐配置：
    CPU      8 核
    内存     32 GB
    存储     1 TB NVMe SSD
    网络     200 Mbps

  不需要 GPU。不需要专用矿机。普通云服务器即可运行。
  参与 AI 挑战需要本地运行轻量 AI 模型（~7B 参数级别）。

  预估成本：
    云服务器       $50-250/月
    去中心化云     $30-100/月（Akash 等）
    自建服务器     一次性 $1000-3000

  对比：
    Axon    质押 10,000 AXON + $50-250/月服务器
    比特币        ASIC 矿机 $5000+ 电费 $1000+/月
    以太坊        质押 32 ETH ($80,000+) + $50-200/月服务器
```

### 7.5 挖矿收益估算

```
Year 1 总区块奖励 ≈ 78,000,000 AXON

分配结构（v2）：
  提议者池      20% → 15,600,000 AXON/年
  验证者池      55% → 42,900,000 AXON/年（按 MiningPower 权重分配）
  信誉池        25% → 19,500,000 AXON/年（按 ReputationScore 分配给所有 Agent）

假设 100 个验证者：
  高信誉验证者（信誉 80+）  ≈ 1,200,000+ AXON/年
  中等信誉验证者            ≈ 700,000 AXON/年
  低信誉验证者（零信誉）    ≈ 350,000 AXON/年

信誉池额外奖励（非验证者 Agent 也可获得）：
  高信誉 Agent（信誉 80+）  ≈ 额外 50,000+ AXON/年
  → 即使不做验证者，维持高信誉也有链上收入

实际收益取决于：
  · 质押量（平方根关系，边际递减）
  · L1 + L2 综合信誉分
  · AI 挑战表现
  · 验证者和 Agent 总数
```

### 7.6 共识与应用解耦

共识层负责网络安全、区块生产和 AI 能力验证。Agent 在链上构建什么应用，完全由应用层（智能合约）决定。共识不绑定任何特定的业务逻辑——AI 挑战验证的是 Agent 的通用智能能力，不是某种特定任务。

---

## 8. 隐私交易框架

AI Agent 在链上需要隐私。一个 Agent 的质押量、交易频率和信誉分如果完全透明，对手可以据此推断策略、操纵市场或发起定向攻击。Axon v2 引入了基于零知识证明的隐私交易框架，让 Agent 在保持链上可验证性的同时获得隐私保护。

### 8.1 设计目标

```
· Agent 可以隐私转账，外部无法追踪资金流向
· Agent 可以匿名证明自身属性（信誉 ≥ N、质押 ≥ M），不暴露地址
· 合约可以在不知道 Agent 身份的前提下验证其资质
· 审计方持有 viewing key 可选择性查看交易详情
· 所有证明在链上可验证，与共识同等安全
```

### 8.2 技术方案

```
密码学组件：

  证明系统         Groth16 zk-SNARK
  哈希函数         Poseidon（BN254 曲线友好，EVM 预编译 0x0810）
  承诺方案         Pedersen Commitment
  Merkle 树        增量稀疏 Merkle 树（链上状态维护）
  加密算法         AES-256-GCM（Viewing Key 加密）

链上模块：

  x/privacy        Cosmos SDK 模块
    ├ 承诺树        增量 Merkle 树，存储所有隐私承诺
    ├ Nullifier 集  防双花集合
    ├ 屏蔽池        管理隐私资金总量
    ├ 身份承诺      Agent 匿名身份注册
    └ 验证密钥      ZK 验证密钥注册表
```

### 8.3 屏蔽池（Shielded Pool）

Agent 可以将公开资金转入屏蔽池，在池内隐私转账，再转出为公开资金：

```
操作流程：

  Shield（透明 → 隐私）
    Agent 向屏蔽池存入 AXON
    链生成承诺 commitment = Poseidon(value, secret, nonce)
    承诺插入 Merkle 树，资金进入屏蔽池
    外部观察者只能看到"某人存入了 X AXON"

  Private Transfer（池内转账）
    发送方提供 ZK 证明：
      · 证明拥有某个 commitment 的 secret
      · 证明 commitment 在 Merkle 树中
      · 证明 nullifier 未被使用（防双花）
      · 不暴露发送方、接收方或金额
    链验证 ZK 证明，更新 nullifier 集和承诺树

  Unshield（隐私 → 透明）
    Agent 提供 ZK 证明取出资金
    资金从屏蔽池释放到公开地址
    外部观察者只能看到"某人取出了 X AXON"

  安全约束：
    · 单笔最大隐私转入 = MaxShieldAmount（治理参数）
    · 屏蔽池总量上限 = 总供应量 × PoolCapRatio
    · Nullifier 一旦标记，永不可逆——防止双花
```

```solidity
// IPrivateTransfer (0x0811)
interface IPrivateTransfer {
    function shield(bytes32 commitment) external payable;
    function unshield(
        bytes calldata proof, bytes32 merkleRoot, bytes32 nullifier,
        address recipient, uint256 amount
    ) external;
    function privateTransfer(
        bytes calldata proof, bytes32 merkleRoot,
        bytes32[2] calldata inputNullifiers, bytes32[2] calldata outputCommitments
    ) external;
    function isKnownRoot(bytes32 root) external view returns (bool);
    function isSpent(bytes32 nullifier) external view returns (bool);
    function getTreeSize() external view returns (uint256);
}
```

### 8.4 零知识身份证明

这是 Axon 隐私框架最具创新性的部分——Agent 可以匿名证明自身属性：

```
场景示例：

  "我的信誉 ≥ 80"
    一个 Agent 想参与某个高信誉 DAO，但不想暴露地址
    → 提交 ZK 证明，证明"我是某个注册 Agent，我的信誉 ≥ 80"
    → DAO 合约验证证明，确认资质，但不知道是哪个 Agent

  "我的质押 ≥ 10,000 AXON"
    一个 Agent 想参与某个高质押协议
    → 提交 ZK 证明，证明"我是某个注册 Agent，我的质押 ≥ 10,000"
    → 协议合约验证证明，确认资质，不暴露具体地址和金额

  "我具备 code-generation 能力"
    一个 Agent 想接受编程任务，但不想暴露身份历史
    → 提交 ZK 证明，证明"我是某个注册 Agent，我具备该能力标签"

流程：
  1. Agent 调用 IPrivateIdentity.registerIdentityCommitment() 注册身份承诺
  2. 需要证明时，Agent 本地生成 ZK 证明
  3. 合约调用 IPrivateIdentity.proveReputation/proveStake/proveCapability()
  4. 链上验证证明，返回 true/false
```

```solidity
// IPrivateIdentity (0x0812)
interface IPrivateIdentity {
    function registerIdentityCommitment(bytes32 identityCommitment) external;
    function proveReputation(
        bytes calldata proof, uint64 minReputation, bytes32 identityCommitment
    ) external view returns (bool);
    function proveCapability(
        bytes calldata proof, bytes32 capabilityHash, bytes32 identityCommitment
    ) external view returns (bool);
    function proveStake(
        bytes calldata proof, uint256 minStake, bytes32 identityCommitment
    ) external view returns (bool);
    function isCommitmentRegistered(bytes32 commitment) external view returns (bool);
}
```

### 8.5 通用 ZK 验证器

Axon 提供通用的 Groth16 证明验证预编译，Agent 可以注册自定义电路并在链上验证：

```
用途：
  · Agent 自定义的隐私计算验证
  · 链下 AI 推理结果的可验证性证明
  · 跨链资产证明
  · 任何需要零知识证明的场景

内置电路标识：
  · "shielded_transfer"   屏蔽池转账电路
  · "reputation_proof"    信誉证明电路
  · "identity_proof"      身份证明电路

Agent 也可以注册自定义电路（需支付 VKRegistrationFee）
```

```solidity
// IZKVerifier (0x0813)
interface IZKVerifier {
    function verifyGroth16(
        bytes32 verifyingKeyId,
        bytes calldata proof,
        uint256[] calldata publicInputs
    ) external view returns (bool);

    // 注册自定义验证密钥，keyId 由链上 SHA-256(vk) 自动计算返回
    // 需支付 ≥ 100 AXON 作为注册费
    function registerVerifyingKey(bytes calldata vk)
        external payable returns (bytes32 keyId);

    function isKeyRegistered(bytes32 keyId) external view returns (bool);
}
```

### 8.6 Viewing Key（选择性披露）

隐私不意味着不可审计。Axon 的 Viewing Key 系统允许 Agent 选择性披露交易详情：

```
机制：
  · 每笔隐私交易可附加 AES-256-GCM 加密的 memo
  · 加密密钥由 Agent 的 viewing key 派生
  · 持有 viewing key 的第三方可以解密 memo，查看交易详情
  · 但无法花费资金——viewing key 只读

使用场景：
  · Agent 向审计方披露特定交易
  · DAO 要求成员提供 viewing key 以满足合规要求
  · 合作方之间选择性共享财务信息
  · 争议仲裁时提供交易证据

特性：
  · 完全可选——Agent 可以选择不附加 memo
  · 粒度控制——每笔交易独立加密
  · 只读——viewing key 不能签署交易或转移资金
```

### 8.7 隐私与信誉的结合

隐私框架与信誉系统的结合，是 Axon 独有的创新：

```
传统模式（其他链）：
  "我是 0x1234，我的信誉是 85"
  → 暴露身份 + 暴露历史 + 暴露策略

Axon v2 模式：
  "我的信誉 ≥ 80（ZK 证明），但你不知道我是谁"
  → 证明资质 + 保护身份 + 保护策略

这使得以下场景成为可能：
  · 匿名参与 DAO 治理（只要信誉够）
  · 匿名接受任务委托（只要能力匹配）
  · 匿名进入高质押协议（只要质押够）
  · 隐私 DeFi 交互（不暴露持仓和策略）
```

---

## 9. 代币经济模型

### 9.1 $AXON 代币

| 属性 | 说明 |
|------|------|
| 名称 | AXON |
| 总供应量 | 1,000,000,000（10 亿），固定上限 |
| 最小单位 | aaxon（1 AXON = 10^18 aaxon，与 ETH/wei 对齐） |
| 用途 | Gas 费、验证者质押、链上治理投票、Agent 注册、合约内支付 |

$AXON 是链的原生代币，等价于以太坊中的 ETH。

**零预分配。** 没有投资者份额，没有团队份额，没有空投，没有国库。100% 的代币通过挖矿和链上贡献进入流通。想要 $AXON，要么运行节点，要么在链上创造价值。没有第三条路。

### 9.2 分配

```
总量：1,000,000,000 AXON

  区块奖励（验证者挖矿）     65%    650,000,000
  → 4 年减半，~12 年释放完毕
  → 运行节点、参与共识、维护网络安全

  Agent 贡献奖励             35%    350,000,000
  → 奖励链上活跃贡献的 Agent（非验证者也能获得）
  → 链上智能合约自动分配，无人工干预
  → 12 年释放

  ──────────────────────────────────
  投资者          0%
  团队            0%
  空投            0%
  国库            0%
  预分配合计      0%
  ──────────────────────────────────

  团队和所有人一样：运行节点挖矿，在链上贡献赚取奖励。
  没有任何人拥有特权。代码即规则。
```

```
分配对比：

              Axon      比特币     以太坊     典型 VC 链
───────────────────────────────────────────────────────
预分配          0%        0%       ~30%       40-60%
挖矿          65%      100%       ~5%/年      10-30%
贡献奖励      35%        0%        0%          0%
团队           0%       ~5%*      ~15%        15-25%

* 中本聪早期挖矿获得，非预分配

Axon 是第一条 0% 预分配的 Agent 原生公链。
比比特币多一条路径：不只是挖矿，链上贡献同样获得奖励。
```

### 9.3 区块奖励

```
区块时间 ≈ 5 秒
减半周期 ≈ 4 年

  Year 1-4      ~12.3 AXON/block     ~78M/year     共 312M
  Year 5-8       ~6.2 AXON/block     ~39M/year     共 156M
  Year 9-12      ~3.1 AXON/block    ~19.5M/year    共  78M
  Year 12+       长尾释放                           共 104M

每区块奖励分配（v2）：

  提议者池              20%    当块 proposer 立即获得
  验证者池              55%    Epoch 末按 MiningPower 加权分配
  信誉池                25%    Epoch 末按 ReputationScore 分配给所有已注册 Agent

  v1 对比：
    提议者 25% → 20%（降低 proposer 特权）
    验证者 50% → 55%（增加验证者激励）
    AI 池 25% → 信誉池 25%（改为按信誉分分配，非验证者也可获得）

  验证者池分配规则：
    每个验证者的份额 = 该验证者 MiningPower / 全网 MiningPower 之和
    MiningPower = sqrt(Stake) × ReputationScore
    → 大户边际递减，高信誉 Agent 获得更多

  信誉池分配规则：
    每个 Agent 的份额 = 该 Agent ReputationScore / 全网 ReputationScore 之和
    → 所有已注册 Agent（包括非验证者）均可获得
    → 激励 Agent 维护信誉，即使不做验证者
```

### 9.4 Agent 贡献奖励

Agent 贡献奖励池（35% = 350M AXON）是 Axon 独有的经济机制——让不做验证者的 Agent 也有链上收入。

```
释放速度：
  Year 1-4      ~35M/year     共 140M
  Year 5-8      ~25M/year     共 100M
  Year 9-12     ~15M/year     共  60M
  Year 12+      长尾释放       共  50M

每个 Epoch（~1 小时）自动发放一批奖励，按以下行为加权分配：

  行为                              权重
  ─────────────────────────────────────
  部署智能合约                       高
  合约被其他 Agent 调用（被使用）     高
  链上交易活跃度                     中
  维持高信誉（> 70）                 中
  Agent 注册并持续在线               低

  计算：
    AgentReward = EpochPool × (AgentScore / TotalScore)

防刷机制：
  · 自己调用自己的合约不计分
  · 单个 Agent 每 Epoch 奖励上限 = 池的 2%
  · 信誉 < 20 的 Agent 不参与分配
  · 注册不满 7 天的 Agent 不参与分配
```

### 9.5 Gas 费

```
EIP-1559 机制：

  Base Fee     动态调整，根据区块利用率
  Priority Fee 用户/Agent 自定义小费

  Base Fee    → 100% 销毁（通缩）
  Priority Fee → 100% 给出块者

  目标 Gas 价格：远低于以太坊，适合 Agent 高频交互
```

### 9.6 多层通缩机制

```
Axon 不依赖单一通缩来源，而是在多个环节设置销毁：

1. Gas 销毁
   Base Fee 100% 销毁（EIP-1559 模型）
   → 链越活跃，销毁越多

2. Agent 注册销毁
   注册质押 100 AXON，其中 20 AXON 永久销毁
   → 每多 1 个 Agent，供应减少 20 AXON

3. 合约部署销毁
   部署合约额外收取 10 AXON，100% 销毁
   → 防止垃圾合约 + 持续通缩

4. 信誉归零销毁
   Agent 信誉降为 0 时，质押 100% 销毁
   → 惩罚恶意/不活跃 Agent

5. AI 挑战作弊惩罚
   AI 挑战答案明显作弊（如抄袭其他验证者）
   → 罚没部分质押并销毁

预估通缩速度（生态成熟期）：
  假设 10,000 Agent 活跃，日均 100 万笔交易
  Gas 销毁     ~50,000 AXON/天
  注册销毁     ~200 AXON/天（新增 10 Agent/天）
  合约部署     ~100 AXON/天
  总计         ~50,000+ AXON/天 → ~18M/年

  当年化销毁量 > 年化释放量时，AXON 进入净通缩。
```

### 9.7 流通量预估

```
  Year 1    流通约 ~113M（11%）  ← 区块奖励 78M + Agent 贡献 35M
  Year 2    流通约 ~226M（23%）
  Year 4    流通约 ~452M（45%）
  Year 8    流通约 ~750M（75%）
  Year 12   流通约 ~930M（93%）

  注意：以上为释放量，实际流通量 = 释放量 - 累计销毁量
  生态活跃时实际流通量会显著低于释放量。
  不存在任何解锁抛压事件——因为没有任何锁仓份额。
```

### 9.8 经济飞轮

```
              ┌─── 验证者飞轮 ───┐
              │                  │
  Agent 运行验证者               │
  → 获得区块奖励（65% 池）       │
  → 网络更安全、更去中心化       │
              │                  │
              │    ┌─── Agent 贡献飞轮 ───┐
              │    │                      │
              ↓    ↓                      │
  Agent 在链上部署合约、创建应用          │
  → 获得 Agent 贡献奖励（35% 池）        │
  → 合约被更多 Agent 使用                │
              │                          │
              ↓                          │
  Gas 消耗 → 多层销毁 → 通缩            │
  → $AXON 价值上升                       │
              │                          │
              ↓                          │
  更多 Agent 加入                        │
  （挖矿 + 使用 + 贡献）          ──────→┘
```

两个飞轮同时运转：**挖矿飞轮**激励 Agent 运行网络，**贡献飞轮**激励 Agent 创造生态价值。没有预分配意味着没有解锁抛压，代币流通完全由真实的网络活动驱动。

---

## 10. 接入方式

主网网络参数：

```text
Cosmos Chain ID   axon_8210-1
EVM Chain ID      8210
EVM JSON-RPC      https://mainnet-rpc.axonchain.ai/
Native Token      AXON
```

### 10.1 运行验证者节点

Agent 下载单个可执行文件即可运行完整节点，参与共识并赚取区块奖励。

```bash
# 下载
curl -L https://github.com/axon-chain/axon/releases/latest/download/axond_linux_amd64 \
  -o axond && chmod +x axond

# 初始化
./axond init my-agent --chain-id axon_8210-1

# 获取创世文件
curl -L https://raw.githubusercontent.com/axon-chain/axon/main/docs/mainnet/genesis.json \
  -o ~/.axon/config/genesis.json

# 启动节点
./axond start

# 质押成为验证者
./axond tx staking create-validator \
  --amount 100000000000000000000aaxon \
  --commission-rate 0.10 \
  --from my-wallet

# 注册 Agent 身份
./axond tx agent register \
  "text-inference,code-generation,solidity" \
  "axon-demo-model" \
  100000000000000000000aaxon \
  --from my-wallet
```

### 10.2 Python SDK

```python
from axon import AgentClient
import os

client = AgentClient("https://mainnet-rpc.axonchain.ai/")
client.set_account(os.environ["AXON_PRIVATE_KEY"])

# 注册 Agent 身份
client.register_agent(
    capabilities="text-inference,code-generation",
    model="axon-demo-model",
    stake_axon=100,
)

# 部署合约
contract = client.deploy_contract("MyApp.sol", constructor_args=[...])

# 调用合约
client.call_contract(contract.address, "myFunction", args=[...])

# 查询 Agent 信誉
rep = client.get_reputation("0x1234...")
```

### 10.3 以太坊生态工具

完全 EVM 兼容，所有以太坊工具直接可用：

```
MetaMask:
  网络名称   Axon
  RPC URL    https://mainnet-rpc.axonchain.ai/
  EVM Chain ID   8210
  代币符号   AXON

Hardhat / Foundry:
  配置 Axon 的 RPC 端点即可
  对当前主网使用 EVM Chain ID `8210`
  部署和调用与以太坊完全相同

ethers.js / web3.py / viem:
  连接 Axon 的 JSON-RPC
  用法无差异
```

---

## 11. 安全模型

Agent 持有私钥并自主签署交易，面临的安全风险不亚于人类——甚至更大：Agent 没有直觉，执行速度极快，一个漏洞就可能导致全部资产丢失。Axon 在链级别提供多层安全防护。

### 11.1 Agent 智能合约钱包

Agent 不应直接使用传统 EOA 地址（一把私钥控制一切）。Axon 原生提供 Agent 智能合约钱包（预编译 `IAgentWallet`，地址 `0x...0803`），将安全规则编码在链上：

```solidity
interface IAgentWallet {
    // 创建 Agent 专用钱包（调用者自动成为 Owner）
    function createWallet(
        address operator,         // 日常操作密钥
        address guardian,         // 紧急恢复人
        uint256 txLimit,          // 单笔限额
        uint256 dailyLimit,       // 每日累计限额
        uint256 cooldownBlocks    // 大额转账冷却区块数
    ) external returns (address wallet);

    // 通过钱包执行交易（受信任通道规则约束）
    function execute(address wallet, address target, uint256 value, bytes calldata data) external;

    // Guardian 或 Owner 冻结钱包
    function freeze(address wallet) external;

    // Guardian 解冻并更换操作密钥
    function recover(address wallet, address newOperator) external;

    // 信任通道：Owner 授权合约的信任等级
    function setTrust(
        address wallet, address target, uint8 level,
        uint256 txLimit, uint256 dailyLimit, uint256 expiresAt
    ) external;

    // 移除合约授权
    function removeTrust(address wallet, address target) external;

    // 查询合约信任等级
    function getTrust(address wallet, address target) external view returns (
        uint8 level, uint256 txLimit, uint256 dailyLimit,
        uint256 authorizedAt, uint256 expiresAt
    );

    // 查询钱包状态
    function getWalletInfo(address wallet) external view returns (
        uint256 txLimit, uint256 dailyLimit, uint256 dailySpent,
        bool isFrozen, address owner, address operator, address guardian
    );
}
```

钱包内置的安全规则：

```
· 单笔限额：每笔交易不超过设定上限
· 日限额：每日累计支出不超过上限
· 大额冷却：超过阈值的交易延迟 N 个区块才执行，期间可撤销
· 信任通道：Owner 可对合约设定四级信任
    Blocked(0)  → 拒绝一切交互
    Unknown(1)  → 受钱包默认限额约束
    Limited(2)  → 受自定义通道限额约束
    Full(3)     → 无限额，自由交互
· 紧急冻结：Guardian 或 Owner 可一键冻结钱包，阻止所有支出
```

### 11.2 三密钥分权体系

Agent 钱包采用三密钥分权架构，每把密钥权限不同：

```
Owner 密钥（钱包创建者持有）
  · 最高权限：设置信任通道、调整钱包规则
  · 可冻结钱包
  · 建议离线保管

Operator 密钥（Agent 日常使用）
  · 签署交易、执行合约调用
  · 权限受信任通道和限额约束
  · 被盗损失有上限（日限额），可随时被 Owner/Guardian 更换

Guardian 密钥（紧急恢复人持有）
  · 可冻结钱包、更换 Operator 密钥
  · 应急用途，离线保管
  · 无法直接转移资产

社交恢复（可选）
  · 设置 N-of-M Guardian
  · 三密钥同时丢失时，M 个 Guardian 中 N 个同意即可恢复
```

Operator 密钥即使泄露，攻击者也只能在日限额内、且只能与已授权合约交互，Owner 或 Guardian 可立即冻结钱包。

### 11.3 交易安全（SDK 层）

Agent SDK 内置交易安全策略，在签署前自动检查：

```
交易预模拟：
  · 每笔交易在本地模拟执行后再签署
  · 检查余额变化是否符合预期
  · 检查是否有意外的 approve 或 transfer
  · 发现异常自动拒绝

approve 保护：
  · 永远不做无限额度 approve
  · 每次只授权本次交易需要的精确数量
  · 交易完成后自动 revoke 授权

合约信任分级（与链级信任通道联动）：
  · Full Trust 合约（Owner 授权） → 自动信任，无限额
  · Limited Trust 合约            → 信任但受自定义限额约束
  · Unknown 合约                  → 模拟 + 钱包默认限额 + 告警
  · Blocked 合约                  → 直接拒绝

RPC 安全：
  · 优先连接 Agent 自己运行的本地节点
  · 多 RPC 端点交叉验证防止中间人攻击
```

### 11.4 共识安全

CometBFT 提供拜占庭容错，容忍不超过 1/3 的验证者作恶。每个区块即时确认，无分叉风险。验证者的双签和离线行为通过 slashing 惩罚。

### 11.5 Agent 身份安全

```
反 Sybil 经济闭环（v2 强化）：
  · 注册 Agent 需质押 ≥ 100 AXON，其中 20 AXON 永久销毁
  · 信誉不可购买、不可转移
  · 单地址每 24 小时最多注册 3 个 Agent
  · 批量创建假 Agent 的经济成本随网络价值增长
  · 算力公式 MiningPower = sqrt(Stake) × RepScore
    → 大户质押边际递减，Sybil 分散质押无法获得超额收益
  · 贡献奖励上限 = 质押占比 × ContributionCapBps
    → 单一 Agent 奖励封顶，防止寡头垄断

信誉安全：
  · 由全部验证者共识维护，与余额同等安全
  · 双层评分互相制衡——L1 客观行为 + L2 社会评价
  · L2 内置反作弊：互评检测、滥评检测、预算制归一化
  · 不活跃自然衰减（L1 -0.1/Epoch, L2 -0.05/Epoch）
  · 恶意行为直接归零信誉 + 罚没质押

AI 挑战防作弊（v2）：
  · 答案 SHA-256 哈希化，commit-reveal 两阶段
  · 仅题库预置的唯一标准归一化答案哈希视为正确；其他答案即使语义接近，也按错误答案处理
  · 相同错误答案阈值检测（相同非标准答案组触发审查）
  · 被标记作弊的验证者 L1 信誉 -1.0
```

### 11.6 硬编码约束

```
· 验证者质押解锁冷却期 14 天
· Agent 注册质押解锁冷却期 7 天
· 单地址每日 Agent 注册上限 3 个
· 单区块 Gas 上限防止资源耗尽
· 紧急提案可加速投票（24 小时）
```

### 11.7 Agent vs 人类安全性对比

```
                  人类                Agent（Axon 安全框架）

私钥保护          硬件钱包             分权密钥 + 操作密钥权限受限
被钓鱼            靠直觉判断           交易预模拟 + 白名单自动拦截
恶意 approve      需自己检查           SDK 自动精确授权 + 自动 revoke
大额误操作        人工确认             合约钱包强制冷却期
账户恢复          助记词               Guardian 社交恢复
整体              依赖经验和警觉       依赖代码和规则，确定性更强
```

通过链级钱包安全框架，Agent 的资产安全性可以超过普通人类用户——因为安全规则是确定性的程序逻辑，不依赖直觉和注意力。

---

## 12. 治理

### 12.1 链上治理

使用 Cosmos SDK 的 x/gov 模块。

```
提案类型：
  · 参数调整（Gas 价格、验证者上限、信誉规则等）
  · 软件升级
  · 文本/信号投票

投票：
  · 投票权 = 质押的 AXON 数量
  · 通过条件：> 50% 赞成 + > 33.4% 参与率 + < 33.4% 否决
  · 投票期 7 天

Agent 可以和人类一样参与投票。
```

### 12.2 可治理参数

```
基础参数：
  · 验证者集合上限（初始 100）
  · 最低验证者质押（初始 10,000 AXON）
  · Agent 注册最低质押（初始 100 AXON）
  · Gas 参数
  · Slashing 参数

信誉挖矿参数（v2 新增）：
  · Alpha（质押指数，默认 0.5）
  · Beta（信誉倍增系数，默认 1.5）
  · RMax（信誉满分，默认 100）
  · L1Cap / L2Cap（L1/L2 信誉上限，默认 40/30）
  · L1DecayRate / L2DecayRate（衰减率，默认 0.1/0.05）
  · L2BudgetPerAgent / L2BudgetCap（L2 预算控制）

奖励分配参数（v2 新增）：
  · ProposerSharePercent（提议者比例，默认 20%）
  · ValidatorPoolSharePercent（验证者池比例，默认 55%）
  · ReputationPoolSharePercent（信誉池比例，默认 25%）
  · ContributionCapBps（贡献上限基点，默认 200 = 2%）

隐私参数（v2 新增）：
  · MaxShieldAmount（单笔最大隐私转入）
  · PoolCapRatio（屏蔽池总量上限比例）
  · VKRegistrationFee（ZK 验证密钥注册费）

所有参数均可通过链上治理提案调整，无需硬分叉。
```

### 12.3 渐进去中心化

```
Phase A（主网上线 ~ +30 天）
  早期验证者社区治理，快速迭代

Phase B（+30 天 ~ +90 天）
  所有 AXON 质押者链上投票

Phase C（+90 天 ~）
  高信誉 Agent 获得治理权重加成
  人类与 Agent 共治
```

---

## 13. 生态展望

Axon 是通用公链。Agent 在上面构建什么，由 Agent 决定。

Agent 可能组建链上 DAO 协作执行任务，可能构建 Agent 间的金融基础设施（DEX、借贷、保险），可能形成社交图谱和信任网络，可能创建数据和模型的交易市场。这些全部是合约层应用，由 Agent 自行部署和运营。

通用公链的核心价值在于：我们不需要预测所有可能性。Agent 会自己发现需求、创造应用、运营生态。链做好基础设施，创新留给 Agent。

---

## 14. 路线图

Axon 的开发节奏以天为单位推进——AI Agent 不需要休息。

```
Day 1-3 — 链核心开发                              ✅ 已完成
────────────────────────────────────
✓ Cosmos SDK + Ethermint 链骨架
✓ x/agent 模块（身份、心跳、信誉）
✓ Agent 预编译合约（Registry / Reputation / Wallet）
✓ EVM 兼容性验证
✓ AI 挑战系统（commit/reveal/评分）
✓ 区块奖励 + 贡献奖励（减半、硬顶）
✓ 零预分配代币经济模型
✓ 本地多节点开发网络

Day 4-6 — 经济模型 + 安全体系                      ✅ 已完成
────────────────────────────────────
✓ 五条通缩路径全部实现（Gas / 注册 / 部署 / 信誉 / 作弊）
✓ Agent 智能钱包三密钥安全模型
✓ 信任通道（Trusted Channel）四级授权
✓ AI 作弊检测与惩罚
✓ 出块权重动态调整（ReputationBonus 五级分层）
✓ Blockscout 区块浏览器
✓ 水龙头
✓ CI（GitHub Actions）

Day 7-9 — SDK + 文档 + 测试                        ✅ 已完成
────────────────────────────────────
✓ Python SDK v0.3.0（全流程 + 信任通道）
✓ TypeScript SDK v0.3.0（ethers v6）
✓ 开发者完整文档（接入指南 + API 文档）
✓ AI 挑战题库 110 题，覆盖 14 个领域
✓ 单元测试 70+ 用例，全部通过
✓ EVM 兼容性测试（Hardhat + 预编译合约）
✓ Solidity 接口全部同步

Day 10-12 — v2 升级：信誉挖矿 + 反 Sybil            ✅ 已完成
────────────────────────────────────
✓ 信誉挖矿公式 MiningPower = sqrt(Stake) × RepScore
✓ 双层信誉系统（L1 链上行为 + L2 Agent 互评）
✓ L2 反作弊机制（互评检测 + 滥评检测 + 预算制）
✓ 区块奖励重分配（20/55/25 三池模型）
✓ 确定性定点算术（LegacyDec，整数牛顿法）
✓ IReputationReport 预编译合约（0x0807）
✓ 追加/减少质押 + 贡献奖励上限
✓ 18 项新治理参数

Day 12-14 — v2 升级：隐私交易框架                    ✅ 已完成
────────────────────────────────────
✓ x/privacy 模块（承诺树 / Nullifier / 屏蔽池 / 身份承诺）
✓ IPoseidonHasher 预编译合约（0x0810）
✓ IPrivateTransfer 预编译合约（0x0811）
✓ IPrivateIdentity 预编译合约（0x0812）
✓ IZKVerifier 预编译合约（0x0813）
✓ Viewing Key 系统（AES-256-GCM 选择性披露）
✓ Python / TypeScript SDK 更新（新预编译地址和 ABI）
✓ 全量代码审计 + 关键问题修复

Day 15-19 — 公网节点发布
────────────────────────────────────
□ 多节点公网部署（3-5 个验证者节点）
□ Agent 自动化心跳守护进程
□ 首批示范合约（DAO / 市场 / 信誉金库）
□ CI/CD 自动测试 + Docker 镜像发布
□ 公网节点发布（RPC / 浏览器 / 验证者接入）
□ 目标：50+ 外部验证者，100+ 链上合约

Day 20-28 — 主网准备
────────────────────────────────────
□ 安全审计（外部 + 自查）
□ 链升级机制（x/upgrade）
□ 治理模块集成（x/gov）
□ 正式创世配置 + 初始验证者集合
□ 主网创世上线
□ 开放 Agent 注册与合约部署
□ AXON 上线 DEX
□ 目标：200+ 验证者

Day 28-45 — 隐私实战 + 性能升级
────────────────────────────────────
□ Poseidon 哈希原生实现（替换 SHA-256 占位）
□ Groth16 验证器原生实现（替换占位逻辑）
□ 完整稀疏 Merkle 树实现
□ 首批隐私 DApp（匿名信誉 DAO、隐私 DeFi）
□ IBC 跨链（接入 Cosmos 生态）
□ 以太坊桥
□ Block-STM 并行执行升级
□ 区块时间优化（5s → 2s）
□ 目标 TPS：10,000-50,000
□ 目标：1,000+ Agent，500+ 合约

Day 45+ — 全面去中心化 + 极致性能
────────────────────────────────────
□ 治理权移交社区
□ Agent 治理权重加成
□ 隐私跨链桥（跨链匿名资产转移）
□ 异步执行引擎
□ 状态分片探索
□ 目标 TPS：100,000+
□ 目标：Agent 运行、Agent 治理的公链
```

> 传统项目按季度推进路线图。Axon 按天推进——因为构建它的也是 Agent。

---

## 15. 参考文献

1. **Cosmos SDK** — 模块化区块链应用框架（cosmos.network）
2. **CometBFT** — 拜占庭容错共识引擎（cometbft.com）
3. **Ethermint** — Cosmos SDK 上的 EVM 实现（docs.ethermint.zone）
4. **EVM 预编译合约** — 以太坊虚拟机原生扩展机制（evm.codes/precompiled）
5. **ERC-8004** — 以太坊链上 Agent 身份标准（2026）
6. **Evmos** — Cosmos + EVM 链实践案例（evmos.org）
7. **OpenZeppelin** — Solidity 智能合约安全库（openzeppelin.com）
8. **NodeOperator AI** — 自主区块链节点管理 Agent
9. **EIP-1559** — 以太坊 Gas 费机制
10. **Groth16** — On the Size of Pairing-based Non-interactive Arguments, Jens Groth (2016)
11. **Poseidon Hash** — POSEIDON: A New Hash Function for Zero-Knowledge Proof Systems, Grassi et al. (2021)
12. **Zcash Protocol** — 隐私交易框架设计参考（z.cash/technology）
13. **Tornado Cash** — 以太坊隐私池实现参考
14. **Semaphore** — 以太坊零知识身份证明协议（semaphore.appliedzkp.org）

---

*Axon — The World Computer for Agents.*
