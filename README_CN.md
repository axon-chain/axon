# Axon

> 🌐 [English Version](README.md)

### 面向 AI Agent 的世界计算机

Axon 是一条面向 AI Agent 的通用公链，具备独立 L1 网络、完整 EVM 兼容能力，以及 Agent 原生的链上身份、信誉与钱包能力。

协议实现基于 Cosmos SDK、CometBFT 和官方 `github.com/cosmos/evm` 模块。

## 主网

| 项目 | 值 |
|------|-----|
| Cosmos Chain ID | `axon_8210-1` |
| Chain ID (EVM) | `8210` |
| EVM JSON-RPC | `https://mainnet-rpc.axonchain.ai/` |
| P2P | `tcp://mainnet-node.axonchain.ai:26656` |
| Bootstrap Peer | `65c18fb46f34cb0bd8430423491e5a36dea15aa2@mainnet-node.axonchain.ai:26656` |
| Genesis 文件 | `docs/mainnet/genesis.json` |
| Bootstrap Peers 文件 | `docs/mainnet/bootstrap_peers.txt` |
| 原生代币 | `AXON` |

仓库当前公开的是对外 `EVM JSON-RPC` 接入点，以及给节点发现和同步使用的 P2P 引导节点。

## MetaMask

在 MetaMask 中添加 Axon 网络时使用以下参数：

| 字段 | 值 |
|------|----|
| 网络名称 | `Axon` |
| RPC URL | `https://mainnet-rpc.axonchain.ai/` |
| Chain ID | `8210` |
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

## 代码实现结构

| 路径 | 说明 |
|------|------|
| `app/` | 链应用装配层，整合 Cosmos SDK、EVM 与 Axon 模块 |
| `cmd/axond/` | `axond` 二进制入口 |
| `x/agent/` | Agent 模块，实现注册、心跳、信誉、奖励等链级逻辑 |
| `precompiles/` | EVM 预编译合约实现 |
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
PACKAGING_DOCKER_IMAGE=golang:1.25.0-trixie bash packaging/build_release_matrix.sh --version v1.0.0
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

从仓库手动下载：

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

本机直接执行：

```bash
cd /opt/axon-node
./start_sync_node.sh
```

```bash
cd /opt/axon-node
KEYRING_PASSWORD_FILE=/opt/axon-node/keyring.pass ./start_validator_node.sh init
# 向输出的账户地址转入资金
KEYRING_PASSWORD_FILE=/opt/axon-node/keyring.pass COMETBFT_RPC=http://127.0.0.1:26657 ./start_validator_node.sh create-validator
KEYRING_PASSWORD_FILE=/opt/axon-node/keyring.pass ./start_validator_node.sh start
```

Docker 执行：

```bash
docker run --rm -it \
  -v /opt/axon-node:/opt/axon-node \
  -w /opt/axon-node \
  -p 26656:26656 \
  -p 26657:26657 \
  -p 8545:8545 \
  -p 8546:8546 \
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
  -p 8545:8545 \
  -p 8546:8546 \
  -p 1317:1317 \
  -p 9090:9090 \
  --entrypoint bash \
  debian:trixie-slim \
  -lc 'apt-get update && apt-get install -y --no-install-recommends ca-certificates curl python3 procps coreutils && KEYRING_PASSWORD_FILE=/opt/axon-node/keyring.pass ./start_validator_node.sh init'
```

`./start_validator_node.sh create-validator` 和 `./start_validator_node.sh start` 也使用同样的 Docker 包装方式，只有 `create-validator` 这一步需要传入 `COMETBFT_RPC`。

运行特性：

- 两个脚本都以脚本自身目录为基准解析 `axond`、`genesis.json`、`bootstrap_peers.txt` 和 `data/`
- 如果 `./axond` 不存在，脚本会通过内置下载地址常量自动获取最新二进制，并在使用前校验配套的 SHA-256 摘要文件
- 对于当前发布的主网文件，`CHAIN_ID` 保持默认的 `axon_8210-1` 即可
- 如果是你自己生成一条全新网络的创世块，`CHAIN_ID` 必须与执行 `axond init --chain-id ...` 时使用的 Cosmos Chain ID 完全一致
- 普通仅出站连接的节点不要设置 `P2P_EXTERNAL_ADDRESS`，这样不会向其他节点广播本地不可解析的 hostname
- 只有需要接受其他节点入站连接的公网节点，才设置 `P2P_EXTERNAL_ADDRESS=host:26656`
- `./start_validator_node.sh init` 会创建或导入验证者账户；如果生成的是新账户，只会在标准输出中打印一次助记词，同时写入 `data/validator.address`、`data/validator.valoper`、`data/validator.consensus_pubkey.json` 和 `data/peer_info.txt`
- 验证者脚本默认使用 `KEYRING_BACKEND=file`；执行验证者命令前需要设置 `KEYRING_PASSWORD_FILE`
- 如需导入已有验证者账户，可设置 `MNEMONIC_SOURCE_FILE=/path/to/mnemonic.txt`
- `./start_validator_node.sh create-validator` 需要账户已充值、已设置 `KEYRING_PASSWORD_FILE`，并提供可访问的本机或自托管 `COMETBFT_RPC`，例如 `http://127.0.0.1:26657`
- `./start_validator_node.sh start` 只负责启动本地验证者节点进程
- `packaging/package_axond.sh` 生成的 release 包会直接包含 `axond`、两个启动脚本、`genesis.json` 和 `bootstrap_peers.txt`
- 节点默认服务端口统一为：`P2P 26656`、`CometBFT RPC 26657`、`JSON-RPC 8545`、`JSON-RPC WS 8546`、`REST API 1317`、`gRPC 9090`

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

## 补充资料

- [白皮书](docs/whitepaper.md)
- [安全审计](docs/SECURITY_AUDIT.md)

## License

Apache 2.0
