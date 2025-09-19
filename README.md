[![CI](https://github.com/JayDubyaEey/yeet/actions/workflows/ci.yml/badge.svg)](https://github.com/JayDubyaEey/yeet/actions/workflows/ci.yml)
[![Auto Release](https://github.com/JayDubyaEey/yeet/actions/workflows/auto-release.yml/badge.svg)](https://github.com/JayDubyaEey/yeet/actions/workflows/auto-release.yml)
[![Release](https://github.com/JayDubyaEey/yeet/actions/workflows/release.yml/badge.svg)](https://github.com/JayDubyaEey/yeet/actions/workflows/release.yml)
[![Publish](https://github.com/JayDubyaEey/yeet/actions/workflows/publish.yml/badge.svg)](https://github.com/JayDubyaEey/yeet/actions/workflows/publish.yml)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)

![Logo](logo.png)

Yeet pulls secrets from Azure Key Vault and generates `.env` and `docker.env` files for local development and Docker environments.

## Features

- 🔐 Pulls secrets from Azure Key Vault using Azure CLI authentication
- 📁 Generates both `.env` and `docker.env` files
- 🚀 Run commands directly with secrets as environment variables (no files needed!)
- 🔄 Supports simple and complex mappings with docker-specific overrides
- ⚡ Concurrent secret fetching for speed
- 🔍 Validates configuration and checks secret existence
- ⚠️  Warns about unmapped environment variables
- 🎯 Perfect for Makefiles and CI/CD pipelines

## Prerequisites

### Runtime Dependencies

- **Azure CLI** (required) - Used for authentication and Key Vault access
  - Install: [Azure CLI Installation Guide](https://learn.microsoft.com/en-us/cli/azure/install-azure-cli)
  - Version: 2.0.0 or higher recommended
  - Must be authenticated (`az login`) with access to your Key Vault
  
  **Quick Install:**
  ```bash
  # Ubuntu/Debian
  curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash
  
  # macOS
  brew update && brew install azure-cli
  
  # Windows (PowerShell as Administrator)
  winget install -e --id Microsoft.AzureCLI
  ```

### Build Dependencies

- **Go 1.24+** (only for building from source)
  - Install: [Go Installation Guide](https://go.dev/doc/install)
  - Note: Uses Go 1.24 for the latest features and performance improvements

### Azure Requirements

- Azure subscription with an existing Key Vault
- Appropriate RBAC permissions:
  - `Key Vault Secrets User` role (minimum) for reading secrets
  - `Key Vault Reader` role for listing secrets

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
go install github.com/JayDubyaEey/yeet/cmd@latest
```

### Build Locally

```bash
git clone https://github.com/JayDubyaEey/yeet
cd yeet
go build -o yeet ./cmd/main.go
sudo mv yeet /usr/local/bin/
```

## Verify Dependencies

Before using yeet, ensure all dependencies are properly installed:

```bash
# Check if Azure CLI is installed
az --version

# Check if you're logged into Azure
az account show

# If not logged in, authenticate with Azure
az login

# List your Key Vaults to verify access
az keyvault list -o table
```

## Configuration

### Quick Start with init

If you have an existing `.env.example` file, use the `init` command to generate your configuration automatically:

```bash
# Initialize with your Key Vault name
yeet init my-keyvault-name

# Initialize with default placeholder (update later)
yeet init

# Use a custom .env.example file
yeet init my-keyvault --env-example .env.sample
```

The `init` command will:
1. Parse your `.env.example` file
2. Generate appropriate secret names from environment variable names
3. Detect local development values and suggest them as docker overrides
4. Create an `env.config.json` file ready for customization

### Manual Configuration

Alternatively, create an `env.config.json` file manually in your project directory:

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

### Initialize Configuration
```bash
# Generate config from .env.example
yeet init my-keyvault-name

# Use custom paths
yeet init --env-example .env.sample --output my-config.json

# Force overwrite existing config
yeet init my-keyvault --force
```

### Login to Azure
```bash
yeet login
# With specific tenant/subscription
yeet login --tenant YOUR_TENANT --subscription YOUR_SUBSCRIPTION
```

### Run Commands with Secrets
```bash
# Run a command with secrets as environment variables (no .env file created)
yeet run make dev
yeet run npm start
yeet run -- docker-compose up

# Use a different vault
yeet run --vault production-vault make deploy

# Load .env file for local overrides
yeet run --load-env make dev
yeet run -e --env-file custom.env npm test
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

## Troubleshooting

### Azure CLI not found

If you get an error about Azure CLI not being found:

1. Verify Azure CLI is installed:
   ```bash
   which az
   ```

2. If not found, install it using the platform-specific instructions in the Prerequisites section.

3. Ensure Azure CLI is in your PATH:
   ```bash
   export PATH=$PATH:/usr/local/bin
   ```

### Authentication errors

If you get authentication errors:

1. Check your current Azure login status:
   ```bash
   az account show
   ```

2. Re-authenticate if needed:
   ```bash
   az login
   ```

3. Set the correct subscription:
   ```bash
   az account set --subscription "Your Subscription Name"
   ```

### Key Vault access errors

1. Verify you have access to the Key Vault:
   ```bash
   az keyvault show --name your-keyvault-name
   ```

2. Check your permissions:
   ```bash
   az role assignment list --assignee $(az account show --query user.name -o tsv) --scope /subscriptions/YOUR_SUB_ID/resourceGroups/YOUR_RG/providers/Microsoft.KeyVault/vaults/YOUR_KV
   ```

3. Ensure you have at least `Key Vault Secrets User` role.

## GitHub Packages

This project publishes releases to GitHub Packages:

### Binary Artifacts

Pre-built binaries are automatically built and attached to each GitHub Release. Additionally, binaries from recent builds are available as workflow artifacts with a 30-day retention period.

Supported platforms:
- Linux (amd64, arm64, 386)
- macOS (amd64, arm64)
- Windows (amd64)

## License

This project is licensed under the GNU Affero General Public License v3.0 (AGPL-3.0). See the [LICENSE](LICENSE) file for details.
