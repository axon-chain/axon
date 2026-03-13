> 🌐 [English Version](README_EN.md)

# Agent Heartbeat Daemon

Axon 节点的 Agent 心跳守护进程（sidecar），自动向链上注册表预编译合约发送心跳交易，保持 Agent 在线状态。

请显式传入正式发布的 EVM JSON-RPC 地址：`https://mainnet-rpc.axonchain.ai/`。

## 功能

- **自动心跳**：每隔 N 个区块（默认 100）自动发送心跳交易
- **注册检查**：启动时验证账户是否已注册为 Agent
- **优雅关闭**：支持 SIGINT / SIGTERM 信号安全退出

## 使用方法

### 命令行参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `--rpc` | （必填） | JSON-RPC 节点地址 |
| `--private-key-file` | （必填） | 包含 Agent 账户十六进制私钥的文件路径 |
| `--heartbeat-interval` | `100` | 每隔多少个区块发送一次心跳 |
| `--log-level` | `info` | 日志级别：debug, info, warn, error |

### 直接运行

```bash
go build -o agent-daemon .

./agent-daemon \
  --rpc https://mainnet-rpc.axonchain.ai/ \
  --private-key-file /path/to/private_key.txt \
  --heartbeat-interval 100
```

### Docker 运行

```bash
docker build -t agent-daemon .

docker run --rm \
  --network host \
  agent-daemon \
  --rpc https://mainnet-rpc.axonchain.ai/ \
  --private-key-file /run/secrets/agent_private_key
```

## 工作原理

1. 启动后连接 RPC 节点，获取链 ID
2. 调用注册表预编译合约 `isAgent(address)` 检查当前账户是否为已注册 Agent
3. 轮询最新区块高度，当距离上次心跳超过设定间隔时，构造并签名 `heartbeat()` 交易
4. 通过 `eth_sendRawTransaction` 发送交易，等待回执确认

## 注意事项

- 仅支持 `--private-key-file`，避免把私钥暴露在 shell 历史和进程列表中
- 私钥请妥善保管，生产环境建议通过挂载密钥文件或密钥管理服务传入
- 确保 Agent 账户有足够的余额支付 Gas 费用
- 注册表预编译合约地址：`0x0000000000000000000000000000000000000801`
