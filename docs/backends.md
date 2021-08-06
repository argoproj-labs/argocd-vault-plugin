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

### IBM Cloud Secrets Manager
For IBM Cloud Secret Manager we only support using IAM authentication at this time.

##### IAM Authentication
For IAM Authentication, these are the required parameters:
```
VAULT_ADDR: Your IBM Cloud Secret Manager Endpoint
AVP_TYPE: ibmsecretsmanager
AVP_AUTH_TYPE: iam
AVP_IBM_API_KEY: Your IBM Cloud API Key
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
