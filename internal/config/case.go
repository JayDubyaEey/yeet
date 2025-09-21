package config

import (
	"strings"
)

func ToKebabCase(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, "_", "-"))
}

func ToShoutingSnakeCase(s string) string {
	return strings.ToUpper(strings.ReplaceAll(s, "-", "_"))
}

func IsShoutingSnakeCase(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '_' {
			return false
		}
	}
	first := rune(s[0])
	return (first >= 'A' && first <= 'Z') || first == '_'
}

func IsKebabCase(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' {
			return false
		}
	}
	first := rune(s[0])
	return first >= 'a' && first <= 'z'
}
