package main

import (
	"context"
	"log/slog"
	"strings"

	"github.com/drone/drone-go/drone"
	"golang.org/x/oauth2"
)

func newDroneClient(host, token string) drone.Client {
	// create an http client with oauth authentication.
	cfg := new(oauth2.Config)
	auther := cfg.Client(
		context.Background(),
		&oauth2.Token{
			AccessToken: token,
		},
	)

	// create the drone client with authenticator
	return drone.NewClient(host, auther)
}

func syncToDrone(
	client drone.Client,
	orgList []string,
	repoList []string,
	secrets map[string]string,
) {
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
