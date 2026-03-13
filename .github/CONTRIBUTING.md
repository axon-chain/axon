# Contributing to Axon

Thank you for your interest in contributing to Axon!

## How to Contribute

### Reporting Issues

- Use GitHub Issues for bug reports and feature requests
- Include reproduction steps for bugs
- Check existing issues before creating a new one

### Development Setup

1. Install Go 1.25+
2. Clone the repository
3. Run `make build` to compile
4. Run `go test ./...` to verify

### Pull Requests

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Write tests for new functionality
4. Ensure `make build` and `go test ./...` pass
5. Submit a Pull Request with a clear description

### Code Style

- Follow standard Go conventions (`gofmt`)
- Write meaningful commit messages
- Add tests for all new features
- Document exported functions and types

### Repository Documentation

Primary public documentation lives in:

- `README.md`
- `README_CN.md`

Supplementary references live in:

- `docs/whitepaper*.md`
- `docs/MAINNET_PARAMS*.md`
- `docs/SECURITY_AUDIT*.md`

### Areas for Contribution

| Area | Skills | Difficulty |
|------|--------|------------|
| x/agent module | Go, Cosmos SDK | Medium |
| EVM Precompiles | Go, EVM internals | Hard |
| Solidity interfaces | Solidity | Easy |
| Python SDK | Python | Easy |
| TypeScript SDK | TypeScript | Easy |
| Documentation | Technical writing | Easy |
| AI Challenge mechanism | Go, AI/ML | Hard |
| Testing | Go | Medium |

### Labels

- `good-first-issue` — Great for newcomers
- `help-wanted` — Community help appreciated
- `module:agent` — x/agent module
- `module:precompile` — EVM precompiles
- `sdk:python` — Python SDK
- `consensus` — Consensus mechanism

## Code of Conduct

Be respectful, constructive, and collaborative.

## License

By contributing, you agree that your contributions will be licensed under Apache 2.0.
