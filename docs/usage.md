### Command Line
The plugin can be used via the command line or any shell script. Since the plugin outputs YAML to standard out, you can run the `generate` command and pipe the output to `kubectl`.

`argocd-vault-plugin generate ./ | kubectl apply -f -`

This will pull the values from the configured secret manager, replace the placeholders and then apply the YAMLs to whatever Kubernetes cluster you are connected to.

You can also read from stdin like so:

`argocd-vault-plugin generate - < example.yaml | kubectl apply -f -`

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
      name: argocd-vault-plugin # Don't include this if installing via sidecar
    repoURL: http://your-repo/
    targetRevision: HEAD
  project: default
```

3. Or you can pass the config-management-plugin flag to the Argo CD CLI app create command:  
`argocd app create you-app-name --config-management-plugin argocd-vault-plugin`

#### With Helm
If you want to use Helm along with argocd-vault-plugin, use the instructions matching your [plugin installation method](../installation).

##### Simple
For `argocd-cm` ConfigMap configured plugins, add this to `argod-cm` ConfigMap:
```yaml
configManagementPlugins: |
  - name: argocd-vault-plugin-helm
    init:
      command: [sh, -c]
      args: ["helm dependency build"]
    generate:
      command: ["sh", "-c"]
      args: ["helm template $ARGOCD_APP_NAME . --include-crds | argocd-vault-plugin generate -"]
```
For sidecar configured plugins, add this to `cmp-plugin` ConfigMap, and then [add a sidecar to run it](../installation#initcontainer-and-configuration-via-sidecar):
```yaml
  avp-helm.yaml: |
    ---
    apiVersion: argoproj.io/v1alpha1
    kind: ConfigManagementPlugin
    metadata:
      name: argocd-vault-plugin-helm
    spec:
      allowConcurrency: true
      discover:
        find:
          command:
            - sh
            - "-c"
            - "find . -name 'Chart.yaml' && find . -name 'values.yaml'"
      generate:
        command:
          - sh
          - "-c"
          - |
            helm template $ARGOCD_APP_NAME --include-crds . |
            argocd-vault-plugin generate -
      lockRepo: false
```

##### With additional Helm arguments

Use this option if you want to use Helm along with argocd-vault-plugin and use additional helm args.

**IMPORTANT**: passing `${ARGOCD_ENV_HELM_ARGS}` effectively allows users to run arbitrary code in the Argo CD 
repo-server (or, if using a sidecar, in the plugin sidecar). Only use this when the users are completely trusted. If  
possible, determine which Helm arguments are needed by your users and explicitly pass only those arguments.

For `argocd-cm` ConfigMap configured plugins, add this to `argod-cm` ConfigMap:
```yaml
configManagementPlugins: |
  - name: argocd-vault-plugin-helm
    init:
      command: [sh, -c]
      args: ["helm dependency build"]
    generate:
      command: ["sh", "-c"]
      args: ["helm template $ARGOCD_APP_NAME -n $ARGOCD_APP_NAMESPACE ${ARGOCD_ENV_HELM_ARGS} . --include-crds | argocd-vault-plugin generate -"]
```
For sidecar configured plugins, add this to `cmp-plugin` ConfigMap, and then [add a sidecar to run it](../installation#initcontainer-and-configuration-via-sidecar):
```yaml
  avp-helm.yaml: |
    ---
    apiVersion: argoproj.io/v1alpha1
    kind: ConfigManagementPlugin
    metadata:
      name: argocd-vault-plugin-helm
    spec:
      allowConcurrency: true
      discover:
        find:
          command:
            - sh
            - "-c"
            - "find . -name 'Chart.yaml' && find . -name 'values.yaml'"
      generate:
        command:
          - sh
          - "-c"
          - |
            helm template $ARGOCD_APP_NAME --include-crds -n $ARGOCD_APP_NAMESPACE ${ARGOCD_ENV_HELM_ARGS} . |
            argocd-vault-plugin generate -
      lockRepo: false
```

Helm args must be defined in the application manifest:
```yaml
  source:
    path: your-app
    plugin:
      name: argocd-vault-plugin-helm
      env:
        - name: HELM_ARGS
          value: -f values-dev.yaml -f values-dev-tag.yaml
```

**Note: Bypassing the parameters like this can be dangerous in a multi-tenant environment as it could allow for malicious injection of arbitrary commands. So be cautious when doing something like in a production environment. Ensuring proper permissions and protections is very important when doing something like this.** 

##### With an inline values file
Alternatively, if you'd like to use values inline in your application manifest (similar to the ArgoCD CLI's `--values-literal-file` option), you can create a plugin like this (note the use of `bash` instead of `sh` here):

For `argocd-cm` ConfigMap configured plugins, add this to `argod-cm` ConfigMap:
```yaml
configManagementPlugins: |
  - name: argocd-vault-plugin-helm
    generate:
      command: ["bash", "-c"]
      args: ['helm template "$ARGOCD_APP_NAME" -f <(echo "$ARGOCD_ENV_HELM_VALUES") . | argocd-vault-plugin generate -']
```
For sidecar configured plugins, add this to `cmp-plugin` ConfigMap, and then [add a sidecar to run it](../installation#initcontainer-and-configuration-via-sidecar):
```yaml
  avp-helm.yaml: |
    ---
    apiVersion: argoproj.io/v1alpha1
    kind: ConfigManagementPlugin
    metadata:
      name: argocd-vault-plugin-helm
    spec:
      allowConcurrency: true
      discover:
        find:
          command:
            - sh
            - "-c"
            - "find . -name 'Chart.yaml' && find . -name 'values.yaml'"
      generate:
        command:
          - bash
          - "-c"
          - |
            helm template $ARGOCD_APP_NAME -n $ARGOCD_APP_NAMESPACE -f <(echo "$ARGOCD_ENV_HELM_VALUES") . |
            argocd-vault-plugin generate -
      lockRepo: false
```

Then you can define your Helm values inline in your application manifest:
```yaml
  source:
    path: your-app
    plugin:
      name: argocd-vault-plugin-helm
      env:
        - name: HELM_VALUES
          value: |
            # non-vault helm values are specified normally
            someValue: lasldkfjlksa
            moreStuff:
              - a
              - b
              - c

            # get mysql credentials from kv2 vault secret at path "kv/mysql"
            mysql:
              username: <path:kv/data/mysql#user>
              password: <path:kv/data/mysql#password>
              hostname: <path:kv/data/mysql#hostname>
```

#### With Kustomize
If you want to use Kustomize along with argocd-vault-plugin, use the instructions matching your [plugin installation method](../installation).


For `argocd-cm` ConfigMap configured plugins, add this to `argod-cm` ConfigMap:
```yaml
configManagementPlugins: |
  - name: argocd-vault-plugin-kustomize
    generate:
      command: ["sh", "-c"]
      args: ["kustomize build . | argocd-vault-plugin generate -"]
```
For sidecar configured plugins, add this to `cmp-plugin` ConfigMap, and then [add a sidecar to run it](../installation#initcontainer-and-configuration-via-sidecar):
```yaml
  avp-kustomize.yaml: |
    ---
    apiVersion: argoproj.io/v1alpha1
    kind: ConfigManagementPlugin
    metadata:
      name: argocd-vault-plugin-kustomize
    spec:
      allowConcurrency: true
      discover:
        find:
          command:
            - find
            - "."
            - -name
            - kustomization.yaml
      generate:
        command:
          - sh
          - "-c"
          - "kustomize build . | argocd-vault-plugin generate -"
      lockRepo: false
```

#### With Jsonnet
If you want to use Jsonnet along with argocd-vault-plugin, use the instructions matching your [plugin installation method](../installation).


For `argocd-cm` ConfigMap configured plugins, add this to `argod-cm` ConfigMap:

```yaml
configManagementPlugins: |
  - name: argocd-vault-plugin-jsonnet
    generate:
      command: ["sh", "-c"]
      args: ["jsonnet . | argocd-vault-plugin generate -"]
```
For sidecar configured plugins, add this to `cmp-plugin` ConfigMap, and then [add a sidecar to run it](../installation#initcontainer-and-configuration-via-sidecar):
```yaml
  avp-jsonnet.yaml: |
    ---
    apiVersion: argoproj.io/v1alpha1
    kind: ConfigManagementPlugin
    metadata:
      name: argocd-vault-plugin-kustomize
    spec:
      allowConcurrency: true
      discover:
        find:
          command:
            - find
            - "."
            - -name
            - *.json
      generate:
        command:
          - sh
          - "-c"
          - "jsonnet . | argocd-vault-plugin generate -"
      lockRepo: false
```

The plugin will work with both YAML and JSON output from jsonnet.

#### Refreshing values from Secrets Managers
If you want to load in a new value from your Secret Manager without making any new code changes you must use the Hard-Refresh concept in Argo CD. This can be done in two ways. You can either use the UI and select the `Hard Refresh` button which is located within the `Refresh Button`.

<img src="https://github.com/argoproj-labs/argocd-vault-plugin/raw/main/assets/hard-refresh.png" width="300">  

You can also use the `argocd app diff` command passing the `--hard-refresh` flag. This will run argocd-vault-plugin again and pull in the new values from your Secret Manager and then you can either have Auto Sync setup or Sync manually to apply the new values.

### Caveats

#### Caching the Hashicorp Vault Token
The plugin tries to cache the Vault token obtained from logging into Vault on the `argocd-repo-server`'s container's disk, at `~/.avp/config.json` for the duration of the token's lifetime. This of course requires that the container user is able to write to that path. Some environments, like Openshift, will force a random user for containers to run with; therefore this feature will not work, and the plugin will attempt to login to Vault on every run. This can be fixed by ensuring the `argocd-repo-server`'s container runs with the user `argocd`.

#### Running argocd-vault-plugin in a sidecar container
As mentioned in the [Installation page](../installation), Argo CD has a newer method of installing custom plugins via sidecar containers to the `argocd-repo-server` deployment. Here are some caveats with running in this configuration:

- Use an image that contains the binaries needed: if attempting to deploy Helm charts with argocd-vault-plugin, ensure the image either contains Helm and argocd-vault-plugin pre-installed, or has some logic to fetch it from somewhere. Unlike the `argocd-cm` ConfigMap based installation, the sidecar image is supplied by the user and is distinct from the one for the `argocd-repo-server`, which means previously pre-included binaries will be absent

- An image with common CA certificates is recommended so that errors with untrusted TLS certificates from trying to retrieve remote Helm charts or Kustomize bases is avoided

- Using `github` URLs (`github.com/*`) with Kustomize requires having `git` in the container, and will fail otherwise. Use a `raw.githubusercontent.com` URL as a workaround if installing `git` isn't an option

- The `argocd-repo-server` container has a default limit, configuration key named `server.repo.server.timeout.seconds` in `argocd-cm` ConfigMap, of 60 seconds - this may need to be increased if lots of placeholders have to be processed for a given Application. The sidecar also has a default timeout, environment variable named `ARGOCD_EXEC_TIMEOUT` in the sidecar, of 90 seconds. This may also need to be increased. Details here: <https://argo-cd.readthedocs.io/en/stable/user-guide/config-management-plugins/#using-a-cmp>
