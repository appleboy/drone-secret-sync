package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/drone/drone-go/drone"
	"golang.org/x/oauth2"
)

var secrets map[string]string

func main() {
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
		oauth2.NoContext,
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
	fmt.Printf("login user: %s", user.Login)

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
		for k, v := range secrets {
			// delete org secret
			if err := client.OrgSecretDelete(org, k); err != nil {
				panic("delete org secret failed: " + err.Error())
			}

			// create org secret
			if _, err := client.OrgSecretCreate(org, &drone.Secret{
				Namespace: org,
				Name:      k,
				Data:      v,
			}); err != nil {
				panic("delete org secret failed: " + err.Error())
			}
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
		for k, v := range secrets {
			// delete org secret
			if err := client.SecretDelete(owner, name, k); err != nil {
				panic("delete repo secret failed: " + err.Error())
			}

			// create org secret
			if _, err := client.SecretCreate(owner, name, &drone.Secret{
				Name: k,
				Data: v,
			}); err != nil {
				panic("delete repo secret failed: " + err.Error())
			}
		}
	}
}

func getGlobalValue(key string) string {
	key = strings.ToUpper(key) // Convert key to uppercase

	// Check if there is an environment variable with the format "INPUT_<KEY>"
	if value := os.Getenv("INPUT_" + key); value != "" {
		return value // Return the value of the "INPUT_<KEY>" environment variable
	}

	// If the "INPUT_<KEY>" environment variable doesn't exist or is empty,
	// return the value of the "<KEY>" environment variable
	return os.Getenv(key)
}
