# argocd-vault-plugin
![Pipeline](https://github.com/IBM/argocd-vault-plugin/workflows/Pipeline/badge.svg)
![Code Scanning](https://github.com/IBM/argocd-vault-plugin/workflows/Code%20Scanning/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/IBM/argocd-vault-plugin)](https://goreportcard.com/report/github.com/IBM/argocd-vault-plugin)
![Downloads](https://img.shields.io/github/downloads/IBM/argocd-vault-plugin/total?logo=github)
[![codecov](https://codecov.io/gh/IBM/argocd-vault-plugin/branch/main/graph/badge.svg?token=p8XJMcip6l)](https://codecov.io/gh/IBM/argocd-vault-plugin)

<img src="https://github.com/IBM/argocd-vault-plugin/raw/main/assets/argo_vault_logo.png" width="300">

An Argo CD plugin to retrieve secrets from Hashicorp Vault and inject them into Kubernetes secrets

<details><summary>Table of Contents</summary>

- [Overview](#overview)
- [Installation](#installation)
    + [`Curl` command](#curl-command)
    + [Installing in Argo CD](#installing-in-argocd)
        + [InitContainer](#initcontainer)
        + [Custom Image](#custom-image)
- [Using the Plugin](#using-the-plugin)
    + [Command Line](#command-line)
    + [Argo CD](#argocd)
- [Backends](#backends)
    + [HashiCorp Vault](#hashicorp-vault)
        + [AppRole Authentication](#approle-authentication)
        + [Github Authentication](#github-authentication)
        + [Kubernetes Authentication](#kubernetes-authentication)
    + [IBM Cloud Secret Manager](#ibm-cloud-secret-manager)
        + [IAM Authentication](#iam-authentication)
- [Configuration](#configuration)
- [Notes](#notes)
- [Contributing](#contributing)

</details>

## Overview

### Why use this plugin?
This plugin is aimed at helping to solve the issue of secret management with GitOps and Argo CD. We wanted to find a simple way to utilize Vault without having to rely on an operator or custom resource definition. This plugin can be used not just for secrets but also for deployments, configMaps or any other Kubernetes resource.

### How it works
The argocd-vault-plugin works by taking a directory of yaml files that have been templated out using the pattern of `<placeholder>` where you would want a value from Vault to go. The inside of the `<>` would be the actual key in Vault.

An annotation can be used to specify exactly where the plugin should look for the vault values. The annotation needs to be in the format `avp.kubernetes.io/path: "path/to/secret"`.

For example, if you have a secret with the key `password-vault-key` that you would want to pull from vault, you might have a yaml that looks something like the below code. In this yaml, the plugin will pull the value of `path/to/secret/password-vault-key` and inject it into the secret yaml.

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: example-secret
  annotations:
    avp.kubernetes.io/path: "path/to/secret"
type: Opaque
data:
  password: <password-vault-key>
```

And then once the plugin is done doing the substitutions, it outputs the yaml to standard out to then be applied by Argo CD. The resulting yaml would look like:
```yaml
kind: Secret
apiVersion: v1
metadata:
  name: example-secret
  annotations:
    avp.kubernetes.io/path: "path/to/secret"
type: Opaque
data:
  password: cGFzc3dvcmQK # The Value from the key password-vault-key in vault
```

The plugin also supports putting the path directly within the placeholder. The format must be `<path:path/to/secret#key>`, where `path/to/secret` is the vault path and the Vault key goes after the `#` symbol. Doing this does not require an `avp.kubernetes.io/path` annotation and will override any `avp.kubernetes.io/path` annotation that is set. For example:

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: example-secret
type: Opaque
data:
  password: <path:path/to/secret#password-vault-key>
```

The plugin tries to be helpful and will ignore strings in the format `<string>` if the `avp.kubernetes.io/path` annotation is missing, and only consider strings in the format `<path:/path/to/secret#key>` as placeholders. This can be very useful when using AVP with YAML that uses `<string>`'s for other purposes, for example in CRD's with usage information:

```yaml
kind: CustomResourceDefinition
apiVersion: v1
metadata:
  name: some-crd

  # Notice, no `avp.kuberenetes.io/path` annotation here
  annotations: {}
type: Opaque
fieldRef:

  # So, <KEY> is NOT a placeholder
  description: 'Selects a field of
    the pod: supports metadata.name,
    metadata.namespace, `metadata.labels[''<KEY>'']`,
    `metadata.annotations[''<KEY>'']`,
    spec.nodeName, spec.serviceAccountName,
    status.hostIP, status.podIP, status.podIPs.'

  # But, THIS is still a placeholder
  some-credential: <path:somewhere/in/my/vault#credential>
```

In addition to the default behavior, the plugin will attempt to decode base64 encoded strings in Secrets to look for placeholders.  If a placeholder is found, it is replaced and the resulting string is re-base64 encoded.  In most cases it is not necessary to use the `base64encode` modifier in this scenario. In the following example the plugin will decode the `POSTGRES_URL` value, find the template `postgres://<username>:<password>@<host>:<port>/<database>?sslmode=require`, and base64 encode it after replacement.

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: example-secret
  annotations:
    avp.kubernetes.io/path: "path/to/secret"
type: Opaque
data:
  POSTGRES_URL: cG9zdGdyZXM6Ly88dXNlcm5hbWU+OjxwYXNzd29yZD5APGhvc3Q+Ojxwb3J0Pi88ZGF0YWJhc2U+P3NzbG1vZGU9cmVxdWlyZQ==
```

Finally, the plugin will ignore any given YAML file outright with the `avp.kubernetes.io/ignore` annotation set to `"true"`:

```yaml
kind: CustomResourceDefinition
apiVersion: v1
metadata:
  name: some-crd

  # Notice, `avp.kuberenetes.io/ignore` annotation is set
  annotations:
    avp.kuberenetes.io/ignore: "true"
type: Opaque
fieldRef:

  # So, <KEY> is NOT a placeholder
  description: 'Selects a field of
    the pod: supports metadata.name,
    metadata.namespace, `metadata.labels[''<KEY>'']`,
    `metadata.annotations[''<KEY>'']`,
    spec.nodeName, spec.serviceAccountName,
    status.hostIP, status.podIP, status.podIPs.'

  # Neither is this
  some-credential: <path:somewhere/in/my/vault#credential
```

##### Modifiers
By default the plugin does not perform any transformation of the secrets in transit. So if you have plain text secrets in Vault, you will need to use the `stringData` field and if you have a base64 encoded secret in Vault, you will need to use the `data` field according to the [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/secret/).

However, as of now, we support one modifier. And that is `base64encode`. So if you have a plain text value in Vault and would like to Base64 encode it on the fly to inject into a Kubernetes secret you can do:

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: example-secret
type: Opaque
data:
  password: <path:path/to/secret#password-vault-key | base64encode>
```
And the plugin will pull the value from Vault, Base64 encode the value and then inject it into the placeholder.

## Installation
There are multiple ways to download and install argocd-vault-plugin depending on your use case.

#### On Linux or macOS via Curl
```
curl -Lo argocd-vault-plugin https://github.com/IBM/argocd-vault-plugin/releases/download/{version}/argocd-vault-plugin_{version}_{linux|darwin}_amd64

chmod +x argocd-vault-plugin

mv argocd-vault-plugin /usr/local/bin
```

#### On macOS via Homebrew

```
brew install argocd-vault-plugin
```

#### Installing in Argo CD

In order to use the plugin in Argo CD you can add it to your Argo CD instance as a volume mount or build your own Argo CD image.

The Argo CD docs provide information on how to get started https://argoproj.github.io/argo-cd/operator-manual/custom_tools/.

*Note*: We have provided a Kustomize app that will install Argo CD and configure the plugin [here](https://github.com/IBM/argocd-vault-plugin/blob/main/manifests/).

##### InitContainer
The first technique is to use an init container and a volumeMount to copy a different version of a tool into the repo-server container.
```yaml
containers:
- name: argocd-repo-server
  volumeMounts:
  - name: custom-tools
    mountPath: /usr/local/bin/argocd-vault-plugin
    subPath: argocd-vault-plugin
  envFrom:
    - secretRef:
        name: argocd-vault-plugin-credentials
volumes:
- name: custom-tools
  emptyDir: {}
initContainers:
- name: download-tools
  image: alpine:3.8
  command: [sh, -c]
  args:
    - >-
      wget -O argocd-vault-plugin
      https://github.com/IBM/argocd-vault-plugin/releases/download/v1.1.1/argocd-vault-plugin_1.1.1_linux_amd64 &&
      chmod +x argocd-vault-plugin &&
      mv argocd-vault-plugin /custom-tools/
  volumeMounts:
    - mountPath: /custom-tools
      name: custom-tools
```

##### Custom Image
The following example builds an entirely customized repo-server from a Dockerfile, installing extra dependencies that may be needed for generating manifests.
```Dockerfile
FROM argoproj/argocd:latest

# Switch to root for the ability to perform install
USER root

# Install tools needed for your repo-server to retrieve & decrypt secrets, render manifests
# (e.g. curl, awscli, gpg, sops)
RUN apt-get update && \
    apt-get install -y \
        curl \
        awscli \
        gpg && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Install the AVP plugin (as root so we can copy to /usr/local/bin)
RUN curl -L -o argocd-vault-plugin https://github.com/IBM/argocd-vault-plugin/releases/download/v1.1.1/argocd-vault-plugin_1.1.1_linux_amd64
RUN chmod +x argocd-vault-plugin
RUN mv argocd-vault-plugin /usr/local/bin

# Switch back to non-root user
USER argocd
```
After making the plugin available, you must then register the plugin, documentation can be found at https://argoproj.github.io/argo-cd/user-guide/config-management-plugins/#plugins on how to do that.

For this plugin, you would add this:
```yaml
data:
  configManagementPlugins: |-
    - name: argocd-vault-plugin
      generate:
        command: ["argocd-vault-plugin"]
        args: ["generate", "./"]
```

If you want to use Helm along with argocd-vault-plugin add:
```yaml
configManagementPlugins: |
  - name: argocd-vault-plugin-helm
    init:
      command: [sh, -c]
      args: ["helm dependency build"]
    generate:
      command: ["sh", "-c"]
      args: ["helm template . > all.yaml && argocd-vault-plugin generate all.yaml"]
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
      args: ["helm template ${helm_args} . > all.yaml && argocd-vault-plugin generate all.yaml"]
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

Or if you are using Kustomize:
```yaml
configManagementPlugins: |
  - name: argocd-vault-plugin-kustomize
    generate:
      command: ["sh", "-c"]
      args: ["kustomize build . > all.yaml && argocd-vault-plugin generate all.yaml"]
```

to the `argocd-cm` configMap.

## Using the Plugin

### Command Line
The plugin can be used via the command line or any shell script. Since the plugin outputs yaml to standard out, you can run the `generate` command and pipe the output to `kubectl`.

`argocd-vault-plugin generate ./ | kubectl apply -f -`

This will pull the values from Vault, replace the placeholders and then apply the yamls to whatever kubernetes cluster you are connected to.

You can also read from stdin like so:

`cat example.yaml | argocd-vault-plugin generate - | kubectl apply -f -`

### Argo CD
Before using the plugin in Argo CD you must follow the [steps](#installing-in-argocd) to install the plugin to your Argo CD instance. Once the plugin is installed, you can use it 3 ways.

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

## Backends
As of today argocd-vault-plugin supports HashiCorp Vault and IBM Cloud Secret Manager as backends.

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

## Configuration
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
You can use it like this: `argocd-vault-plugin generate /some/path -s vault-configuration`.
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
| AVP_TYPE           | The type of Vault backend  | Supported values: `vault`, `ibmsecretsmanager` and `awssecretsmanager` |
| AVP_KV_VERSION    | The vault secret engine  | Supported values: `1` and `2` (defaults to 2). KV_VERSION will be ignored if the `avp.kubernetes.io/kv-version` annotation is present in a YAML resource.|
| AVP_AUTH_TYPE      | The type of authentication | Supported values: vault: `approle, github, k8s`   ibmsecretsmanager: `iam` |
| AVP_GITHUB_TOKEN   | Github token               | Required with `AUTH_TYPE` of `github` |
| AVP_ROLE_ID        | Vault AppRole Role_ID      | Required with `AUTH_TYPE` of `approle` |
| AVP_SECRET_ID      | Vault AppRole Secret_ID    | Required with `AUTH_TYPE` of `approle` |
| AVP_K8S_MOUNT_PATH | Kuberentes Auth Mount PATH | Optional for `AUTH_TYPE` of `k8s` defaults to `auth/kubernetes` |
| AVP_K8S_ROLE       | Kuberentes Auth Role      | Required with `AUTH_TYPE` of `k8s` |
| AVP_K8S_TOKEN_PATH | Path to JWT for Kubernetes Auth  | Optional for `AUTH_TYPE` of `k8s` defaults to `/var/run/secrets/kubernetes.io/serviceaccount/token` |
| AVP_IBM_API_KEY    | IBM Cloud IAM API Key      | Required with `TYPE` of `ibmsecretsmanager` and `AUTH_TYPE` of `iam` |
| AWS_REGION    | AWS Secrets Manager Region      | Only valid with `TYPE` `awssecretsmanager` |

### Full List of Supported Annotation
We support several different annotations that can be used inside a kubernetes resource. These annotations will override any corresponding configuration set via Environment Variable or Configuration File.

| Annotation | Description |  
| ---------- | ----------- |  
| avp.kubernetes.io/path | Path to the Vault Secret |
| avp.kubernetes.io/ignore | Boolean to tell the plugin whether or not to process the file. Invalid values translate to `false` |
| avp.kubernetes.io/kv-version | Version of the KV Secret Engine |

## 0.x to 1.x Migration Guide
#### Annotation Changes
In order to follow Kubernetes annotations, we have updated the supported annotations

| Old        | New |  
| ---------- | ----------- |  
| avp_path   | avp.kubernetes.io/path  |
| avp_ignore | avp.kubernetes.io/ignore |
| kv_version | avp.kubernetes.io/kv-version |

#### AVP Prefix
The `AVP` prefix is now required for all configurations options not including `VAULT` environment variables (https://www.vaultproject.io/docs/commands#environment-variables).

#### PATH_PREFIX
The `PATH_PREFIX` environment variable has now been removed and is no longer available.

#### `secretmanager` is now `ibmsecretsmanager`
With the addition of `awssecretsmanager` we have renamed `secretmanager` to be `ibmsecretsmanager` to follow a more consistent naming convention.

## Notes
- The plugin tries to cache the Vault token obtained from logging into Vault on the `argocd-repo-server`'s container's disk, at `~/.avp/config.json` for the duration of the token's lifetime. This of course requires that the container user is able to write to that path. Some environments, like Openshift 4, will force a random user for containers to run with; therefore this feature will not work, and the plugin will attempt to login to Vault on every run. This can be fixed by ensuring the `argocd-repo-server`'s container runs with the user `argocd`.

## Contributing
Interested in contributing? Please read our contributing documentation [here](./CONTRIBUTING.md) to get started!
