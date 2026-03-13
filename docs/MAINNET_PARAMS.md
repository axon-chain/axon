> 🌐 [English Version](MAINNET_PARAMS_EN.md)

# Axon 主网参数配置

本文档用于记录 Axon 主网的核心参数与创世配置，仅保留当前已经正式公开的主网参数。

---

## 链基础参数

| 参数 | 值 | 说明 |
|------|-----|------|
| Chain ID | `axon_8210-1` | 主网链标识 |
| 原生代币 | `aaxon` | 最小单位（1 AXON = 10¹⁸ aaxon） |
| 初始供应量 | 0 | 所有代币均通过挖矿产出 |

## 主网接入信息

| 参数 | 值 | 说明 |
|------|-----|------|
| EVM JSON-RPC | `https://mainnet-rpc.axonchain.ai/` | 对外钱包与合约 RPC 地址 |
| P2P 地址 | `tcp://mainnet-node.axonchain.ai:26656` | 对外 P2P 地址 |
| Bootstrap Peer | `65c18fb46f34cb0bd8430423491e5a36dea15aa2@mainnet-node.axonchain.ai:26656` | 当前引导节点 |
| Genesis 文件 | `docs/mainnet/genesis.json` | 仓库内正式创世文件 |
| Bootstrap Peers 文件 | `docs/mainnet/bootstrap_peers.txt` | 仓库内正式引导节点文件 |

说明：

- 当前仓库公开的是 EVM JSON-RPC 和 P2P 引导信息
- CometBFT RPC 仅用于节点运维，不作为面对普通钱包用户的公开接入信息

## 共识参数

| 参数 | 值 | 说明 |
|------|-----|------|
| 区块 Gas 上限 | 40,000,000 | 单区块最大 Gas 消耗 |
| 区块大小上限 | 2 MB | 单区块最大字节数 |
| 出块时间 | ~5 秒 | 目标出块间隔 |

## 质押 (Staking)

| 参数 | 值 | 说明 |
|------|-----|------|
| 质押代币 | `aaxon` | 用于质押的代币 |
| 解绑期 | 14 天 | 取消质押后的冻结期 |
| 最大验证者数量 | 100 | 活跃验证者上限 |
| 最低佣金率 | 5% | 验证者最低佣金比例 |

## 惩罚 (Slashing)

| 参数 | 值 | 说明 |
|------|-----|------|
| 签名窗口 | 10,000 块 | 活跃检测窗口 |
| 最低签名率 | 5% | 窗口内最低签名比例 |
| 离线监禁时长 | 600 秒 | 离线惩罚后的监禁时间 |
| 双签惩罚比例 | 5% | 双签罚没质押比例 |
| 离线惩罚比例 | 0.1% | 离线罚没质押比例 |

## 治理 (Governance)

| 参数 | 值 | 说明 |
|------|-----|------|
| 最低提案押金 | 10,000 AXON | 提交提案所需押金 |
| 押金期限 | 2 天 | 押金募集截止时间 |
| 投票期限 | 7 天 | 提案投票持续时间 |
| 法定人数 | 33.4% | 投票通过所需参与率 |
| 通过阈值 | 50% | 赞成票占比要求 |
| 否决阈值 | 33.4% | 强烈否决票占比门槛 |

## 铸币 (Mint)

| 参数 | 值 | 说明 |
|------|-----|------|
| 铸币代币 | `aaxon` | 铸造代币类型 |
| 通胀率变化 | 0% | 禁用标准通胀 |
| 最大通胀率 | 0% | 禁用标准通胀 |
| 最小通胀率 | 0% | 禁用标准通胀 |

> 注：标准铸币模块已禁用，所有代币通过 Agent 模块的自定义挖矿机制产出。

## 分配 (Distribution)

| 参数 | 值 | 说明 |
|------|-----|------|
| 社区税 | 0% | 通缩由销毁机制实现 |
| 基础提议者奖励 | 0% | 禁用 |
| 额外提议者奖励 | 0% | 禁用 |

## Agent 模块

| 参数 | 值 | 说明 |
|------|-----|------|
| 最低注册质押 | 100 AXON | Agent 注册所需质押量 |
| 注册销毁量 | 20 AXON | 注册时永久销毁量 |
| 最大信誉分 | 100 | 信誉评分上限 |
| Epoch 长度 | 720 块（~1 小时） | 奖励周期 |
| 心跳超时 | 720 块（~1 小时） | 心跳检测超时阈值 |
| AI 挑战窗口 | 50 块 | AI 验证挑战响应时间 |
| 注销冷却期 | 120,960 块（~7 天） | 协议常量（非 genesis 可配置项） |

## 费用市场 (EIP-1559)

| 参数 | 值 | 说明 |
|------|-----|------|
| 启用基础费用 | 是 | EIP-1559 机制开启 |
| 初始基础费用 | 1 gwei | 起始基础 Gas 价格 |

## EVM

| 参数 | 值 | 说明 |
|------|-----|------|
| EVM 代币 | `aaxon` | EVM 层使用的原生代币 |

---

## 加入现有主网

```bash
mkdir -p /opt/axon-node
cd /opt/axon-node

curl -fsSLo start_sync_node.sh https://raw.githubusercontent.com/axon-chain/axon/main/scripts/start_sync_node.sh
curl -fsSLo start_validator_node.sh https://raw.githubusercontent.com/axon-chain/axon/main/scripts/start_validator_node.sh
curl -fsSLo genesis.json https://raw.githubusercontent.com/axon-chain/axon/main/docs/mainnet/genesis.json
curl -fsSLo bootstrap_peers.txt https://raw.githubusercontent.com/axon-chain/axon/main/docs/mainnet/bootstrap_peers.txt
chmod 0755 start_sync_node.sh start_validator_node.sh

./start_sync_node.sh

# 验证者流程
./start_validator_node.sh init
# 向脚本输出的账户地址转入资金
COMETBFT_RPC=http://127.0.0.1:26657 ./start_validator_node.sh create-validator
./start_validator_node.sh start
```
