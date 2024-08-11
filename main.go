package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
)

var (
	secrets     = make(map[string]string)
	showVersion bool
	Version     string
	Commit      string
)

func toBool(value string) bool {
	return strings.ToLower(value) == "true"
}

func withContextFunc(ctx context.Context, f func()) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(c)

		select {
		case <-ctx.Done():
		case <-c:
			cancel()
			f()
		}
	}()

	return ctx
}

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

	ctx := withContextFunc(context.Background(), func() {})

	token := getGlobalValue("drone_token")
	host := getGlobalValue("drone_server")

	if token == "" {
		panic("missing drone host")
	}

	if host == "" {
		panic("missing drone token")
	}

	droneClient := newDroneClient(host, token, toBool(getGlobalValue("drone_skip_verify")))

	// gets the current user
	user, err := droneClient.Self()
	if err != nil {
		panic("get user failed: " + err.Error())
	}
	slog.Info("login user", "user", user.Login)

	giteaServer := getGlobalValue("gitea_server")
	giteaToken := getGlobalValue("gitea_token")
	giteaSkip := getGlobalValue("gitea_skip_verify")
	syncToGitea := toBool(getGlobalValue("sync_to_gitea"))

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

	if syncToGitea {
		g, err := newGiteaClient(
			ctx,
			giteaServer,
			giteaToken,
			toBool(giteaSkip),
			slog.New(slog.NewTextHandler(os.Stdout, nil)),
		)
		if err != nil {
			slog.Error("failed to init gitea client", "error", err)
			return
		}

		err = g.syncSecret(orgList, repoList, secrets)
		if err != nil {
			slog.Error("failed to sync secret to gitea", "error", err)
			return
		}
		return
	}

	// sync to drone
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
