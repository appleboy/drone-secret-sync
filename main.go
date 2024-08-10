package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

var (
	secrets     = make(map[string]string)
	showVersion bool
	Version     string
	Commit      string
)

func main() {
	var envfile string
	flag.StringVar(&envfile, "env-file", ".env", "Read in a file of environment variables")
	flag.BoolVar(&showVersion, "version", false, "Show version")
	flag.Parse()

	if showVersion {
		fmt.Printf("Version: %s Commit: %s\n", Version, Commit)
		return
	}

	_ = godotenv.Load(envfile)

	token := getGlobalValue("drone_token")
	host := getGlobalValue("drone_server")

	if token == "" {
		panic("missing drone host")
	}

	if host == "" {
		panic("missing drone token")
	}

	droneClient := newDroneClient(host, token)
	// gets the current user
	user, err := droneClient.Self()
	if err != nil {
		panic("get user failed: " + err.Error())
	}
	slog.Info("login user", "user", user.Login)

	// get global secrets
	orgValue := getGlobalValue("org_list")
	repoValue := getGlobalValue("repo_list")
	keyValue := getGlobalValue("key_list")

	// split org, repo, key
	orgList := strings.Split(orgValue, ",")
	repoList := strings.Split(repoValue, ",")
	keyList := strings.Split(keyValue, ",")

	for _, key := range keyList {
		// check value is empty
		value := getGlobalValue(key)
		if value == "" {
			continue
		}

		key = strings.ToLower(key) // Convert key to lowercase
		secrets[key] = value
	}

	syncToDrone(droneClient, orgList, repoList, secrets)
}

func getGlobalValue(key string) string {
	key = strings.ToUpper(key) // Convert key to uppercase

	// Check if there is an environment variable with the format "PLUGIN_<KEY>"
	if value := os.Getenv("PLUGIN_" + key); value != "" {
		return value // Return the value of the "PLUGIN_<KEY>" environment variable
	}

	// If the "PLUGIN_<KEY>" environment variable doesn't exist or is empty,
	// return the value of the "<KEY>" environment variable
	return os.Getenv(key)
}
