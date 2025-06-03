# GitHub Custom Property Manager

A CLI tool for efficiently managing GitHub repository custom properties.

## Features

- Bulk configuration of GitHub repository custom properties
- Preview changes before applying (plan)
- Apply changes (apply)
- Property management via YAML configuration files

## Installation

### Prerequisites

- Go 1.24.3 or higher
- GitHub CLI (`gh`) installed and authenticated

### Build

```bash
git clone https://github.com/hi120ki/gh-custom-property-manager.git
cd gh-custom-property-manager
go build -o gh-custom-property-manager main.go
```

## Usage

### 1. GitHub Authentication

Set up your token using GitHub CLI:

```bash
gh auth login
```

### 2. Create Configuration Files

Define custom properties in YAML files:

```yaml
property_name: property-a
values:
  - value: "true"
    repositories:
      - name: owner/repo1
  - value: "false"
    repositories:
      - name: owner/repo2
      - name: owner/repo3
```

### 3. Preview Changes

```bash
make plan
# or
GITHUB_TOKEN=$(gh auth token) go run main.go plan --config property/property-a.yaml
```

### 4. Apply Changes

```bash
make apply
# or
GITHUB_TOKEN=$(gh auth token) go run main.go apply --config property/property-a.yaml
```

## Commands

- `plan`: Display changes (dry-run)
- `apply`: Actually apply changes

## Configuration File Format

```yaml
property_name: "property name"
values:
  - value: "property value"
    repositories:
      - name: "organization/repository"
```

## Makefile

The following commands are available:

```bash
make help    # Display help
make plan    # Preview changes
make apply   # Apply changes
```

## License

This project is licensed under the [MIT License](LICENSE).

## Author

[Hi120ki](https://github.com/hi120ki)

## Contributing

Bug reports and feature requests are welcome via [Issues](https://github.com/hi120ki/gh-custom-property-manager/issues). Pull requests are also welcome.

## Disclaimer

This software is provided "as is" without any express or implied warranties. Use at your own risk.
