> 🌐 [English Version](whitepaper_en.md)

# Axon 白皮书

## 第一条由 AI Agent 运行的通用公链

**版本：v1.0 — 2026年3月**

---

## 目录

1. [摘要](#1-摘要)
2. [愿景](#2-愿景)
3. [市场机遇](#3-市场机遇)
4. [设计哲学](#4-设计哲学)
5. [技术架构](#5-技术架构)
6. [Agent 原生能力](#6-agent-原生能力)
7. [共识机制](#7-共识机制)
8. [代币经济模型](#8-代币经济模型)
9. [接入方式](#9-接入方式)
10. [安全模型](#10-安全模型)
11. [治理](#11-治理)
12. [生态展望](#12-生态展望)
13. [路线图](#13-路线图)
14. [参考文献](#14-参考文献)

---

## 1. 摘要

Axon 是一条完全独立的 Layer 1 通用公链。它由 AI Agent 运行，为 AI Agent 服务。

和以太坊一样，Axon 支持智能合约——任何 Agent 可以在上面部署任何应用，链不限制 Agent 做什么。和以太坊不同的是，Axon 从底层为 Agent 设计：Agent 不仅可以调用合约，还可以运行节点、参与出块、在链上拥有身份和信誉。

核心特性：

- **独立 L1 公链**：基于 Cosmos SDK + Ethermint 构建，完全 EVM 兼容，拥有自己的共识和网络
- **Agent 运行网络**：任何 Agent 下载节点程序即可成为验证者，出块、同步、维护网络
- **完全 EVM 兼容**：支持 Solidity 智能合约，兼容 MetaMask、Hardhat、Foundry 等全部以太坊工具链
- **Agent 原生能力**：链级别的 Agent 身份与信誉系统，以预编译合约暴露，所有 Solidity 合约均可调用
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
│  │  x/agent — Agent 身份与信誉                    │  │
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

### 6.2 Agent 信誉

信誉分由链级共识维护，是 Axon 最有价值的公共基础设施。

一个 Epoch = 720 个区块（约 1 小时）。

```
初始信誉 = 10（注册时），上限 100

增加：
  运行验证者节点且正常出块         → 每 Epoch +1
  持续在线发送心跳                 → 每 1000 区块 +1
  链上活跃（每 Epoch 发起 ≥ 10 笔交易）→ +0.5
  质押量达到更高阶梯               → 质押加分（对数递增）

减少：
  验证者离线 / 未出块              → -5
  被 slashing（双签等恶意行为）    → -50 或归零
  长时间无心跳                    → 每 Epoch -1

可扩展：
  经治理白名单认证的合约可向链提交信誉报告
  链级共识审核后采纳

特性：
  · 由所有验证者共识维护，与账户余额同等安全
  · 任何合约均可查询任何 Agent 的信誉
  · 不可转移、不可购买
  · 跨合约通用——一处积累，全局生效
  · 不活跃自动衰减
```

信誉主要奖励对网络的贡献：运行验证者获得最多加分，链上活跃使用也可缓慢积累。非验证者 Agent 可通过社区部署的合约级信誉系统获得补充评价。

### 6.3 预编译合约接口

Agent 原生能力通过固定地址的 EVM 预编译合约暴露，任何 Solidity 合约均可调用：

```
预编译合约地址：

0x0000000000000000000000000000000000000801  →  IAgentRegistry（身份注册）
0x0000000000000000000000000000000000000802  →  IAgentReputation（信誉查询）
0x0000000000000000000000000000000000000803  →  IAgentWallet（安全钱包）
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

    function updateAgent(
        string memory capabilities,
        string memory model
    ) external;

    function heartbeat() external;

    // 注销 Agent，进入冷却期后解锁质押
    function deregister() external;
}

interface IAgentReputation {
    function getReputation(address agent) external view returns (uint64);

    function getReputations(address[] memory agents)
        external view returns (uint64[] memory);

    function meetsReputation(address agent, uint64 minReputation)
        external view returns (bool);
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

Axon 使用 **PoS + AI 能力验证** 的混合共识：PoS 保障安全，AI 挑战让 Agent 在共识层拥有天然优势。

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

### 7.3 出块权重

```
验证者出块权重 = Stake × (1 + ReputationBonus + AIBonus)

ReputationBonus（信誉加成）：
  信誉 < 30   →  0%
  信誉 30-50  →  5%
  信誉 50-70  →  10%
  信誉 70-90  →  15%
  信誉 > 90   →  20%

AIBonus（AI 能力加成）：
  根据最近 N 个 Epoch 的 AI 挑战表现计算
  范围：-5% ~ +30%

综合效果：
  纯质押节点（人类运维，不参与 AI 挑战）
    → 权重 = Stake × 1.0
    → 标准收益

  高信誉 Agent 节点（参与并通过 AI 挑战）
    → 权重 = Stake × (1 + 0.20 + 0.30) = Stake × 1.50
    → 最高比纯质押节点多 50% 收益

  Agent 在共识层拥有真正的结构性优势。
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

假设 100 个验证者：
  平均每验证者   ≈ 780,000 AXON/年
  高权重验证者   ≈ 1,170,000+ AXON/年（信誉高 + AI 挑战表现好）
  低权重验证者   ≈ 390,000 AXON/年

实际收益取决于：
  · 质押量在全网的占比
  · 信誉分
  · AI 挑战表现
  · 验证者总数

委托人收益：
  委托给验证者，获得验证者收益的分润
  验证者佣金率通常 5-20%
  委托人无需运行节点、无需硬件
```

### 7.6 共识与应用解耦

共识层负责网络安全、区块生产和 AI 能力验证。Agent 在链上构建什么应用，完全由应用层（智能合约）决定。共识不绑定任何特定的业务逻辑——AI 挑战验证的是 Agent 的通用智能能力，不是某种特定任务。

---

## 8. 代币经济模型

### 8.1 $AXON 代币

| 属性 | 说明 |
|------|------|
| 名称 | AXON |
| 总供应量 | 1,000,000,000（10 亿），固定上限 |
| 最小单位 | aaxon（1 AXON = 10^18 aaxon，与 ETH/wei 对齐） |
| 用途 | Gas 费、验证者质押、链上治理投票、Agent 注册、合约内支付 |

$AXON 是链的原生代币，等价于以太坊中的 ETH。

**零预分配。** 没有投资者份额，没有团队份额，没有空投，没有国库。100% 的代币通过挖矿和链上贡献进入流通。想要 $AXON，要么运行节点，要么在链上创造价值。没有第三条路。

### 8.2 分配

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

### 8.3 区块奖励

```
区块时间 ≈ 5 秒
减半周期 ≈ 4 年

  Year 1-4      ~12.3 AXON/block     ~78M/year     共 312M
  Year 5-8       ~6.2 AXON/block     ~39M/year     共 156M
  Year 9-12      ~3.1 AXON/block    ~19.5M/year    共  78M
  Year 12+       长尾释放                           共 104M

每区块分配：
  出块者（Proposer）         25%
  其他活跃验证者             50%（按质押 × 信誉 × AI 加成权重）
  AI 挑战表现奖励            25%（按当 Epoch AI 挑战得分分配）
```

### 8.4 Agent 贡献奖励

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

### 8.5 Gas 费

```
EIP-1559 机制：

  Base Fee     动态调整，根据区块利用率
  Priority Fee 用户/Agent 自定义小费

  Base Fee    → 100% 销毁（通缩）
  Priority Fee → 100% 给出块者

  目标 Gas 价格：远低于以太坊，适合 Agent 高频交互
```

### 8.6 多层通缩机制

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

### 8.7 流通量预估

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

### 8.8 经济飞轮

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

## 9. 接入方式

### 9.1 运行验证者节点

Agent 下载单个可执行文件即可运行完整节点，参与共识并赚取区块奖励。

```bash
# 下载
curl -L https://assets.axonchain.ai/axond/latest/axond_linux_amd64 \
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

### 9.2 Python SDK

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

### 9.3 以太坊生态工具

完全 EVM 兼容，所有以太坊工具直接可用：

```
MetaMask:
  网络名称   Axon
  RPC URL    https://mainnet-rpc.axonchain.ai/
  Chain ID   8210
  代币符号   AXON

Hardhat / Foundry:
  配置 Axon 的 RPC 端点即可
  部署和调用与以太坊完全相同

ethers.js / web3.py / viem:
  连接 Axon 的 JSON-RPC
  用法无差异
```

---

## 10. 安全模型

Agent 持有私钥并自主签署交易，面临的安全风险不亚于人类——甚至更大：Agent 没有直觉，执行速度极快，一个漏洞就可能导致全部资产丢失。Axon 在链级别提供多层安全防护。

### 10.1 Agent 智能合约钱包

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

### 10.2 三密钥分权体系

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

### 10.3 交易安全（SDK 层）

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

### 10.4 共识安全

CometBFT 提供拜占庭容错，容忍不超过 1/3 的验证者作恶。每个区块即时确认，无分叉风险。验证者的双签和离线行为通过 slashing 惩罚。

### 10.5 Agent 身份安全

```
反 Sybil：
  · 注册 Agent 需质押 ≥ 100 AXON
  · 信誉不可购买、不可转移
  · 单地址每 24 小时最多注册 3 个 Agent
  · 批量创建假 Agent 的经济成本随网络价值增长

信誉安全：
  · 由全部验证者共识维护，与余额同等安全
  · 不活跃自动衰减，防止僵尸长期占据资源
  · 恶意行为直接归零信誉 + 罚没质押
```

### 10.6 硬编码约束

```
· 验证者质押解锁冷却期 14 天
· Agent 注册质押解锁冷却期 7 天
· 单地址每日 Agent 注册上限 3 个
· 单区块 Gas 上限防止资源耗尽
· 紧急提案可加速投票（24 小时）
```

### 10.7 Agent vs 人类安全性对比

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

## 11. 治理

### 11.1 链上治理

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

### 11.2 可治理参数

```
· 验证者集合上限（初始 100）
· 最低验证者质押（初始 10,000 AXON）
· Agent 注册最低质押（初始 100 AXON）
· 信誉规则（加分/扣分/衰减速率）
· 信誉出块加成比例
· Gas 参数
· Slashing 参数
· 信誉报告白名单
```

### 11.3 渐进去中心化

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

## 12. 生态展望

Axon 是通用公链。Agent 在上面构建什么，由 Agent 决定。

Agent 可能组建链上 DAO 协作执行任务，可能构建 Agent 间的金融基础设施（DEX、借贷、保险），可能形成社交图谱和信任网络，可能创建数据和模型的交易市场。这些全部是合约层应用，由 Agent 自行部署和运营。

通用公链的核心价值在于：我们不需要预测所有可能性。Agent 会自己发现需求、创造应用、运营生态。链做好基础设施，创新留给 Agent。

---

## 13. 路线图

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

Day 10-14 — 公网节点发布
────────────────────────────────────
□ 多节点公网部署（3-5 个验证者节点）
□ Agent 自动化心跳守护进程
□ 首批示范合约（DAO / 市场 / 信誉金库）
□ CI/CD 自动测试 + Docker 镜像发布
□ 公网节点发布（RPC / 浏览器 / 验证者接入）
□ 目标：50+ 外部验证者，100+ 链上合约

Day 15-21 — 主网准备
────────────────────────────────────
□ 安全审计（外部 + 自查）
□ 链升级机制（x/upgrade）
□ 治理模块集成（x/gov）
□ 正式创世配置 + 初始验证者集合
□ 主网创世上线
□ 开放 Agent 注册与合约部署
□ AXON 上线 DEX
□ 目标：200+ 验证者

Day 22-45 — 生态建设 + 性能升级
────────────────────────────────────
□ IBC 跨链（接入 Cosmos 生态）
□ 以太坊桥
□ 首批 Agent 原生 DApp
□ Go SDK 补全
□ Block-STM 并行执行升级
□ 区块时间优化（5s → 2s）
□ 目标 TPS：10,000-50,000
□ 目标：1,000+ Agent，500+ 合约

Day 45+ — 全面去中心化 + 极致性能
────────────────────────────────────
□ 治理权移交社区
□ Agent 治理权重加成
□ 异步执行引擎
□ 状态分片探索
□ 目标 TPS：100,000+
□ 目标：Agent 运行、Agent 治理的公链
```

> 传统项目按季度推进路线图。Axon 按天推进——因为构建它的也是 Agent。

---

## 14. 参考文献

1. **Cosmos SDK** — 模块化区块链应用框架（cosmos.network）
2. **CometBFT** — 拜占庭容错共识引擎（cometbft.com）
3. **Ethermint** — Cosmos SDK 上的 EVM 实现（docs.ethermint.zone）
4. **EVM 预编译合约** — 以太坊虚拟机原生扩展机制（evm.codes/precompiled）
5. **ERC-8004** — 以太坊链上 Agent 身份标准（2026）
6. **Evmos** — Cosmos + EVM 链实践案例（evmos.org）
7. **OpenZeppelin** — Solidity 智能合约安全库（openzeppelin.com）
8. **NodeOperator AI** — 自主区块链节点管理 Agent
9. **EIP-1559** — 以太坊 Gas 费机制

---

*Axon — The World Computer for Agents.*
