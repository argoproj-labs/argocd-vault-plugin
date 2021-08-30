### Command Line
The plugin can be used via the command line or any shell script. Since the plugin outputs yaml to standard out, you can run the `generate` command and pipe the output to `kubectl`.

`argocd-vault-plugin generate ./ | kubectl apply -f -`

This will pull the values from Vault, replace the placeholders and then apply the yamls to whatever kubernetes cluster you are connected to.

You can also read from stdin like so:

`cat example.yaml | argocd-vault-plugin generate - | kubectl apply -f -`

### Argo CD
Before using the plugin in Argo CD you must follow the [steps](installation.md#installing-in-argo-cd) to install the plugin to your Argo CD instance. Once the plugin is installed, you can use it 3 ways.

1. Select your plugin via the UI by selecting `New App` and then changing `Directory` at the bottom of the form to be `argocd-vault-plugin`.

2. Apply a Argo CD Application yaml that has `argocd-vault-plugin` as the plugin.
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: your-application-name
spec:
  destination:
    name: ''
    namespace: default
    server: 'https://kubernetes.default.svc'
  source:
    path: .
    plugin:
      name: argocd-vault-plugin
    repoURL: http://your-repo/
    targetRevision: HEAD
  project: default
```

3. Or you can pass the config-management-plugin flag to the Argo CD CLI app create command:  
`argocd app create you-app-name --config-management-plugin argocd-vault-plugin`

#### With Helm
If you want to use Helm along with argocd-vault-plugin, register a plugin in the `argocd-cm` ConfigMap like this:
```yaml
configManagementPlugins: |
  - name: argocd-vault-plugin-helm
    init:
      command: [sh, -c]
      args: ["helm dependency build"]
    generate:
      command: ["sh", "-c"]
      args: ["helm template $ARGOCD_APP_NAME . | argocd-vault-plugin generate -"]
```

If you want to use Helm along with argocd-vault-plugin and use additional helm args :
```yaml
configManagementPlugins: |
  - name: argocd-vault-plugin-helm
    init:
      command: [sh, -c]
      args: ["helm dependency build"]
    generate:
      command: ["sh", "-c"]
      args: ["helm template $ARGOCD_APP_NAME ${helm_args} . | argocd-vault-plugin generate -"]
```
Helm args must be defined in the application manifest:
```yaml
  source:
    path: your-app
    plugin:
      name: argocd-vault-plugin-helm
      env:
        - name: helm_args
          value: -f values-dev.yaml -f values-dev-tag.yaml
```

#### With Kustomize
If you want to use Kustomize along with argocd-vault-plugin, register a plugin in the `argocd-cm` ConfigMap like this:
```yaml
configManagementPlugins: |
  - name: argocd-vault-plugin-kustomize
    generate:
      command: ["sh", "-c"]
      args: ["kustomize build . | argocd-vault-plugin generate -"]
```

#### With Jsonnet
If you want to use Jsonnet along with argocd-vault-plugin, register a plugin in the `argocd-cm` ConfigMap like this:

```yaml
configManagementPlugins: |
  - name: argocd-vault-plugin-jsonnet
    generate:
      command: ["sh", "-c"]
      args: ["jsonnet . | argocd-vault-plugin generate -"]
```

The plugin will work with both YAML and JSON output from jsonnet.

#### Refreshing values from Secrets Managers
If you want to load in a new value from your Secret Manager without making any new code changes you must use the Hard-Refresh concept in Argo CD. This can be done in two ways. You can either use the UI and select the `Hard Refresh` button which is located within the `Refresh Button`.

<img src="https://github.com/IBM/argocd-vault-plugin/raw/main/assets/hard-refresh.png" width="300">  

You can also use the `argocd app diff` command passing the `--hard-refresh` flag. This will run argocd-vault-plugin again and pull in the new values from you Secret Manager and then you can either have Auto Sync setup or Sync manually to apply the new values.

### Caching the Vault Token
The plugin tries to cache the Vault token obtained from logging into Vault on the `argocd-repo-server`'s container's disk, at `~/.avp/config.json` for the duration of the token's lifetime. This of course requires that the container user is able to write to that path. Some environments, like Openshift, will force a random user for containers to run with; therefore this feature will not work, and the plugin will attempt to login to Vault on every run. This can be fixed by ensuring the `argocd-repo-server`'s container runs with the user `argocd`.
