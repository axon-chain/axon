# Packaging

Release artifact packaging scripts live here.

- `package_axond.sh`: package the `axond` runtime bundle archive
- `package_agent.sh`: package the `agent-daemon` binary archive
- `package_all.sh`: build both binary archives
- `build_release_matrix.sh`: build versioned multi-platform binary archives with checksums

All source builds in `packaging/` run inside Docker by default. Override the builder image with `PACKAGING_DOCKER_IMAGE=<image>` when required.

The `axond` runtime bundle includes:

- `axond`
- `start_validator_node.sh`
- `start_sync_node.sh`
- `genesis.json`
- `bootstrap_peers.txt`
