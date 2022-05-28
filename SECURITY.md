# Security policy for argocd-vault-plugin

## Policy

- We aim to keep project dependencies up to date with Github's Dependabot feature, acting on any relevant security notices published to the repo

- We aim to frequently use `go mod tidy` to limit dependencies to the bare necessities

- This repo has [Github's CodeQL](./.github/workflows/codeql.yml) analysis enabled as part of the CI checks on all PRs


## Discloure

Please report any security vulnerabilities found to `echo <name of this repo> | sed 's/-//g`, at `gmail.com`