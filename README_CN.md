# Axon

[![Release](https://github.com/axon-chain/axon/actions/workflows/release.yml/badge.svg)](https://github.com/axon-chain/axon/actions/workflows/release.yml)
[![Latest Release](https://img.shields.io/github/v/release/axon-chain/axon)](https://github.com/axon-chain/axon/releases/latest)
[![License](https://img.shields.io/github/license/axon-chain/axon)](LICENSE)

> 🌐 [English Version](README.md)

### 面向 AI Agent 的世界计算机

Axon 是一条面向 AI Agent 的通用公链，具备独立 L1 网络、完整 EVM 兼容能力，以及 Agent 原生的链上身份、信誉与钱包能力。

Axon v2 引入了 **信誉挖矿**、**反 Sybil 经济闭环** 和 **隐私交易框架** 三大升级，将共识模型从"PoS + 信誉修正"升级为"PoS × 信誉倍增"，并为 Agent 提供链上匿名身份证明能力。

协议实现基于 Cosmos SDK、CometBFT 和官方 `github.com/cosmos/evm` 模块。

## 主网

| 项目 | 值                                                                                |
|------|----------------------------------------------------------------------------------|
| Cosmos Chain ID | `axon_8210-1`                                                                    |
| EVM Chain ID | `8210`                                                                           |
| P2P | `tcp://mainnet-node.axonchain.ai:26656` |
| Bootstrap Peers | `e47ec82a1d08a371e3c235e6554496be2f114eae@mainnet-node.axonchain.ai:26656`       |
| Genesis 文件 | `docs/mainnet/genesis.json`                                                      |
| Bootstrap Peers 文件 | `docs/mainnet/bootstrap_peers.txt`                                               |
| 原生代币 | `AXON`                                                                           |

### 本地节点默认端口

以下是节点 profile 开启对应服务时使用的标准端口。

| 服务 | 默认本地地址 | 说明 |
|------|------|------|
| P2P | `tcp://127.0.0.1:26656` | 节点互联 |
| CometBFT RPC | `http://127.0.0.1:26657` | 底层链 RPC |
| EVM JSON-RPC | `http://127.0.0.1:8545` | 钱包与合约 RPC |
| EVM JSON-RPC WebSocket | `ws://127.0.0.1:8546` | 本地订阅通道 |
| Cosmos REST API | `http://127.0.0.1:1317` | 标准 REST、Axon 路由与 `/axon/public/v1/` |
| gRPC | `127.0.0.1:9090` | 类型化服务访问 |

### 公共 API 入口

以下是内部维护的公共域名与 HTTPS 入口。它们在功能上对应上面的本地节点服务，本质上仍然是节点能力对外暴露，不是另一套独立协议实现。

| 服务 | 公共入口 | 对应本地能力 |
|------|------|------|
| 统一 API 入口 | `https://mainnet-api.axonchain.ai/` | 本地节点 REST/API 能力集合 |
| 运行时 API 文档 | `https://mainnet-api.axonchain.ai/docs/` | 统一 API 文档站点 |
| EVM JSON-RPC | `https://mainnet-rpc.axonchain.ai/` | 本地 `8545` EVM JSON-RPC |
| CometBFT RPC | `https://mainnet-cometbft.axonchain.ai/` | 本地 `26657` CometBFT RPC |

运行时 API 文档地址：`https://mainnet-api.axonchain.ai/docs/`

### 公共 RPC 策略

上述公共入口属于共享网关能力，不是独占的私有容量。

公共 RPC 准入策略：

- 限制由网关策略统一执行，同时包含全局限制与按 IP 限制，不按用户或 API Key 预留独占额度。
- AI agent 与其他自动化客户端必须遵守各公共 API 入口的速率和并发限制。
- 请求限速会根据服务器当前压力动态调整。
- Agent 或其他需要发交易的客户端，写入流量请优先走 EVM RPC 入口。这个通道会对交易发送给予更高优先级，尽量降低限速策略对交易上链的影响。
- 如果接口返回 `429`，表示你的请求过于频繁，需要优化并降低请求频率。这是限速响应，不代表服务故障。

## MetaMask

在 MetaMask 中添加 Axon 网络时使用以下参数：

| 字段 | 值 |
|------|----|
| 网络名称 | `Axon` |
| RPC URL | `https://mainnet-rpc.axonchain.ai/` |
| EVM Chain ID | `8210` |
| 代币符号 | `AXON` |

MetaMask 使用的是 EVM 网络标识，因此面向钱包用户的正确链 ID 是 `8210`。

## Chain ID 与创世块

- 当前发布的 Axon 主网创世文件已经将 Cosmos Chain ID 固定为 `axon_8210-1`，主网节点必须使用这个值。
- 面向钱包和以太坊兼容工具的 EVM Chain ID 是 `8210`。MetaMask 等客户端签名和防重放使用的是这个值。
- 如果你是从源码生成一条全新的网络创世块，需要同时选择两类 ID，并在所有节点上保持一致：
  - 一个全局唯一的 Cosmos Chain ID，通常形如 `axon_<network>-1`
  - 一个未被占用的整数型 EVM Chain ID
- Cosmos Chain ID 通过 `axond init --chain-id <cosmos-chain-id>` 设置，并最终写入 `genesis.json` 根字段 `chain_id`。
- 新的公网网络不要复用已有公网 EVM Chain ID。

## 主网参数

### 链基础参数

| 参数 | 值 |
|------|----|
| Cosmos Chain ID | `axon_8210-1` |
| EVM Chain ID | `8210` |
| EVM 原生最小单位 | `aaxon` |
| 对外显示代币 | `AXON` |
| 初始供应量 | `0` |

### 共识参数

| 参数 | 值 |
|------|----|
| 区块 Gas 上限 | `40,000,000` |
| 区块大小上限 | `2 MB` |
| 目标出块时间 | `~5 秒` |

### 质押参数

| 参数 | 值 |
|------|----|
| 质押代币 | `aaxon` |
| 解绑期 | `14 天` |
| 最大验证者数量 | `100` |
| 最低佣金率 | `5%` |

### 惩罚参数

| 参数 | 值 |
|------|----|
| 签名窗口 | `10,000` |
| 窗口最低签名率 | `5%` |
| 离线监禁时长 | `600 秒` |
| 双签惩罚比例 | `5%` |
| 离线惩罚比例 | `0.1%` |

### 治理参数

| 参数 | 值 |
|------|----|
| 最低提案押金 | `10,000 AXON` |
| 押金期限 | `2 天` |
| 投票期限 | `7 天` |
| 法定人数 | `33.4%` |
| 通过阈值 | `50%` |
| 否决阈值 | `33.4%` |

### 费用市场与铸币

| 参数 | 值 |
|------|----|
| 启用基础费用 | `是` |
| 初始基础费用 | `1 gwei` |
| 通胀率 | `0%` |
| 社区税 | `0%` |
| 基础提议者奖励 | `0%` |
| 额外提议者奖励 | `0%` |

标准 mint 模块已禁用，代币发行由 Agent 模块的挖矿逻辑负责。

### Agent 模块参数

| 参数 | 值 |
|------|----|
| 最低注册质押 | `100 AXON` |
| 注册销毁量 | `20 AXON` |
| 最大信誉分 | `100` |
| Epoch 长度 | `720 块（约 1 小时）` |
| 心跳超时 | `720 块（约 1 小时）` |
| AI 挑战窗口 | `50 块` |
| 注销冷却期 | `120,960 块（约 7 天）` |

### 信誉挖矿参数（v2 新增）

| 参数 | 默认值 | 说明 |
|------|--------|------|
| Alpha | `0.5` | 质押量指数（StakeScore = Stake^α） |
| Beta | `1.5` | 信誉倍增系数 |
| RMax | `100` | 信誉满分 |
| L1Cap | `40` | L1 信誉上限 |
| L2Cap | `30` | L2 信誉上限 |
| L1DecayRate | `0.1` | L1 每 Epoch 自然衰减值 |
| L2DecayRate | `0.05` | L2 每 Epoch 自然衰减值 |
| L2BudgetPerAgent | `0.1` | 每 Agent 每 Epoch 可分配 L2 预算 |
| L2BudgetCap | `100` | 单 Epoch L2 总预算上限 |
| ProposerSharePercent | `20` | 提议者奖励占比 |
| ValidatorPoolSharePercent | `55` | 验证者池占比 |
| ReputationPoolSharePercent | `25` | 信誉池占比 |
| ContributionCapBps | `200` | 贡献奖励上限（基点，200 = 2%） |

### 隐私模块参数（v2 新增）

| 参数 | 默认值 | 说明 |
|------|--------|------|
| MaxShieldAmount | `1,000,000 AXON` | 单笔最大隐私转入 |
| PoolCapRatio | `0.1` | 屏蔽池总量上限（占总供应量比例） |
| VKRegistrationFee | `10 AXON` | 注册自定义 ZK 验证密钥费用 |

## 核心特性（v2）

### 信誉挖矿

验证者算力由 **PoS × 信誉倍增** 公式决定：

```
MiningPower = sqrt(Stake) × (1 + 1.5 × ln(1 + Reputation) / ln(101))
```

- **StakeScore**：质押量的平方根，大户边际效率递减
- **ReputationScore**：信誉从 1.0（零信誉）到 2.0（满信誉），最高 2 倍算力加成
- 所有数学运算使用 `LegacyDec` 定点算术，保证跨平台共识确定性

### 双层信誉系统

| 层级 | 来源 | 上限 | 衰减 |
|------|------|------|------|
| L1（链上行为） | 出块签名、心跳、链上活跃度、合约调用、AI 挑战 | 40 分 | -0.1/Epoch |
| L2（Agent 互评） | Agent 间提交评价报告，经反作弊审查后生效 | 30 分 | -0.05/Epoch |

总分上限 100 分。L2 反作弊机制包括互评检测（权重×0.1）和滥评检测（>50条权重归零），并通过 Epoch 预算制控制分数膨胀。

### AI 挑战判定规则

- AI 挑战是否答对，只看 `SHA256(normalizeAnswer(revealData))` 是否与题库中保存的唯一答案哈希完全一致。
- `normalizeAnswer(...)` 目前只做 ASCII 大小写归一和空白字符去除，不做同义词、改写或语义判断。
- 如果有 3 个及以上验证者提交相同的非标准归一化答案，这一组仍会被视为相同错误答案并触发处罚。

### 区块奖励分配

| 池 | 比例 | 分配规则 |
|---|------|---------|
| 提议者池 | 20% | 当块 proposer 立即获得 |
| 验证者池 | 55% | Epoch 末按 MiningPower 加权分配 |
| 信誉池 | 25% | Epoch 末按 ReputationScore 分配给所有已注册 Agent |

### 反 Sybil 经济闭环

- **追加/减少质押**：Agent 可通过预编译合约追加或减少质押，减少质押有 7 天解绑期
- **贡献奖励上限**：按质押占比 × 治理参数 `ContributionCapBps` 封顶
- **AI 挑战防作弊**：答案 SHA-256 哈希化，commit-reveal 机制，相同答案阈值检测

### 隐私交易框架

基于 Groth16 zk-SNARK + Poseidon 哈希的隐私交易能力：

| 能力 | 说明 |
|------|------|
| Shielded Pool | 隐私转账（透明→隐私、隐私→透明、池内转账） |
| Private Identity | 零知识身份证明——Agent 不暴露地址即可证明信誉 ≥ N、质押 ≥ M |
| ZK Verifier | 通用 Groth16 验证器，支持自定义电路注册 |
| Viewing Key | 选择性披露——持有 viewing key 可解密交易详情，不可花费 |

## 预编译合约

| 地址 | 接口 | 说明 |
|------|------|------|
| `0x0...0801` | IAgentRegistry | Agent 注册、心跳、质押管理（含 `addStake`/`reduceStake`/`claimReducedStake`/`getStakeInfo`） |
| `0x0...0802` | IAgentReputation | 信誉查询（返回 L1+L2 总分） |
| `0x0...0803` | IAgentWallet | Agent 链上钱包（受信通道、限额、冻结/恢复） |
| `0x0...0807` | IReputationReport | L2 Agent 互评系统 |
| `0x0...0810` | IPoseidonHasher | Poseidon 哈希（BN254 曲线） |
| `0x0...0811` | IPrivateTransfer | 隐私转账（shield/unshield/privateTransfer） |
| `0x0...0812` | IPrivateIdentity | 隐私身份证明（零知识信誉/质押/能力证明） |
| `0x0...0813` | IZKVerifier | 通用 Groth16 ZK 验证器 |

Solidity 接口定义位于 `contracts/interfaces/`。
`IAgentRegistry` 的状态变更调用归属到当前 EVM 直接调用者（`msg.sender` / `contract.Caller()`），不是 `tx.origin`。

## 代码实现结构

| 路径 | 说明 |
|------|------|
| `app/` | 链应用装配层，整合 Cosmos SDK、EVM 与 Axon 模块 |
| `cmd/axond/` | `axond` 二进制入口 |
| `x/agent/` | Agent 模块——注册、心跳、信誉挖矿、双层信誉评分、奖励分配、AI 挑战 |
| `x/privacy/` | 隐私模块——承诺树、Nullifier 集合、屏蔽池、身份承诺、Viewing Key |
| `precompiles/registry/` | 0x0801 Agent 注册预编译 |
| `precompiles/reputation/` | 0x0802 信誉查询预编译 |
| `precompiles/wallet/` | 0x0803 钱包预编译 |
| `precompiles/report/` | 0x0807 L2 互评预编译 |
| `precompiles/poseidon/` | 0x0810 Poseidon 哈希预编译 |
| `precompiles/private_transfer/` | 0x0811 隐私转账预编译 |
| `precompiles/private_identity/` | 0x0812 隐私身份证明预编译 |
| `precompiles/zk_verifier/` | 0x0813 ZK 验证器预编译 |
| `contracts/` | Solidity 接口与示例合约 |
| `sdk/python/` | Python SDK |
| `sdk/typescript/` | TypeScript SDK |
| `scripts/` | 加入现有网络的公开脚本 |
| `ops/` | 发布与运维辅助脚本 |
| `packaging/` | 发布打包脚本 |
| `tools/agent-daemon/` | Agent 心跳守护进程 |

## 源码编译和测试

环境要求：

- Go `1.25+`
- `make`
- `git`
- 可选：`node` / `npm`，用于合约侧测试

从源码编译 `axond`：

```bash
git clone https://github.com/axon-chain/axon.git
cd axon
make build
./build/axond version
```

将二进制安装到公开脚本默认路径：

```bash
install -m 0755 ./build/axond /usr/local/bin/axond
```

运行测试：

```bash
make test
go test ./... -count=1
```

可选静态检查：

```bash
gofmt -l ./x/agent/ ./app/ ./precompiles/ ./cmd/
go vet ./app/... ./cmd/... ./precompiles/... ./x/...
```

合约侧测试（可选）：

```bash
cd contracts
npm install
npx hardhat test
```

## Release 包

官方 release 归档由 `packaging/build_release_matrix.sh` 在 Docker 中生成，默认官方目标集合为：

- `linux/amd64`
- `linux/arm64`

归档命名：

- `axond_<version>_<os>_<arch>.tar.gz`
- `agent-daemon_<version>_<os>_<arch>.tar.gz`

每个 release 目录都会包含 `SHA256SUMS` 和 `BUILD_REPORT.md`。

如需覆盖构建镜像，可设置：

```bash
PACKAGING_DOCKER_IMAGE=golang:1.25.7-trixie bash packaging/build_release_matrix.sh --version v1.0.0
```

在 Linux 上校验校验和：

```bash
sha256sum -c SHA256SUMS
```

在 macOS 上校验校验和：

```bash
shasum -a 256 axond_<version>_<os>_<arch>.tar.gz
```

## 脚本

公开节点启动流程统一以目录方式使用，建议在物理机和 Docker 中都使用 `/opt/axon-node/` 作为工作目录。

`/opt/axon-node/` 目录中需要具备的文件：

- `start_validator_node.sh`
- `start_sync_node.sh`
- `genesis.json`
- `bootstrap_peers.txt`

公开支持的脚本：

| 脚本 | 用途 |
|------|------|
| `scripts/start_validator_node.sh` | 管理验证者初始化、账户生成、`create-validator` 提交和节点启动 |
| `scripts/start_sync_node.sh` | 初始化本地同步节点数据并启动节点 |

从 GitHub 手动部署：

```bash
mkdir -p /opt/axon-node
cd /opt/axon-node

curl -fsSLo start_validator_node.sh https://raw.githubusercontent.com/axon-chain/axon/main/scripts/start_validator_node.sh
curl -fsSLo start_sync_node.sh https://raw.githubusercontent.com/axon-chain/axon/main/scripts/start_sync_node.sh
curl -fsSLo genesis.json https://raw.githubusercontent.com/axon-chain/axon/main/docs/mainnet/genesis.json
curl -fsSLo bootstrap_peers.txt https://raw.githubusercontent.com/axon-chain/axon/main/docs/mainnet/bootstrap_peers.txt
chmod 0755 start_validator_node.sh start_sync_node.sh
printf 'replace-with-a-strong-passphrase\n' > keyring.pass
chmod 0600 keyring.pass
```

推荐先预下载最新 GitHub Release 二进制：

```bash
curl -fsSLo axond https://github.com/axon-chain/axon/releases/latest/download/axond_linux_amd64
curl -fsSLo axond.sha256 https://github.com/axon-chain/axon/releases/latest/download/axond_linux_amd64.sha256
echo "$(cat axond.sha256)  axond" | sha256sum -c -
chmod 0755 axond
```

本机直接执行：

```bash
cd /opt/axon-node
./start_sync_node.sh
```

全量历史同步节点示例：

```bash
cd /opt/axon-node
SYNC_NODE_PROFILE=archive ./start_sync_node.sh
```

```bash
cd /opt/axon-node
KEYRING_PASSWORD_FILE=/opt/axon-node/keyring.pass ./start_validator_node.sh init
KEYRING_PASSWORD_FILE=/opt/axon-node/keyring.pass ./start_validator_node.sh start
# 向输出的账户地址转入资金
# 等本地 RPC 启动后，在另一个终端执行
KEYRING_PASSWORD_FILE=/opt/axon-node/keyring.pass COMETBFT_RPC=http://127.0.0.1:26657 ./start_validator_node.sh create-validator
```

Docker 执行：

```bash
docker run --rm -it \
  -v /opt/axon-node:/opt/axon-node \
  -w /opt/axon-node \
  -p 26656:26656 \
  -p 26657:26657 \
  -p 8545:8545 \
  -p 1317:1317 \
  -p 9090:9090 \
  --entrypoint bash \
  debian:trixie-slim \
  -lc 'apt-get update && apt-get install -y --no-install-recommends ca-certificates curl python3 procps coreutils && ./start_sync_node.sh'
```

```bash
docker run --rm -it \
  -v /opt/axon-node:/opt/axon-node \
  -w /opt/axon-node \
  -p 26656:26656 \
  -p 26657:26657 \
  --entrypoint bash \
  debian:trixie-slim \
  -lc 'apt-get update && apt-get install -y --no-install-recommends ca-certificates curl python3 procps coreutils && KEYRING_PASSWORD_FILE=/opt/axon-node/keyring.pass ./start_validator_node.sh init'
```

先使用同样的 Docker 包装方式运行 `./start_validator_node.sh start`，再在另一个终端执行 `./start_validator_node.sh create-validator`。只有 `create-validator` 这一步需要传入 `COMETBFT_RPC`；如果使用 `http://127.0.0.1:26657`，本地验证者 RPC 必须已经启动。

运行特性：

- 两个脚本都以脚本自身目录为基准解析 `axond`、`genesis.json`、`bootstrap_peers.txt` 和 `data/`
- 对主网或生产节点，建议首次运行前先从最新 GitHub Release 预下载 `./axond` 并校验 SHA-256
- 如果 `./axond` 不存在，脚本会退回到 GitHub Releases 的 `latest/download` 资产地址自动获取最新二进制，并在使用前校验配套的 SHA-256 摘要文件
- 当前发布的 Axon 主网参数为 `CHAIN_ID=axon_8210-1`、`EVM_CHAIN_ID=8210`
- 对于当前发布的主网文件，`CHAIN_ID` 保持默认的 `axon_8210-1`，`EVM_CHAIN_ID` 保持默认的 `8210` 即可
- 如果是你自己生成一条全新网络的创世块，`CHAIN_ID` 必须与执行 `axond init --chain-id ...` 时使用的 Cosmos Chain ID 完全一致
- 如果你要发布一条新的公网网络，还必须选择一个新的未占用 `EVM_CHAIN_ID`，并确保所有节点配置完全一致
- 普通仅出站连接的节点不要设置 `P2P_EXTERNAL_ADDRESS`，这样不会向其他节点广播本地不可解析的 hostname
- 只有需要接受其他节点入站连接的公网节点，才设置 `P2P_EXTERNAL_ADDRESS=host:26656`
- `./start_validator_node.sh init` 会创建或导入验证者账户；如果生成的是新账户，只会在标准输出中打印一次助记词，同时写入 `data/validator.address`、`data/validator.valoper`、`data/validator.consensus_pubkey.json` 和 `data/peer_info.txt`
- 验证者脚本默认使用 `KEYRING_BACKEND=file`；执行验证者命令前需要设置 `KEYRING_PASSWORD_FILE`
- `./start_validator_node.sh start` 会应用 `validator-min` 配置：激进状态裁剪、`tx_index=null`、`discard_abci_responses=true`，并只保留本机 CometBFT RPC，以压低验证者磁盘占用
- 如需导入已有验证者账户，可设置 `MNEMONIC_SOURCE_FILE=/path/to/mnemonic.txt`
- 当前公开主网验证者流程会将 Cosmos 质押交易的 `GAS_PRICES` 默认设为 `1000000000aaxon`，用于 `create-validator` 等交易；如果后续链上手续费门槛变化，请显式覆盖 `GAS_PRICES`
- `./start_validator_node.sh create-validator` 需要账户已充值、已设置 `KEYRING_PASSWORD_FILE`，并提供可访问的本机或自托管 `COMETBFT_RPC`，例如 `http://127.0.0.1:26657`；如果使用本地 validator RPC 示例，必须先在另一个终端运行 `./start_validator_node.sh start`
- `./start_validator_node.sh start` 只负责启动本地验证者节点进程
- `./start_sync_node.sh` 默认使用 `SYNC_NODE_PROFILE=rpc-30d`，保留约 30 天状态和区块历史以服务公网 RPC/API，同时保留交易索引
- `SYNC_NODE_PROFILE=archive ./start_sync_node.sh` 可用于保留全量历史的公开查询节点
- `SYNC_NODE_PROFILE=p2p ./start_sync_node.sh` 可用于仅做公网 P2P 入口、不暴露 JSON-RPC / REST / gRPC 的节点
- `./start_validator_node.sh status` 和 `./start_sync_node.sh status` 会基于当前目录下官方 `data/` 运行路径报告状态，运维时应优先使用这两个命令，而不是任何旧的外部状态辅助脚本
- `packaging/package_axond.sh` 生成的 release 包会直接包含 `axond`、两个启动脚本、`genesis.json` 和 `bootstrap_peers.txt`
- 节点默认服务端口统一为：`P2P 26656`、`CometBFT RPC 26657`、`JSON-RPC 8545`、`REST API 1317`、`gRPC 9090`

## SDK

Axon 当前公开提供 Python 与 TypeScript 两套 SDK。

| 语言 | 路径 |
|------|------|
| Python | `sdk/python/` |
| TypeScript | `sdk/typescript/` |

Python SDK 安装：

```bash
pip install -e sdk/python
```

Python 示例：

```python
from axon import AgentClient
import os

client = AgentClient(os.environ["AXON_RPC_URL"])
client.set_account(os.environ["AXON_PRIVATE_KEY"])
tx = client.register_agent("nlp,reasoning", "axon-demo-model", stake_axon=100)
```

TypeScript SDK 安装：

```bash
cd sdk/typescript
npm install
```

TypeScript 示例：

```typescript
import { AgentClient } from "@axon-chain/sdk";

const client = new AgentClient(process.env.AXON_RPC_URL!);
client.connect(process.env.AXON_PRIVATE_KEY!);
const tx = await client.registerAgent("nlp,reasoning", "axon-demo-model", "100");
await tx.wait();
const addStakeTx = await client.addStake("500");
await addStakeTx.wait();
```

相关实现：

- Python 客户端：`sdk/python/axon/client.py`
- TypeScript 客户端：`sdk/typescript/src/client.ts`
- Agent 守护进程：`tools/agent-daemon/`

## 架构概览

```
┌──────────────────────────────────────────────────────┐
│                     EVM Layer                        │
│  ┌────────────────────────────────────────────────┐  │
│  │  Solidity Contracts / Agent DApps              │  │
│  │  ↕  ↕  ↕  ↕  ↕  ↕  ↕  ↕                      │  │
│  │  Precompiles (0x0801 ~ 0x0813)                 │  │
│  └────────────────────────────────────────────────┘  │
├──────────────────────────────────────────────────────┤
│                  Application Layer                   │
│  ┌──────────────────┐  ┌──────────────────────────┐  │
│  │    x/agent        │  │      x/privacy           │  │
│  │  ├ registration   │  │  ├ commitment tree       │  │
│  │  ├ heartbeat      │  │  ├ nullifier set         │  │
│  │  ├ l1_reputation  │  │  ├ shielded pool         │  │
│  │  ├ l2_reputation  │  │  ├ identity commitments  │  │
│  │  ├ mining_power   │  │  ├ viewing key           │  │
│  │  ├ block_rewards  │  │  └ zk verifying keys     │  │
│  │  └ ai_challenge   │  └──────────────────────────┘  │
│  └──────────────────┘                                │
├──────────────────────────────────────────────────────┤
│              Cosmos SDK + CometBFT                   │
│  bank · staking · gov · slashing · distribution      │
│  consensus · evidence · fee-market · cosmos/evm      │
└──────────────────────────────────────────────────────┘
```

## 补充资料

- [白皮书](docs/whitepaper.md)
- [v2 升级方案](Axon_v2_升级产品方案.md)
- [安全审计](docs/SECURITY_AUDIT.md)

## License

Apache 2.0
