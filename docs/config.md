There are 3 different ways that parameters can be passed along to argocd-vault-plugin.

##### Kubernetes Secret

You can define a Secret with the Vault configuration. The keys of the secret's `data`/`stringData`
should be the exact names given below, case-sensitive:

```yaml
apiVersion: v1
data:
  VAULT_ADDR: Zm9v
  AVP_AUTH_TYPE: Zm9v
  AVP_GITHUB_TOKEN: Zm9v
  AVP_TYPE: Zm9v
kind: Secret
metadata:
  name: vault-configuration
  namespace: argocd
type: Opaque
```

You can use it like this: `argocd-vault-plugin generate /some/path -s vault-configuration`.

By default, the secret is assumed to be in the `argocd` namespace. However, the namespace containing the secret can be provided by using the format `<namespace>:<name>`

<b>Note</b>: this requires the `argocd-repo-server` to have a service account token mounted in the standard location.

###### ArgoCD 2.4.0 Environment Variable Prefix

Starting with ArgoCD 2.4.0, environment variables passed into the `init` and `generate` steps are prefixed with `ARGOCD_ENV` to prevent users from setting potentially-sensitive environment variables. All environment variables defined here will be prepended with the new prefix, e.g. `ARGOCD_ENV_AVP_TYPE`. The configuration will honor both prefixed and non-prefixed environment variables, preferring the prefixed variable if both are presented. There are no changes needed to the secret.

```yaml
apiVersion: v1
stringData:
  # Will be renamed to ARGOCD_ENV_AVP_AUTH_TYPE by ArgoCD before reaching the plugin.
  AVP_AUTH_TYPE: vault
kind: Secret
metadata:
  name: vault-configuration
  namespace: argocd
type: Opaque
```

See the [ArgoCD Upgrade Guide](https://argo-cd.readthedocs.io/en/latest/operator-manual/upgrading/2.3-2.4/#update-plugins-to-use-newly-prefixed-environment-variables) for more information.
##### Configuration File

The configuration can be given in a file reachable from the plugin, in any Viper supported format (YAML, JSON, etc.). The keys must match the same names used in the the Kubernetes secret:

```yaml
VAULT_ADDR: http://vault
AVP_AUTH_TYPE: github
AVP_GITHUB_TOKEN: t0ke3n
AVP_TYPE: vault
```

You can use it like this: `argocd-vault-plugin generate /some/path -c /path/to/config/file.yaml`. This can be useful for use-cases not involving Argo CD.

##### Environment Variables

The configuration can be set via environment variables, where each key is prefixed by `AVP_`:

```shell
AVP_TYPE=vault # corresponds to TYPE key
```

Make sure that these environment variables are available to the plugin when running it, whether that is in Argo CD or as a CLI tool. Note that any _set_
environment variables take precedence over configuration pulled from a Kubernetes Secret or a file.

### Full List of Supported Parameters
We support all the backend specific environment variables each backend's SDK will accept (e.g, `VAULT_NAMESPACE`, `AWS_REGION`, etc). Refer to the [specific backend's documentation](../backends) for details.

We also support these AVP specific variables:

| Name                       | Description                                         | Notes                                                                                                                                                                        |
| -------------------------- |-----------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| AVP_TYPE                   | The type of Vault backend                           | Supported values: `vault`, `ibmsecretsmanager`, `awssecretsmanager`, `gcpsecretmanager`, `yandexcloudlockbox` and `1passwordconnect`                                         |
| AVP_KV_VERSION             | The vault secret engine                             | Supported values: `1` and `2` (defaults to 2). KV_VERSION will be ignored if the `avp.kubernetes.io/kv-version` annotation is present in a YAML resource.                    |
| AVP_AUTH_TYPE              | The type of authentication                          | Supported values: vault: `approle, github, k8s, token`. Only honored for `AVP_TYPE` of `vault`                                                                               |
| AVP_GITHUB_TOKEN           | Github token                                        | Required with `AUTH_TYPE` of `github`                                                                                                                                        |
| AVP_ROLE_ID                | Vault AppRole Role_ID                               | Required with `AUTH_TYPE` of `approle`                                                                                                                                       |
| AVP_SECRET_ID              | Vault AppRole Secret_ID                             | Required with `AUTH_TYPE` of `approle`                                                                                                                                       |
| AVP_MOUNT_PATH             | Vault Auth Mount PATH                               | Optional. Defaults to the appropriate path based on `AUTH_TYPE` (i.e, `auth/approle` for AppRole authentication, `auth/github` for Github, `auth/kubernetes` for Kubernetes) |
| AVP_K8S_MOUNT_PATH         | Kuberentes Auth Mount PATH                          | Optional for `AUTH_TYPE` of `k8s` defaults to `auth/kubernetes`. Takes precedence over `$AVP_MOUNT_PATH`                                                                     |
| AVP_K8S_ROLE               | Kuberentes Auth Role                                | Required with `AUTH_TYPE` of `k8s`                                                                                                                                           |
| AVP_K8S_TOKEN_PATH         | Path to JWT for Kubernetes Auth                     | Optional for `AUTH_TYPE` of `k8s` defaults to `/var/run/secrets/kubernetes.io/serviceaccount/token`                                                                          |
| AVP_IBM_API_KEY            | IBM Cloud IAM API Key                               | Required with `TYPE` of `ibmsecretsmanager`                                                                                                                                  |
| AVP_IBM_INSTANCE_URL       | Endpoint URL for IBM Cloud Secrets Manager instance | If absent, fall back to `$VAULT_ADDR`                                                                                                                                        |
| AWS_REGION                 | AWS Secrets Manager Region                          | Only valid with `TYPE` `awssecretsmanager`                                                                                                                                   |
| AVP_YCL_SERVICE_ACCOUNT_ID | Yandex Cloud Lockbox service account ID             | Required with `TYPE` of `yandexcloudlockbox`                                                                                                                                 |
| AVP_YCL_KEY_ID             | Yandex Cloud Lockbox service account Key ID         | Required with `TYPE` of `yandexcloudlockbox`                                                                                                                                 |
| AVP_YCL_PRIVATE_KEY        | Yandex Cloud Lockbox service account private key    | Required with `TYPE` of `yandexcloudlockbox`                                                                                                                                 |
| AVP_PATH_VALIDATION        | Regular Expression to validate the Vault path       | Optional. Can be used for e.g. to prevent path traversals.                                                                                                                   |

### Full List of Supported Annotation

We support several different annotations that can be used inside a kubernetes resource. These annotations will override any corresponding configuration set via Environment Variable or Configuration File.

| Annotation                       | Description                                                                                                                                        |
| -------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| avp.kubernetes.io/path           | Path to the Vault Secret                                                                                                                           |
| avp.kubernetes.io/ignore         | Boolean to tell the plugin whether or not to process the file. Invalid values translate to `false`                                                 |
| avp.kubernetes.io/kv-version     | Version of the KV Secret Engine                                                                                                                    |
| avp.kubernetes.io/secret-version | Version of the secret to retrieve. Only effective on generic `<placeholder>`s so `avp.kubernetes.io/path` is required when this annotation is used |
| avp.kubernetes.io/remove-missing | Plugin will not throw error when a key is missing from Vault Secret. Only works on `Secret` or `ConfigMap` resources                               |

### Multitenancy

A common use-case is to be able to use _multiple_ secret backends for generating secrets within different Argo CD applications, such as when a team hosts a multi-tenant Argo CD instance.

For this to work, AVP must be configured to use specific credentials for generating the manifests of an app. This can be done in one of 2 ways:

#### Using Kubernetes secrets for supplying AVP configuration

This method requires having one Kubernetes secret with AVP configuration for each backend. For example, if there are 2 teams `foo` and `bar` using different instances of AWS Secret Manager, there should be at least 2 Kubernetes secrets containing AVP configuration: `foo-team-aws-sm-credentials` and `bar-team-aws-sm-credentials`.

Then, AVP can be registered as a config management plugin in `argocd-cm` like this:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-cm
data:
  configManagementPlugins: |
    - name: aws-avp
      generate:
        command: ["sh", "-c"]
        args: ["argocd-vault-plugin generate -s ${AVP_SECRET} ./"]
```

Notice that the secret name is parametrized via an environment variable. This means each Argo app manifest can set `AVP_SECRET` to be the name of the Kubernetes secret that contains the configuration for the backend needed to generate its secrets.

With the above setup, team `foo` can then deploy an Argo app like this:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-nginx
spec:
  destination:
    namespace: default
    server: https://kubernetes.default.svc
  project: default
  source:
    repoURL: 'https://github.com/jkayani/avp-demo-kubecon-2021'
    targetRevision: HEAD
    path: apps/git/nginx/manifests
    plugin:
      name: aws-avp
      env:
        - name: AVP_SECRET
          value: foo-team-aws-sm-credentials
```

The above is just one approach. If there is a select number of different backends, registering AVP in multiple config management plugins, each using the appropriate backend, is also a valid solution. The `argocd-cm` configuration would look like this:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-cm
data:
  configManagementPlugins: |

    - name: foo-aws-avp
      generate:
        command: ["sh", "-c"]
        args: ["argocd-vault-plugin generate -s foo-team-aws-sm-credentials ./"]

    - name: bar-aws-avp
      generate:
        command: ["sh", "-c"]
        args: ["argocd-vault-plugin generate -s bar-team-aws-sm-credentials ./"]
```

The `foo` team would then deploy an Argo app like this:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-nginx
spec:
  destination:
    namespace: default
    server: https://kubernetes.default.svc
  project: default
  source:
    repoURL: 'https://github.com/jkayani/avp-demo-kubecon-2021'
    targetRevision: HEAD
    path: apps/git/nginx/manifests
    plugin:
      name: foo-aws-avp
```

#### Passing AVP configuration as environment variables in the app manifest

This method simply requires passing the appropriate AVP configuration environment variables in the Argo CD app manifest. This is best used when non-sensitive data, like the Vault namespace, is the only thing that varies between tenants (and therefore, Vault configuration variables like `AVP_ROLE_ID` and `AVP_SECRET_ID` can be specified once via environment variables in the `argocd-repo-server` pod, Kubernetes secrets, etc and not put directly in the app manifest).

For example, if there are 2 teams `foo` and `bar` that use the same Vault but different namespaces, the only configuration that needs to be specified per manifest is the `VAULT_NAMESPACE`.

The plugin can be registered like normal in `argocd-cm`:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-cm
data:
  configManagementPlugins: |
    - name: vault-avp
      generate:
        command: ["argocd-vault-plugin"]
        args: ["generate", "./"]
```

In the app manifest, team `foo` just needs to set `VAULT_NAMESPACE` to the appropriate value:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-nginx
spec:
  destination:
    namespace: default
    server: https://kubernetes.default.svc
  project: default
  source:
    repoURL: 'https://github.com/jkayani/avp-demo-kubecon-2021'
    targetRevision: HEAD
    path: apps/git/nginx/manifests
    plugin:
      name: vault-avp
      env:
        - name: VAULT_NAMESPACE
          value: foo-team-namespace
```

**Note**: Exposing tokens (like `AVP_ROLE_ID` or `AVP_SECRET_ID`) in plain-text in Argo CD app manifests should be avoided. Prefer to pass those tokens through one of the means mentioned above.