There are 3 different ways that parameters can be passed along to argocd-vault-plugin.

##### Kubernetes Secret
You can define a Secret in the `argocd` namespace of your Argo CD cluster with the Vault configuration. The keys of the secret's `data`/`stringData`
should be the exact names given above, case-sensitive:
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
You can use it like this: `argocd-vault-plugin generate /some/path -s vault-configuration`. You can also use `-s namespace/vault-configuration` to specify 
the namespace, or `-s vault-configuration --service-account-namespace` to grab the secret from the service account's namespace.
<b>Note</b>: this requires the `argocd-repo-server` to have a service account token mounted in the standard location.

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
```
AVP_TYPE=vault # corresponds to TYPE key
```
Make sure that these environment variables are available to the plugin when running it, whether that is in Argo CD or as a CLI tool. Note that any _set_
environment variables take precedence over configuration pulled from a Kubernetes Secret or a file.

### Full List of Supported Parameters
We support all Vault Environment Variables listed [here](https://www.vaultproject.io/docs/commands#environment-variables) as well as:

| Name            | Description | Notes |
| --------------- | ----------- | ----- |
| AVP_TYPE           | The type of Vault backend  | Supported values: `vault`, `ibmsecretsmanager`, `awssecretsmanager` and `gcpsecretmanager` |
| AVP_KV_VERSION    | The vault secret engine  | Supported values: `1` and `2` (defaults to 2). KV_VERSION will be ignored if the `avp.kubernetes.io/kv-version` annotation is present in a YAML resource.|
| AVP_AUTH_TYPE      | The type of authentication | Supported values: vault: `approle, github, k8s`. Only honored for `AVP_TYPE` of `vault` |
| AVP_GITHUB_TOKEN   | Github token               | Required with `AUTH_TYPE` of `github` |
| AVP_ROLE_ID        | Vault AppRole Role_ID      | Required with `AUTH_TYPE` of `approle` |
| AVP_SECRET_ID      | Vault AppRole Secret_ID    | Required with `AUTH_TYPE` of `approle` |
| AVP_K8S_MOUNT_PATH | Kuberentes Auth Mount PATH | Optional for `AUTH_TYPE` of `k8s` defaults to `auth/kubernetes` |
| AVP_K8S_ROLE       | Kuberentes Auth Role      | Required with `AUTH_TYPE` of `k8s` |
| AVP_K8S_TOKEN_PATH | Path to JWT for Kubernetes Auth  | Optional for `AUTH_TYPE` of `k8s` defaults to `/var/run/secrets/kubernetes.io/serviceaccount/token` |
| AVP_IBM_API_KEY      | IBM Cloud IAM API Key      | Required with `TYPE` of `ibmsecretsmanager` |
| AVP_IBM_INSTANCE_URL | Endpoint URL for IBM Cloud Secrets Manager instance | If absent, fall back to `$VAULT_ADDR` |
| AWS_REGION    | AWS Secrets Manager Region      | Only valid with `TYPE` `awssecretsmanager` |

### Full List of Supported Annotation
We support several different annotations that can be used inside a kubernetes resource. These annotations will override any corresponding configuration set via Environment Variable or Configuration File.

| Annotation | Description |  
| ---------- | ----------- |  
| avp.kubernetes.io/path | Path to the Vault Secret |
| avp.kubernetes.io/ignore | Boolean to tell the plugin whether or not to process the file. Invalid values translate to `false` |
| avp.kubernetes.io/kv-version | Version of the KV Secret Engine |
| avp.kubernetes.io/secret-version | Version of the secret to retrieve. Only effective on generic `<placeholder>`s so `avp.kubernetes.io/path` is required when this annotation is used |
| avp.kubernetes.io/remove-missing | Plugin will not throw error when a key is missing from Vault Secret. Only works on `Secret` or `ConfigMap` resources |
