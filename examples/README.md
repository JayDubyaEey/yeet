# Enhanced Configuration Examples

This directory contains examples demonstrating the enhanced configuration features of yeet.

## Problem Statement

The original scenarios that needed to be addressed:

1. **Different local vs Docker variables**: Need different values for local development vs Docker compose
2. **Direct values vs Key Vault references**: Some values should be literals, others from Key Vault
3. **Environment-aware `yeet run`**: The run command should know which environment to use

## Solutions

### 1. Environment-Specific Configuration

**Problem**: Database URL differs between local development and Docker compose

**Solution**: Use environment-specific configuration:

```json
{
  "keyVaultName": "my-app-vault",
  "mappings": {
    "DATABASE_URL": {
      "local": {
        "type": "literal",
        "value": "postgresql://localhost:5432/myapp_dev"
      },
      "docker": {
        "type": "literal",
        "value": "postgresql://db:5432/myapp_dev"
      }
    }
  }
}
```

**Usage**:
```bash
# For local development
yeet run npm start                     # Uses localhost:5432
yeet fetch                            # Generates .env with localhost:5432

# For Docker compose
yeet run --env docker docker-compose up   # Uses db:5432
yeet fetch                                # Generates docker.env with db:5432
```

### 2. Mixed Value Types

**Problem**: Some values should come from Key Vault, others should be literals

**Solution**: Specify `type` for each value:

```json
{
  "keyVaultName": "my-app-vault",
  "mappings": {
    "DATABASE_URL": {
      "local": {
        "type": "keyvault",
        "value": "dev-database-url"
      },
      "docker": {
        "type": "keyvault", 
        "value": "docker-database-url"
      }
    },
    "REDIS_URL": {
      "local": {
        "type": "literal",
        "value": "redis://localhost:6379"
      },
      "docker": {
        "type": "literal",
        "value": "redis://redis:6379"
      }
    },
    "JWT_SECRET": {
      "local": {
        "type": "literal",
        "value": "local-dev-secret-not-secure"
      },
      "docker": {
        "type": "keyvault",
        "value": "production-jwt-secret"
      }
    }
  }
}
```

### 3. Environment-Aware yeet run

**Problem**: `yeet run` didn't know whether to use local or Docker values

**Solution**: New `--env` flag:

```bash
# Local development (default)
yeet run npm start
yeet run --env local npm start

# Docker environment
yeet run --env docker docker-compose up
yeet run -e docker -- docker-compose up
```
yeet run -e docker -- docker-compose up
```

## Configuration Format Reference

### Enhanced Format

```json
{
  "keyVaultName": "vault-name",
  "mappings": {
    "ENV_VAR": {
      "local": {
        "type": "keyvault|literal",
        "value": "secret-name-or-literal-value"
      },
      "docker": {
        "type": "keyvault|literal", 
        "value": "secret-name-or-literal-value"
      }
    },
    "GLOBAL_VAR": {
      "type": "keyvault|literal",
      "value": "applies-to-both-environments"
    },
    "SIMPLE_VAR": "shorthand-for-keyvault-secret"
  }
}
```

### Value Types

- **`keyvault`**: Fetch from Azure Key Vault
- **`literal`**: Use the value directly

### Environment Types

- **`local`**: Local development (`.env` file, `yeet run` default)
- **`docker`**: Docker environment (`docker.env` file, `yeet run --env docker`)

## Complete Example

See `env.config.json.example` for a comprehensive example covering all scenarios.

## Test Program

Use `test.go` to verify which environment variables are being passed to your application:

```bash
# Test with local environment
yeet run --env local go run examples/test.go

# Test with docker environment  
yeet run --env docker go run examples/test.go
```
