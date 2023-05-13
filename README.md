# drone-secrets

Synchronize Drone secrets across multiple organizations or repository configurations.

## Motivation

When using Drone as a CI/CD system, there are many organizations or repositories that require the same set of shared secrets, such as Docker credentials or Kubernetes secrets. Consequently, it becomes necessary to reconfigure these secrets every time a new repository or organization is created. This package is designed to address this issue.
