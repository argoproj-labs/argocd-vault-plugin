Generate manifests from templates with Vault values

```
argocd-vault-plugin generate PATH [flags]
```

### Options
```
  -c, --config-path string         path to a file containing Vault configuration (YAML, JSON, envfile) to use
  -h, --help                       help for generate
  -s, --secret-name string         name of a Kubernetes Secret in the argocd namespace containing Vault configuration data in the argocd namespace of your ArgoCD host (Only available when used in ArgoCD). The namespace can be overridden by using the format <namespace>:<name>
      --verbose-sensitive-output   enable verbose mode for detailed info to help with debugging. Includes sensitive data (credentials), logged to stderr
```

### SEE ALSO

* [argocd-vault-plugin](avp.md) - replace <placeholder\>'s with Vault secrets
