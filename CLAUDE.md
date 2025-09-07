# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a CLI tool for managing GitHub repository custom properties at scale. Built in Go using Cobra CLI framework, it allows bulk configuration of GitHub repository custom properties via YAML files with plan/apply workflow similar to Terraform.

## Development Commands

### Building and Running
```bash
# Build the binary
go build -o gh-custom-property-manager main.go

# Run directly
go run main.go [command] [flags]

# Quick development commands via Makefile
make plan    # Preview changes using example configs
make apply   # Apply changes using example configs
```

### Testing and Quality
```bash
# Run all tests with coverage
go test -v ./... -cover
make test

# Run Go vet for static analysis
go vet ./...

# Linting (via CI)
golangci-lint run
```

### Authentication Setup
The tool requires GitHub authentication:
```bash
# Set up GitHub CLI authentication
gh auth login

# The tool uses GITHUB_TOKEN environment variable
export GITHUB_TOKEN=$(gh auth token)
```

## Code Architecture

### Core Components

**Entry Point**: `main.go` → `cmd.Execute()` - Simple entry point that delegates to Cobra CLI

**CLI Commands** (`cmd/`):
- `root.go` - Base command with version info, uses goreleaser build variables
- `plan.go` - Shows diff of proposed changes (dry-run mode)
- `apply.go` - Executes the property changes

**Configuration System** (`config/`):
- `Config` struct manages multiple YAML configuration files
- `ConfigFile` struct defines the YAML schema for property definitions
- `PropertyDiff` represents changes to be applied
- Validates against duplicate repository/property combinations across config files
- Generates diffs by comparing current GitHub state with desired state

**GitHub Client** (`client/`):
- Wraps `google/go-github` v74 client with OAuth2 authentication
- `GetRepository()` - Fetches repository with custom properties
- `UpdateCustomProperties()` - Applies property changes via GitHub API

### Configuration File Structure
YAML files in `property/` directory follow this schema:
```yaml
property_name: "property-name"
values:
  - value: "property-value"
    repositories:
      - name: "org/repo"
```

### Key Design Patterns

**Interface-Based GitHub Client**: `GitHubClient` interface in config package enables testing and decoupling

**Validation Layer**: `validateNoDuplicateRepositoryValues()` prevents conflicts between and within config files for the same property

**Diff-Based Changes**: Three-step process:
1. Load configs and validate
2. Generate repository list and fetch current state  
3. Calculate diffs and apply changes

**Error Handling**: Early returns with descriptive error messages throughout the pipeline

### Dependencies
- `github.com/spf13/cobra` - CLI framework
- `github.com/google/go-github/v74` - GitHub API client
- `github.com/goccy/go-yaml` - YAML parsing
- `golang.org/x/oauth2` - GitHub authentication

### Release Process
- Uses GoReleaser for automated releases on git tags
- Builds for Linux, Windows, macOS (amd64, arm64)
- CI pipeline runs tests on Go 1.23 and 1.24
- Includes linting via golangci-lint