# drone-secret-sync

[![Lint and Testing](https://github.com/appleboy/drone-secret-sync/actions/workflows/testing.yml/badge.svg)](https://github.com/appleboy/drone-secret-sync/actions/workflows/testing.yml)

Synchronize Drone secrets across multiple organizations or repository configurations.

## Motivation

When using Drone as a CI/CD system, there are many organizations or repositories that require the same set of shared secrets, such as Docker credentials or Kubernetes secrets. Consequently, it becomes necessary to reconfigure these secrets every time a new repository or organization is created. This package is designed to address this issue.

## How it works

This package uses the Drone API to synchronize secrets across multiple organizations or repository configurations. It can be used as a CLI tool or as a Docker image.

### Using CLI

Create a `.env` file with the following content. You can also use the `export` command to set the environment variables.

```sh
DRONE_SERVER=https://cloud.drone.io
DRONE_TOKEN=xxxxx
ORG_LIST=appleboy
REPO_LIST=go-training/golang-in-ecr-ecs,go-training/drone-git-push-example
KEY_LIST=FOOBAR2
FOOBAR2=1234
```

* DRONE_SERVER: Drone server URL.
* DRONE_TOKEN: Drone token.
* ORG_LIST: Comma-separated list of organizations.
* REPO_LIST: Comma-separated list of repositories.
* KEY_LIST: Comma-separated list of secret keys.

Download the latest release from [GitHub](https://github.com/appleboy/drone-secret-sync/releases) and run the following command:

```sh
drone-secret-sync
```

The above command will sync the `FOOBAR2` secret to the `appleboy` organization and `go-training/golang-in-ecr-ecs` and `go-training/drone-git-push-example` repositories. See the following output:

```sh
login user: appleboy
org: appleboy, update secret key: foobar2
repo: go-training/golang-in-ecr-ecs, update secret key: foobar2
repo: go-training/drone-git-push-example, update secret key: foobar2
```

### Using Drone CI/CD

You can also use this package as a Docker image in your Drone CI/CD pipeline. See the following example:

```yaml
---
kind: pipeline
name: linux-amd64

platform:
  os: linux
  arch: amd64

steps:

- name: publish
  pull: always
  image: ghcr.io/appleboy/drone-secret-sync:1
  settings:
    drone_token:
      from_secret: drone_token
    drone_server:
      from_secret: drone_server
    org_list: appleboy,go-training
    repo_list: go-training/golang-in-ecr-ecs,go-training/drone-git-push-example
    key_list: foobar,docker_test_username,docker_test_token
    foobar: test
  environment:
    DOCKER_TEST_USERNAME: appleboy
    DOCKER_TEST_TOKEN: 1234
```
