package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/drone/drone-go/drone"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
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

	// create an http client with oauth authentication.
	cfg := new(oauth2.Config)
	auther := cfg.Client(
		context.Background(),
		&oauth2.Token{
			AccessToken: token,
		},
	)

	// create the drone client with authenticator
	client := drone.NewClient(host, auther)

	// gets the current user
	user, err := client.Self()
	if err != nil {
		panic("get user failed: " + err.Error())
	}
	slog.Info("login user", "user", user.Login)

	orgValue := getGlobalValue("org_list")
	orgList := strings.Split(orgValue, ",")
	repoValue := getGlobalValue("repo_list")
	repoList := strings.Split(repoValue, ",")
	keyValue := getGlobalValue("key_list")
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

	// update org secrets
	for _, org := range orgList {
		// get all org secrets
		secretMaps := make(map[string]string)
		allSecrets, err := client.OrgSecretList(org)
		if err != nil {
			panic("get org secret failed: " + err.Error())
		}
		for _, secret := range allSecrets {
			secretMaps[secret.Name] = secret.Data
		}

		for k, v := range secrets {
			// update org secret
			if _, ok := secretMaps[k]; ok {
				// create org secret
				if _, err := client.OrgSecretUpdate(org, &drone.Secret{
					Namespace: org,
					Name:      k,
					Data:      v,
				}); err != nil {
					panic("update org secret failed: " + err.Error())
				}
				slog.Info("update org secret", "org", org, "key", k)
				continue
			}

			// create org secret
			if _, err := client.OrgSecretCreate(org, &drone.Secret{
				Namespace: org,
				Name:      k,
				Data:      v,
			}); err != nil {
				panic("delete org secret failed: " + err.Error())
			}
			slog.Info("create org secret", "org", org, "key", k)
		}
	}

	// update repo secrets
	for _, repo := range repoList {
		val := strings.Split(repo, "/")
		if len(val) != 2 {
			continue
		}
		owner := val[0]
		name := val[1]

		// get all org secrets
		secretMaps := make(map[string]string)
		allSecrets, err := client.SecretList(owner, name)
		if err != nil {
			panic("get repo secret failed: " + err.Error())
		}
		for _, secret := range allSecrets {
			secretMaps[secret.Name] = secret.Data
		}

		for k, v := range secrets {
			// update repo secret
			if _, ok := secretMaps[k]; ok {
				// create repo secret
				if _, err := client.SecretUpdate(owner, name, &drone.Secret{
					Name: k,
					Data: v,
				}); err != nil {
					panic("update repo secret failed: " + err.Error())
				}
				slog.Info("update repo secret", "repo", repo, "key", k)
				continue
			}

			// create repo secret
			if _, err := client.SecretCreate(owner, name, &drone.Secret{
				Name: k,
				Data: v,
			}); err != nil {
				panic("delete repo secret failed: " + err.Error())
			}
			slog.Info("create repo secret", "repo", repo, "key", k)
		}
	}
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
