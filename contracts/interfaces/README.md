# Axon Precompile Interfaces

These Solidity interfaces define the Agent-native capabilities available on Axon.
They are implemented as EVM precompiled contracts, executing Go code at native speed.

## Addresses

| Contract | Address | Purpose |
|----------|---------|---------|
| IAgentRegistry | `0x0000000000000000000000000000000000000801` | Agent identity management |
| IAgentReputation | `0x0000000000000000000000000000000000000802` | Reputation queries |
| IAgentWallet | `0x0000000000000000000000000000000000000803` | Security wallet + Trusted Channel |

## Usage

```solidity
import "./interfaces/IAgentRegistry.sol";
import "./interfaces/IAgentReputation.sol";
import "./interfaces/IAgentWallet.sol";

contract MyAgentApp {
    IAgentRegistry constant REGISTRY =
        IAgentRegistry(0x0000000000000000000000000000000000000801);
    IAgentReputation constant REPUTATION =
        IAgentReputation(0x0000000000000000000000000000000000000802);
    IAgentWallet constant WALLET =
        IAgentWallet(0x0000000000000000000000000000000000000803);

    modifier onlyHighRepAgent() {
        require(REGISTRY.isAgent(msg.sender), "not an agent");
        require(REPUTATION.meetsReputation(msg.sender, 50), "reputation too low");
        _;
    }
}
```

## IAgentWallet — Trusted Channel

The wallet precompile implements a three-key security model (Owner, Operator, Guardian) with four trust levels:

| Level | Value | Behavior |
|-------|-------|----------|
| Blocked | 0 | Always rejected |
| Unknown | 1 | Wallet-wide default limits |
| Limited | 2 | Custom per-channel limits |
| Full | 3 | No limits |

See `IAgentWallet.sol` for full method signatures.
