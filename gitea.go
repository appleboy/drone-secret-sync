package main

import (
	"context"
	"crypto/tls"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	gsdk "code.gitea.io/sdk/gitea"
	"github.com/drone/drone-go/drone"
)

// gitea is a struct that holds the gitea client.
type gitea struct {
	ctx        context.Context
	server     string
	token      string
	skipVerify bool
	client     *gsdk.Client
	logger     *slog.Logger
}

// init initializes the gitea client.
func (g *gitea) init() error {
	if g.server == "" || g.token == "" {
		return errors.New("mission gitea server or token")
	}

	g.server = strings.TrimRight(g.server, "/")

	opts := []gsdk.ClientOption{
		gsdk.SetToken(g.token),
	}

	if g.skipVerify {
		// add new http client for skip verify
		httpClient := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
		opts = append(opts, gsdk.SetHTTPClient(httpClient))
	}

	client, err := gsdk.NewClient(g.server, opts...)
	if err != nil {
		return err
	}
	g.client = client

	return nil
}

func (g *gitea) syncSecret(
	client drone.Client,
	orgList []string,
	repoList []string,
	secrets map[string]string,
) error {
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
			if _, ok := secretMaps[k]; !ok {
				continue
			}
			_, err := g.client.CreateOrgActionSecret(org, gsdk.CreateSecretOption{
				Name: k,
				Data: v,
			})
			if err != nil {
				slog.Error("failed to update org secrets", "org", org, "key", k, "error", err)
				continue
			}
			slog.Info("update org secret", "org", org, "key", k)
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
			if _, ok := secretMaps[k]; !ok {
				continue
			}
			_, err := g.client.CreateRepoActionSecret(owner, name, gsdk.CreateSecretOption{
				Name: k,
				Data: v,
			})
			if err != nil {
				slog.Error("failed to update repo secrets", "repo", repo, "key", k, "error", err)
				continue
			}
			slog.Info("update repo secret", "repo", repo, "key", k)
		}
	}
	return nil
}

// newGiteaClient creates a new instance of the gitea struct.
func newGiteaClient(
	ctx context.Context,
	server string,
	token string,
	skipVerify bool,
	logger *slog.Logger,
) (*gitea, error) {
	g := &gitea{
		ctx:        ctx,
		server:     server,
		token:      token,
		skipVerify: skipVerify,
		logger:     logger,
	}

	err := g.init()
	if err != nil {
		return nil, err
	}

	return g, nil
}