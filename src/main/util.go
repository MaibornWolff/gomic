package main

import (
	"os"
	"strings"
)

func getEnv(key string, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	trimmedValue := strings.TrimSpace(value)
	if exists && len(trimmedValue) > 0 {
		return trimmedValue
	}
	return defaultValue
}
