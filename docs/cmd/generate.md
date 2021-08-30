Generate manifests from templates with Vault values

```
argocd-vault-plugin generate PATH [flags]
```

### Options
```
  -c, --config-path string   path to a file containing Vault configuration (YAML, JSON, envfile) to use
  -h, --help                 help for generate
  -s, --secret-name string   name of a Kubernetes Secret containing Vault configuration data in the argocd namespace of your ArgoCD host (Only available when used in ArgoCD)
```

### SEE ALSO

* [argocd-vault-plugin](avp.md) - replace <placeholder\>'s with Vault secrets
