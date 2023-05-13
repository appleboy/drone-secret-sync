package main

import (
	"os"
	"strings"
)

func main() {
	os.Getenv("PORT")
}

func getGlobalValue(key string) string {
	key = strings.ToUpper(key)

	if value := os.Getenv("INPUT_" + key); value != "" {
		return value
	}

	return os.Getenv(key)
}
