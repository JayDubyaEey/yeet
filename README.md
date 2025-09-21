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
- 📊 Compare configuration with Kubernetes deployment files
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

Create an `env.config.json` file in your project directory. Yeet supports both simple and advanced configuration formats:

### Enhanced Configuration Format (Recommended)

The enhanced format allows you to specify different values for local development vs Docker environments, and distinguish between Key Vault secrets and literal values:

```json
{
  "keyVaultName": "my-keyvault-name",
  "mappings": {
    "DATABASE_URL": {
      "local": {
        "type": "keyvault",
        "value": "postgres-connection-string"
      },
      "docker": {
        "type": "keyvault", 
        "value": "postgres-docker-connection-string"
      }
    },
    "REDIS_URL": {
      "local": {
        "type": "keyvault",
        "value": "redis-connection-string"
      },
      "docker": {
        "type": "literal",
        "value": "redis://redis:6379"
      }
    },
    "JWT_SECRET": {
      "local": {
        "type": "literal",
        "value": "local-dev-secret-123"
      },
      "docker": {
        "type": "keyvault",
        "value": "jwt-secret-production"
      }
    },
    "PORT": {
      "type": "literal",
      "value": "8080"
    },
    "SIMPLE_VAR": "simple-keyvault-secret"
  }
}
```

### Configuration Options

#### Environment-Specific Values
- **`local`**: Values used for local development (`.env` file and `yeet run --env local`)
- **`docker`**: Values used for Docker environments (`docker.env` file and `yeet run --env docker`)

#### Value Types
- **`keyvault`**: Fetch value from Azure Key Vault using the specified secret name
- **`literal`**: Use the specified value directly (no Key Vault lookup)

#### Global Values
- **`type` + `value`**: Applied to both environments when no environment-specific config exists
- **Simple string**: Shorthand for `{"type": "keyvault", "value": "secret-name"}`

## Usage

### Login to Azure
```bash
yeet login
# With specific tenant/subscription
yeet login --tenant YOUR_TENANT --subscription YOUR_SUBSCRIPTION
```

### Run Commands with Secrets
```bash
# Run with local environment (default)
yeet run make dev
yeet run npm start

# Run with docker environment
yeet run --env docker docker-compose up
yeet run -e docker -- docker-compose up

# Use a different vault
yeet run --vault production-vault make deploy

# Load .env file for local overrides
yeet run --load-env make dev
yeet run -l --env-file custom.env npm test
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

### Compare with Kubernetes Deployments
```bash
# Compare config with Kubernetes deployment file
yeet compare

# Specify custom deployment file path
yeet compare --deployment-path path/to/deployment.yml

# Use different environment for comparison
yeet compare --env docker
```

The compare command analyzes your configuration against Kubernetes deployment files and shows:
- Variables in your config but missing from the deployment
- Variables in the deployment but missing from your config
- Environment variable mismatches and recommendations

This helps ensure your configuration stays in sync with your Kubernetes deployments.

### Other Commands
```bash
# Compare with Kubernetes deployment files
yeet compare

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

## Kubernetes Integration

### Comparing with Deployment Files

Yeet can compare your configuration with Kubernetes deployment files to ensure your environment variables are properly aligned:

```bash
# Compare with default deployment file (deploy/base/deployment.yml)
yeet compare

# Compare with custom deployment file
yeet compare --deployment-path k8s/production/deployment.yaml

# Compare using docker environment settings
yeet compare --env docker --deployment-path k8s/staging/deployment.yaml
```

#### What the Compare Command Checks

The compare command analyzes both your `env.config.json` and Kubernetes deployment YAML files to identify:

1. **Missing in Deployment**: Variables defined in your config but not present in the Kubernetes deployment
2. **Missing in Config**: Environment variables in the deployment that aren't defined in your config
3. **Value Type Mismatches**: Helps identify when you're using literal values vs secrets in different environments

#### Example Output

```bash
$ yeet compare
✅ Configuration loaded successfully
✅ Deployment file loaded: deploy/base/deployment.yml

📊 Comparison Results:

❌ Variables in config but missing from deployment:
  • REDIS_URL
  • API_KEY

⚠️  Variables in deployment but missing from config:
  • USER_SERVICE_BASE_URL
  • POSTGRES_DATABASE

✅ Matching variables (5):
  • LOG_LEVEL
  • JWT_SECRET
  • POSTGRES_HOST
  • POSTGRES_PORT
  • POSTGRES_PASSWORD

💡 Recommendations:
  • Add missing variables to your Kubernetes deployment
  • Consider adding USER_SERVICE_BASE_URL to your config if needed
  • Review if POSTGRES_DATABASE should be configurable
```

#### Supported Kubernetes Resources

The compare command supports:
- **Deployments** - Extracts env vars from container specifications
- **StatefulSets** - Analyzes environment variables in pod templates
- **DaemonSets** - Checks environment configuration across daemon pods
- **Jobs/CronJobs** - Validates job container environment variables

It handles both direct environment variable values and references to ConfigMaps/Secrets via `valueFrom`.

## Global Flags

- `--config` - Path to configuration file (default: `env.config.json`)
- `--vault` - Override Key Vault name from config
- `--env` - Environment to use (local/docker, default: local)
- `--deployment-path` - Path to Kubernetes deployment file (compare command)
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
