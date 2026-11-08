package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
	"testing"

	gsdk "code.gitea.io/sdk/gitea"
)

func TestGiteaSyncSecret(t *testing.T) {
	const secretName = "DEPLOY_TOKEN"
	const secretValue = "test-only-value"
	orgPath := "/api/v1/orgs/team/actions/secrets/" + secretName
	repoPath := "/api/v1/repos/team/project/actions/secrets/" + secretName
	for _, tc := range []struct {
		name      string
		orgStatus int
	}{
		{name: "create", orgStatus: http.StatusCreated},
		{name: "update", orgStatus: http.StatusNoContent},
		{name: "continue after org failure", orgStatus: http.StatusForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var mu sync.Mutex
			var paths []string
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					mu.Lock()
					defer mu.Unlock()
					paths = append(paths, r.URL.Path)
					if r.Method != http.MethodPut {
						t.Errorf("method = %s, want PUT", r.Method)
					}
					var payload map[string]string
					if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
						t.Errorf("decode payload: %v", err)
					}
					if payload["data"] != secretValue {
						t.Error("secret data did not reach the server")
					}
					if _, ok := payload["name"]; ok {
						t.Error("secret name belongs in the URL, not the payload")
					}
					switch r.URL.Path {
					case orgPath:
						w.WriteHeader(tc.orgStatus)
					case repoPath:
						w.WriteHeader(http.StatusCreated)
					default:
						t.Errorf("unexpected request path: %s", r.URL.Path)
						w.WriteHeader(http.StatusNotFound)
					}
				}),
			)
			t.Cleanup(server.Close)
			client, err := gsdk.NewClient(server.URL, gsdk.SetGiteaVersion("1.26.0"))
			if err != nil {
				t.Fatal(err)
			}
			g := &gitea{client: client}
			if err := g.syncSecret(
				[]string{"team"},
				[]string{"invalid", "team/project"},
				map[string]string{
					secretName: secretValue,
				},
			); err != nil {
				t.Fatal(err)
			}
			mu.Lock()
			defer mu.Unlock()
			if !slices.Equal(paths, []string{orgPath, repoPath}) {
				t.Errorf("request paths = %v, want organization then repository", paths)
			}
		})
	}
}
