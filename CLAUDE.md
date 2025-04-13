# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands
- Build: `make build`
- Run standup: `make standup`
- Create test plugin: `make create-test-plugin`
- Build test plugin: `make build-test-plugin`
- Install test plugin: `make test-plugin`

## Testing
- No test files found; use Go's standard testing package when implementing tests
- Test single package: `go test ./path/to/package`
- Test with verbose output: `go test -v ./...`

## Code Style Guidelines
- Format with Go standard style: `go fmt ./...`
- Error handling: Check returned errors with explicit error checks
- Imports: Group standard library first, then external packages
- Package structure: Follow Go idioms (cmd/, internal/, docs/)
- Naming: Use camelCase for variables, PascalCase for exported functions/types
- Use Cobra for CLI commands and Viper for configuration
- Plugin development follows Go plugin system with specific interfaces from daivplug

## Plugin Development
- Create new plugin: `daiv plugin create <name>`
- See `docs/plugins/GETTING_STARTED.md` for detailed instructions
