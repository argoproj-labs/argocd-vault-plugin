### HashiCorp Vault
We support AppRole and Github Auth Method for getting secrets from Vault.

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

1. Configuring Argo CD
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

2. Configuring Kubernetes  
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

    You can find the full documentation on configuring Kubernetes Authentication [Here](vaultproject.io/docs/auth/kubernetes#configuration).


Once Argo CD and Kubernetes are configured, you can then set the required environment variables for the plugin:
```
VAULT_ADDR: Your HashiCorp Vault Address
AVP_TYPE: vault
AVP_AUTH_TYPE: k8s
AVP_K8S_MOUNT_PATH: Mount Path of your kubernetes Auth (optional)
AVP_K8S_ROLE: Your Kuberetes Auth Role
AVP_K8S_TOKEN_PATH: Path to JWT (optional)
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

Additionally, we only support secrets of type `arbitrary`, retrieved from a secret group. Since [`arbitrary` secrets are not versioned](https://cloud.ibm.com/apidocs/secrets-manager?code=go#get-secret-version), any version specified in a placeholder is ignored and the latest version is retrieved.

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

### AWS Secrets Manager

##### AWS Authentication
Refer to the [AWS go SDK README](https://github.com/aws/aws-sdk-go#configuring-credentials) for supplying AWS credentials.
Supported credentials and the order in which they are loaded are described [here](https://github.com/aws/aws-sdk-go/blob/v1.38.62/aws/session/doc.go#L22).

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
