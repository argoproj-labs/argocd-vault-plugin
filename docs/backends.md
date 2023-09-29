### HashiCorp Vault
We support AppRole, Token, Github, Kubernetes and Userpass Auth Method for getting secrets from Vault.

We currently support retrieving secrets from KV-V1 and KV-V2 backends.

**Note**: For KV-V2 backends, the path needs to be specified as `${vault-kvv2-backend-path}/data/{path-to-secret}` where `vault-kvv2-backend-path` is the path to the KV-V2 backend (usually just `secret`) and `path-to-secret` is the path to the secret in Vault.

##### AppRole Authentication
For AppRole Authentication, these are the required parameters:
```
VAULT_ADDR: Your HashiCorp Vault Address
AVP_TYPE: vault
AVP_AUTH_TYPE: approle
AVP_ROLE_ID: Your AppRole Role ID
AVP_SECRET_ID: Your AppRole Secret ID
```

##### Vault Token Authentication
For Vault Token Authentication, these are the required parameters:
```
VAULT_ADDR: Your HashiCorp Vault Address
VAULT_TOKEN: Your Vault token
AVP_TYPE: vault
AVP_AUTH_TYPE: token
```

This option may be the easiest to test with locally, depending on your Vault setup.

##### Github Authentication
For Github Authentication, these are the required parameters:
```
VAULT_ADDR: Your HashiCorp Vault Address
AVP_TYPE: vault
AVP_AUTH_TYPE: github
AVP_GITHUB_TOKEN: Your Github Personal Access Token
```

##### Kubernetes Authentication
In order to use Kubernetes Authentication a couple of things are required.

###### 1. Configuring Argo CD
You can either use your own Service Account or the default Argo CD service account. To use the default Argo CD service account all you need to do is set `automountServiceAccountToken` to true in the `argocd-repo-server`.

```yaml
kind: Deployment
apiVersion: apps/v1
metadata:
  name: argocd-repo-server
spec:
  template:
    spec:
      automountServiceAccountToken: true
```

This will put the Service Account token in the default path of `/var/run/secrets/kubernetes.io/serviceaccount/token`.

If you want to use your own Service Account, you would first create the Service Account.
`kubectl create serviceaccount your-service-account`.

<b>*Note*</b>: The service account that you use must have access to the Kubernetes TokenReview API. You can find the Vault documentation on configuring Kubernetes [here](https://www.vaultproject.io/docs/auth/kubernetes#configuring-kubernetes).

And then you will update the `argocd-repo-server` to use that service account.

```yaml
kind: Deployment
apiVersion: apps/v1
metadata:
  name: argocd-repo-server
spec:
  template:
    spec:
      serviceAccount: your-service-account
      automountServiceAccountToken: true
```

###### 2. Configuring Kubernetes
Use the /config endpoint to configure Vault to talk to Kubernetes. Use `kubectl cluster-info` to validate the Kubernetes host address and TCP port. For the list of available configuration options, please see the [API documentation](https://www.vaultproject.io/api/auth/kubernetes).

```
$ vault write auth/kubernetes/config \
    token_reviewer_jwt="<your service account JWT>" \
    kubernetes_host=https://192.168.99.100:<your TCP port or blank for 443> \
    kubernetes_ca_cert=@ca.crt
```

And then create a named role:
```
vault write auth/kubernetes/role/argocd \
    bound_service_account_names=your-service-account \
    bound_service_account_namespaces=argocd \
    policies=argocd \
    ttl=1h
```
This role authorizes the "vault-auth" service account in the default namespace and it gives it the default policy.

You can find the full documentation on configuring Kubernetes Authentication [here](https://www.vaultproject.io/docs/auth/kubernetes#configuration).

--- 

Once Argo CD and Kubernetes are configured, you can then set the required environment variables for the plugin:
```
VAULT_ADDR: Your HashiCorp Vault Address
AVP_TYPE: vault
AVP_AUTH_TYPE: k8s
AVP_K8S_MOUNT_PATH: Mount Path of your kubernetes Auth (optional)
AVP_K8S_ROLE: Your Kuberetes Auth Role
AVP_K8S_TOKEN_PATH: Path to JWT (optional)
```

##### Userpass Authentication
For Userpass Authentication, these are the required parameters:
```
VAULT_ADDR: Your HashiCorp Vault Address
AVP_TYPE: vault
AVP_AUTH_TYPE: userpass
AVP_USERNAME: Your Username
AVP_PASSWORD: Your Password
```

##### Examples

###### Path Annotation

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: vault-example
  annotations:
    avp.kubernetes.io/path: "secret/data/database"
type: Opaque
data:
  username: <username>
  password: <password>
```

###### Inline Path

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: vault-example
type: Opaque
data:
  username: <path:secret/data/database#username>
  password: <path:secret/data/database#password>
```

###### Versioned secrets

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: vault-example
  annotations:
    avp.kubernetes.io/path: "secret/data/database"
    avp.kubernetes.io/secret-version: "2" # 2 is the latest revision in this example
type: Opaque
data:
  username: <username>
  password: <password>
  username-current: <path:secret/data/database#username#2> # same as <username>
  password-current: <path:secret/data/database#password#2> # same as <password>
  username-old: <path:secret/data/database#username#1>
  password-old: <path:secret/data/database#password#1>
```

**Note**: Only Vault KV-V2 backends support versioning. Versions specified with a KV-V1 Vault will be ignored and the latest version will be retrieved.

### IBM Cloud Secrets Manager
For IBM Cloud Secret Manager we only support using IAM authentication at this time.

We support all types of secrets that can be retrieved from IBM Cloud Secret Manager. Please note:

- Secrets that are JSON data (i.e, non `arbitrary` secrets or an `arbitrary` secret with JSON `payload`) can have the select keys (i.e, the `username` in a `username_password` type secret) interpolated with the [jsonPath](./howitworks.md#jsonPath) modifier. Not all keys are available for extraction with `jsonPath`. Refer to the [IBM Cloud Secret Manager API docs](https://cloud.ibm.com/apidocs/secrets-manager#get-secret) for more details

##### IAM Authentication
For IAM Authentication, these are the required parameters:
```
AVP_IBM_INSTANCE_URL or VAULT_ADDR: Your IBM Cloud Secret Manager Endpoint
AVP_TYPE: ibmsecretsmanager
AVP_IBM_API_KEY: Your IBM Cloud API Key
```

##### Examples

###### Path Annotation

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: ibm-example
  annotations:
    avp.kubernetes.io/path: "ibmcloud/arbitrary/secrets/groups/123" # 123 represents your Secret Group ID
type: Opaque
data:
  username: <username>
  password: <password>
```

###### Inline Path

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: ibm-example
type: Opaque
data:
  username: <path:ibmcloud/arbitrary/secrets/groups/123#username>
  password: <path:ibmcloud/arbitrary/secrets/groups/123#password>
```

###### Non-arbitrary secret

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: ibm-example
  annotations:
    avp.kubernetes.io/path: "ibmcloud/imported_cert/secrets/groups/123" # 123 represents your Secret Group ID
type: Opaque
stringData:
  PUBLIC_CRT: |
    <my-cert-secret | jsonPath {.certificate}>
  PUBLIC_CRT_PREVIOUS: |
    <path:ibmcloud/imported_cert/secrets/groups/123#my-cert-secret#previous | jsonPath {.certificate}>
  PRIVATE_KEY: |
    <my-cert-secret | jsonPath {.private_key}>
```

### AWS Secrets Manager

##### AWS Authentication
Refer to the [AWS SDK for Go V2
documentation](https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/#specifying-credentials) for
supplying AWS credentials. Supported credentials and the order in which they are loaded are
described [here](https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/#specifying-credentials).

**Note About Region**
If you provide the full AWS ARN as the secret path, ex. `arn:aws:secretsmanager:us-east-1:123123123:secret:some-secret`, 
the region from the ARN (us-east-1) in this example, will take precedents over the AWS_REGION environment variable listed below. 

These are the parameters for AWS:
```
AVP_TYPE: awssecretsmanager
AWS_REGION: Your AWS Region (Optional: defaults to us-east-2)
```

##### Examples

###### Path Annotation

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: aws-example
  annotations:
    avp.kubernetes.io/path: "test-aws-secret" # The name of your AWS Secret
stringData:
  sample-secret: <test-secret>
type: Opaque
```

###### Inline Path

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: aws-example
stringData:
  sample-secret: <path:test-aws-secret#test-secret>
type: Opaque
```

###### Versioned secrets

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: aws-example
  annotations:
    avp.kubernetes.io/path: "some-path/secret"
    avp.kubernetes.io/secret-version: "AWSCURRENT"
stringData:
  sample-secret: <test-secret>
  sample-secret-again: <path:some-path/secret#test-secret#AWSCURRENT>
  sample-secret-old: <path:some-path/secret#test-secret#AWSPREVIOUS>
type: Opaque
```

###### Secret in the same account

The 'friendly' name of the secret can be used in this case.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: aws-example
stringData:
  sample-secret: <path:test-aws-secret#test-secret>
type: Opaque
```

###### Secret in a different account

The arn of the secret needs to be used in this case:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: aws-example
stringData:
  sample-secret: <path:arn:aws:secretsmanager:<REGION>:<ACCOUNT_NUMBER>:<SECRET_ID>#<key>>
type: Opaque
```

###### Retrieving of binary data

Since there is no way to set a key for binary type in AWS Secret Manager, set the `<key>` part to `SecretBinary` to retrieve binary data:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: aws-example
stringData:
  sample-secret: <path:arn:aws:secretsmanager:<REGION>:<ACCOUNT_NUMBER>:<SECRET_ID>#SecretBinary>
type: Opaque
```

**NOTE**
For cross account access there is the need to configure the correct permissions between accounts, please check:
https://aws.amazon.com/premiumsupport/knowledge-center/secrets-manager-share-between-accounts
https://docs.aws.amazon.com/secretsmanager/latest/userguide/auth-and-access_examples_cross.html

### GCP Secret Manager

##### GCP Authentication
Refer to the [Authentication Overview](https://cloud.google.com/docs/authentication) for Google Cloud APIs.

These are the parameters for GCP:
```
AVP_TYPE: gcpsecretmanager
```

##### Examples

###### Path Annotation

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
  annotations:
    avp.kubernetes.io/path: projects/12345678987/secrets/test-secret
type: Opaque
data:
  password: <test-secret>
```

###### Inline Path

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
type: Opaque
data:
  password: <path:projects/12345678987/secrets/test-secret#test-secret>
```

###### Versioned secrets

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
  annotations:
    avp.kubernetes.io/path: "projects/12345678987/secrets/test-secret"
    avp.kubernetes.io/secret-version: "latest"
type: Opaque
data:
  current-password: <password>
  current-password-again: <path:projects/12345678987/secrets/test-secret#password#latest>
  password-old: <path:projects/12345678987/secrets/test-secret#password#another-version-id>
```

### AZURE Key Vault

##### Azure Authentication
Refer to the [Use environment-based authentication](https://docs.microsoft.com/en-us/azure/developer/go/azure-sdk-authorization#use-environment-based-authentication) in the Azure SDK for Go.

Any secrets that are disabled in the key vault will be skipped if found.

For Azure, `path` is the unique name of your key vault.

**Note**: Versioning is only supported for inline paths.

**Note**: Due to the way the Azure backend works, templates that use _inline-path placeholders are more efficient_
(fewer HTTP calls and therefore lower chance of hitting rate limit) than generic placeholders.

These are the parameters for Azure:
```
AVP_TYPE: azurekeyvault
```

##### Examples

###### Path Annotation

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
  annotations:
    avp.kubernetes.io/path: "keyvault"
type: Opaque
data:
  password: <test-secret>
```

###### Inline Path

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
type: Opaque
data:
  password: <path:keyvault#test-secret>
```

###### Versioned secrets

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
type: Opaque
data:
  current-password: <path:keyvault#password>
  current-password-again: <path:keyvault#password#8f8da2e06c8240808ee439ff093803b5>
  password-old: <path:keyvault#password#33740fc26214497f8904d93f20f7db6d>
```

### SOPS
##### SOPS Authentication
Refer to the [SOPS project page](https://github.com/mozilla/sops) for authentication options/environment variables.

For SOPS, `path` is file path to a JSON or YAML file encrypted using SOPS  and `key` is a top level key in the document, `jsonpath` can be used to fetch subkeys.

**Note**: Versioning is not supported.

These are the parameters for SOPS:
```
AVP_TYPE: sops
```

##### Examples
Given a file encrypted with SOPS named `example.yaml` and containing the following data:
```yaml
test-secret: test-data
parent:
  child: value
```

###### Path Annotation

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
  annotations:
    avp.kubernetes.io/path: "example.yaml"
type: Opaque
data:
  password: <test-secret>
```

###### Inline Path

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
type: Opaque
data:
  password: <path:example.yaml#test-secret>
```

###### Sub key

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
  annotations:
    avp.kubernetes.io/path: "example.yaml"
type: Opaque
stringData:
  password: <parent | jsonPath {.child}>
```

### Yandex Cloud Lockbox
##### YCL Authentication
Refer to the [IAM overview](https://cloud.yandex.com/en/docs/iam/concepts/) for yandex cloud APIs authorization.

These are the parameters for YCL:
```
AVP_TYPE: yandexcloudlockbox
AVP_YCL_SERVICE_ACCOUNT_ID: Service account ID
AVP_YCL_KEY_ID: Service account authorized Key ID
AVP_YCL_PRIVATE_KEY: Service account authorized private key
```
##### Examples

###### Path Annotation

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
  annotations:
    avp.kubernetes.io/path: "secret-id"
type: Opaque
data:
  password: <key>
```

###### Inline Path

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
type: Opaque
data:
  password: <path:secret-id#key>
```

###### Versioned secrets

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
    avp.kubernetes.io/path: "secret-id"
    avp.kubernetes.io/secret-version: "version-id"
type: Opaque
data:
  current-password: <password>
  current-password-again: <path:secret-id#password#version-id>
  password-old: <path:secret-id#password#old-version-id>
```

### 1Password Connect

**Note**: The 1Password Connect backend does not support versioning, so specifying a version will be ignored.

##### 1Password Connect Authentication

Refer to the [1Password Secrets Automation overview](https://support.1password.com/secrets-automation/) for 1Password Connect usage.

These are the parameters for 1Password Connect:

```
AVP_TYPE: 1passwordconnect
OP_CONNECT_TOKEN: Your 1Password Connect access token
OP_CONNECT_HOST: The hostname of your 1Password Connect server
```

##### Examples

###### Path Annotation

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
  annotations:
    avp.kubernetes.io/path: "vaults/vault-uuid/items/item-uuid"
type: Opaque
data:
  password: <key>
```

###### Inline Path

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
type: Opaque
data:
  password: <path:vaults/vault-uuid/items/item-uuid#key>
```

### Keeper Secrets Manager

**Note**: The Keeper Secrets Manager backend does not support versioning, or annotations. It does not support injecting attached files.

#### Keeper Authentication

These are the parameters for Keeper:
```
AVP_TYPE: keepersecretsmanager
AVP_KEEPER_CONFIG_PATH: the path to the keeper configuration file on disk.
```

##### Examples
Examples assume that the secrets are not saved base64 encoded in the Secret Server.

###### Path Annotation

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
  annotations:
    avp.kubernetes.io/path: "secret-id"
type: Opaque
stringData:
  password: <key>
```

###### Inline Path

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
type: Opaque
data:
  password: <path:secret-id#key | base64encode>
```

### Delinea Secret Server

**Note**: The Delinea Secret Server backend does not support versioning.

##### Delinea Authentication
Refer to the [REST API documentation on your Delinea Secret Server](https://your-delinea-server/SecretServer/Documents/restapi/) for API authentication.

These are the parameters for Delinea:
```
AVP_TYPE: delineasecretserver
AVP_DELINEA_URL: The URL of the Dilenea Secret Server
AVP_DELINEA_USER: The account for authentication
AVP_DELINEA_PASSWORD: The password for authentication

Optional:
AVP_DELINEA_DOMAIN: The authentication domain (e.g. the Active Directory domain)
```
##### Examples
Examples assume that the secrets are not saved base64 encoded in the Secret Server.

###### Path Annotation

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
  annotations:
    avp.kubernetes.io/path: "secret-id"
type: Opaque
stringData:
  password: <key>
```

###### Inline Path

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
type: Opaque
data:
  password: <path:secret-id#key | base64encode>
```

### Kubernetes Secret

Inject values from any kubernetes secret

**Note**: The Kubernetes Secret backend does not support versioning

##### Kubernetes Secret Authentication

Backend inherits same in-cluster service-account as the plugin itself

These are the parameters for Kubernetes Secret:

```
AVP_TYPE: kubernetessecret
```

##### Examples

###### Path Annotation

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
  annotations:
    avp.kubernetes.io/path: "my-secret"
type: Opaque
data:
  password: <key>
```

###### Path Annotation With Namespace

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
  annotations:
    avp.kubernetes.io/path: "prod:my-secret"
type: Opaque
data:
  password: <key>
```

###### Inline Path

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
type: Opaque
data:
  password: <path:my-secret#key>
```

###### Inline Path With Namespace

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: test-secret
type: Opaque
data:
  password: <path:prod:my-secret#key>
```