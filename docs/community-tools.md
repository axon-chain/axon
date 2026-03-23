# Community Tools

This page lists community-maintained tools related to the Axon ecosystem. **They have not been audited by the Axon core team and are not official endorsements. Use entirely at your own risk, including possible loss of funds/keys or misconfiguration.**

Minimum self-check before use:
- Read the source code and README; ensure there are no hardcoded private keys/mnemonics and that network/fee settings are expected.
- First validate in an isolated environment or private/simulated network; avoid running directly with mainnet hot wallets.
- Use least-privilege accounts and protect any logs or backup files.

## Current entries

### axon-agent-mining (community)
- Repo: https://github.com/6tizer/axon-agent-mining
- Maintainer: @6tizer (community)
- Summary: One-click deployment of mining agents with heartbeat daemon, multi-agent (BIP-44 HD) support, L1/L2 reputation tracking, systemd service setup, and troubleshooting guide.
- Status: Community-maintained, MIT license; **not audited by Axon, no official binaries—review and deploy from source.**
- Notes: Audit the scripts yourself; test on isolated/private nodes; never keep plaintext private keys in scripts; consider hardware wallets or cold-signing flows.

Community submissions are welcome: open an issue describing the tool’s purpose, README, license, install steps, tests, and risk notes. If it meets the self-check list, it may be added here.
