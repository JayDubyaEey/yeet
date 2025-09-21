package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

func main() {
	fmt.Println("🔍 Application Environment Variables")
	fmt.Println("=" + strings.Repeat("=", 50))

	envVars := os.Environ()
	sort.Strings(envVars)

	customVars := filterCustomVars(envVars)
	displayResults(customVars, envVars)
	checkCommonAppVars()
	displayUsage()
}

func filterCustomVars(envVars []string) []string {
	systemPrefixes := []string{
		"PATH", "HOME", "USER", "SHELL", "TERM", "PWD", "OLDPWD",
		"LANG", "LC_", "XDG_", "DISPLAY", "TMPDIR", "TMP", "TEMP",
		"GOPATH", "GOROOT", "GOCACHE", "GOMODCACHE", "GO111MODULE",
		"SSH_", "GPG_", "_", "SHLVL", "LOGNAME", "MAIL", "HOSTNAME",
		"EDITOR", "PAGER", "LESS", "MORE", "MANPATH", "INFOPATH",
	}

	var customVars []string
	for _, env := range envVars {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		isSystem := false
		for _, prefix := range systemPrefixes {
			if strings.HasPrefix(key, prefix) {
				isSystem = true
				break
			}
		}

		if !isSystem {
			customVars = append(customVars, env)
		}
	}
	return customVars
}

func displayResults(customVars, envVars []string) {
	if len(customVars) > 0 {
		fmt.Printf("📊 Found %d application variables:\n\n", len(customVars))
		fmt.Println("🎯 Application Variables:")
		fmt.Println("-" + strings.Repeat("-", 40))
		for _, env := range customVars {
			parts := strings.SplitN(env, "=", 2)
			key := parts[0]
			value := parts[1]

			if len(value) > 100 {
				value = value[:97] + "..."
			}

			fmt.Printf("  %-30s = %s\n", key, value)
		}
		fmt.Println()
	} else {
		fmt.Println("📊 No application variables found")
		fmt.Println("💡 Only system variables are present")
		fmt.Println()
	}

	fmt.Println("📋 Summary:")
	fmt.Printf("  • Application variables: %d\n", len(customVars))
	fmt.Printf("  • Total variables found: %d\n", len(envVars))
}

func checkCommonAppVars() {
	fmt.Println("\n🔍 Looking for common application variables:")
	commonAppVars := []string{
		"DATABASE_URL", "DB_HOST", "DB_USER", "DB_PASSWORD", "DB_NAME",
		"API_KEY", "SECRET_KEY", "JWT_SECRET",
		"REDIS_URL", "REDIS_HOST", "REDIS_PORT",
		"PORT", "HOST", "NODE_ENV", "GO_ENV", "ENV",
		"DEBUG", "LOG_LEVEL",
	}

	found := false
	for _, varName := range commonAppVars {
		if value := os.Getenv(varName); value != "" {
			if !found {
				fmt.Println("  Found:")
				found = true
			}
			if len(value) > 50 {
				value = value[:47] + "..."
			}
			fmt.Printf("    %-20s = %s\n", varName, value)
		}
	}

	if !found {
		fmt.Println("  No common application variables found")
	}
}

func displayUsage() {
	fmt.Println("\n💡 Usage:")
	fmt.Println("  go run test-env.go                    # Show application variables only")
	fmt.Println("  source .env && go run test-env.go     # Test with .env file")
	fmt.Println("  export $(cat docker.env | xargs) && go run test-env.go  # Test with docker.env")
}
