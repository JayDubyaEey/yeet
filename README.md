![Logo](logo.png)

# yeet 🚀

[![CI](https://github.com/JayDubyaEey/yeet/actions/workflows/ci.yml/badge.svg)](https://github.com/JayDubyaEey/yeet/actions/workflows/ci.yml)
[![Release](https://github.com/JayDubyaEey/yeet/actions/workflows/release.yml/badge.svg)](https://github.com/JayDubyaEey/yeet/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/JayDubyaEey/yeet)](https://goreportcard.com/report/github.com/JayDubyaEey/yeet)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)

Yeet pulls secrets from Azure Key Vault and generates `.env` and `docker.env` files for local development and Docker environments.

## Features

- 🔐 Pulls secrets from Azure Key Vault using Azure CLI authentication
- 📁 Generates both `.env` and `docker.env` files
- 🔄 Supports simple and complex mappings with docker-specific overrides
- 🎨 Colorful terminal output with status indicators
- ⚡ Concurrent secret fetching for speed
- 🔍 Validates configuration and checks secret existence
- ⚠️  Warns about unmapped environment variables

## Prerequisites

- Go 1.22+ (for building from source)
- Azure CLI installed and configured
- Access to an Azure Key Vault with required secrets

## Installation

### Download Pre-built Binary (Recommended)

Download the latest release for your platform from the [releases page](https://github.com/JayDubyaEey/yeet/releases).

#### Linux/macOS
```bash
# Linux example (replace VERSION and ARCH as needed)
curl -L https://github.com/JayDubyaEey/yeet/releases/download/vVERSION/yeet_VERSION_linux_x86_64.tar.gz | tar xz
sudo mv yeet /usr/local/bin/

# macOS example (Intel)
curl -L https://github.com/JayDubyaEey/yeet/releases/download/vVERSION/yeet_VERSION_darwin_x86_64.tar.gz | tar xz
sudo mv yeet /usr/local/bin/

# macOS example (Apple Silicon)
curl -L https://github.com/JayDubyaEey/yeet/releases/download/vVERSION/yeet_VERSION_darwin_arm64.tar.gz | tar xz
sudo mv yeet /usr/local/bin/
```

#### Windows

Download the Windows zip file from the releases page and extract `yeet.exe` to a directory in your PATH.

### From Source

```bash
go install github.com/JayDubyaEey/yeet/cmd/yeet@latest
```

### Build Locally

```bash
git clone https://github.com/JayDubyaEey/yeet
cd yeet
go build -o yeet ./cmd/yeet
sudo mv yeet /usr/local/bin/
```

## Configuration

Create an `env.config.json` file in your project directory:

```json
{
  "keyVaultName": "my-keyvault-name",
  "mappings": {
    "DATABASE_URL": "postgres-connection-string",
    "REDIS_URL": "redis-connection-string",
    "API_KEY": "api-key",
    "JWT_SECRET": {
      "secret": "jwt-secret-key",
      "docker": "local-dev-jwt-secret"
    },
    "LOG_LEVEL": {
      "secret": "log-level"
    }
  }
}
```

### Mapping Types

1. **Simple mapping**: `"ENV_VAR": "secret-name"`
   - Both `.env` and `docker.env` will use the secret value

2. **Complex mapping**: `"ENV_VAR": { "secret": "secret-name", "docker": "override-value" }`
   - `.env` uses the secret value
   - `docker.env` uses the override value

## Usage

### Login to Azure
```bash
yeet login
# With specific tenant/subscription
yeet login --tenant YOUR_TENANT --subscription YOUR_SUBSCRIPTION
```

### Fetch Secrets
```bash
# Fetch secrets and generate .env and docker.env
yeet fetch

# Use a different config file
yeet fetch --config path/to/config.json

# Override vault name
yeet fetch --vault different-vault-name
```

### Validate Configuration
```bash
# Check if all secrets exist in Key Vault
yeet validate
```

### List Mappings
```bash
# List all mappings and their status
yeet list

# Show only missing secrets
yeet list --missing-only

# Show only existing secrets
yeet list --exists-only

# Output as JSON for scripting
yeet list --raw
```

### Other Commands
```bash
# Refresh environment files (same as fetch)
yeet refresh

# Logout from Azure CLI
yeet logout

# Show version
yeet version

# Help
yeet --help
yeet fetch --help
```

## Global Flags

- `--config` - Path to configuration file (default: `env.config.json`)
- `--vault` - Override Key Vault name from config
- `--no-color` - Disable colored output
- `-v, --verbose` - Enable verbose logging

## Environment Variables

- `NO_COLOR` - Set to any value to disable colored output

## Security Notes

- Never commit `.env` or `docker.env` files to version control
- Add them to your `.gitignore`
- Secret values are never printed to the console
- Uses Azure CLI's built-in authentication (session persists ~1 week)

## Exit Codes

- `0` - Success
- `1` - General error
- `2` - Validation error
- `3` - Authentication error
- `4` - Secret not found

## License

This project is licensed under the GNU Affero General Public License v3.0 (AGPL-3.0). See the [LICENSE](LICENSE) file for details.
