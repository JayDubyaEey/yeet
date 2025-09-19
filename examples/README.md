# Yeet Examples

This directory contains examples of how to use yeet in your projects.

## The Problem

Traditional approach with `.env` files:
1. Run `yeet fetch` to create `.env` file
2. Source the `.env` file in your scripts/Makefile
3. Files can get out of sync
4. Secrets are written to disk

## The Solution: `yeet run`

With `yeet run`, you can run any command with secrets injected as environment variables:

```bash
# No .env file needed!
yeet run make dev
yeet run npm start
yeet run python manage.py runserver
```

## Makefile Integration

See the [Makefile](./Makefile) in this directory for a complete example.

### Before (traditional approach):
```makefile
dev:
	@source .env && npm start
```

### After (with yeet run):
```makefile
dev:
	@yeet run npm start
```

## Benefits

1. **No files on disk** - Secrets never touch the filesystem
2. **Always up-to-date** - Fetches latest secrets from Key Vault every time
3. **Vault switching** - Easy to switch between vaults: `yeet run -v staging make dev`
4. **Local overrides** - Still support .env for local development: `yeet run --load-env make dev`
5. **CI/CD friendly** - Perfect for pipelines where you don't want to create files

## Advanced Usage

### Multiple Vaults
```bash
# Development
yeet run --vault dev-vault make dev

# Staging
yeet run --vault staging-vault make deploy

# Production
yeet run --vault prod-vault make deploy
```

### Local Overrides
```bash
# Create a .env with overrides
echo "DEBUG=true" > .env

# Run with both Key Vault secrets and local overrides
yeet run --load-env make dev
```

### Complex Commands
```bash
# Use -- to separate yeet flags from command flags
yeet run -- docker-compose up -d
yeet run -- npm run build --production
```
