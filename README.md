# argocd-vault-plugin
![Pipeline](https://github.com/IBM/argocd-vault-plugin/workflows/Pipeline/badge.svg) ![Code Scanning](https://github.com/IBM/argocd-vault-plugin/workflows/Code%20Scanning/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/IBM/argocd-vault-plugin)](https://goreportcard.com/report/github.com/IBM/argocd-vault-plugin) [![codecov](https://codecov.io/gh/IBM/argocd-vault-plugin/branch/main/graph/badge.svg?token=p8XJMcip6l)](https://codecov.io/gh/IBM/argocd-vault-plugin)

<img src="https://github.com/IBM/argocd-vault-plugin/raw/main/assets/argo_vault_logo.png" width="300">

An ArgoCD plugin to retrieve secrets from Hashicorp Vault and inject them into Kubernetes secrets

<details><summary>Table of Contents</summary>

- [Overview](#overview)
- [Installation](#installation)
    + [`Curl` command](#curl-command)
    + [Installing in ArgoCD](#installing-in-argocd)
        + [InitContainer](#initcontainer)
        + [Custom Image](#custom-image)
- [Using the Plugin](#using-the-plugin)
    + [Command Line](#command-line)
    + [ArgoCD](#argocd)
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
This plugin is aimed at helping to solve the issue of secret management with GitOps and ArgoCD. We wanted to find a simple way to utilize Vault without having to rely on an operator or custom resource definition. This plugin can be used not just for secrets but also for deployments, configMaps or any other Kubernetes resource.

### How it works
The argocd-vault-plugin works by taking a directory of yaml files that have been templated out using the pattern of `<placeholder>` where you would want a value from Vault to go. The inside of the `<>` would be the actual key in Vault.

An annotation can be used to specify exactly where the plugin should look for the vault values. The annotation needs to be in the format `avp_path: "path/to/secret"`.

For example, if you have a secret with the key `password-vault-key` that you would want to pull from vault, you might have a yaml that looks something like the below code. In this yaml, the plugin will pull the value of `path/to/secret/password-vault-key` and inject it into the secret yaml.

```
kind: Secret
apiVersion: v1
metadata:
  name: example-secret
  annotations:
    avp_path: "path/to/secret"
type: Opaque
data:
  password: <password-vault-key>
```

And then once the plugin is done doing the substitutions, it outputs the yaml to standard out to then be applied by Argo CD. The resulting yaml would look like:
```
kind: Secret
apiVersion: v1
metadata:
  name: example-secret
  annotations:
    avp_path: "path/to/secret"
type: Opaque
data:
  password: cGFzc3dvcmQK # The Value from the key password-vault-key in vault
```

The plugin also supports putting the path directly within the placeholder. The format must be `<path:path/to/secret#key>`, where `path/to/secret` is the vault path and the Vault key goes after the `#` symbol. Doing this does not require an `avp_path` annotation and will override any `avp_path` annotation that is set. For example:

```
kind: Secret
apiVersion: v1
metadata:
  name: example-secret
type: Opaque
data:
  password: <path:path/to/secret#password-vault-key>
```

<b>*Note*</b>: The plugin will attempt to read any strings that match the following PCRE regex: `<.*>` (any characters between matching angle brackets), in all YAML files at the path given as the `<path>` argument. If there are YAML files that use `<string>`'s for other purposes and should _not_ be replaced, you can tell AVP to skip that file by adding the annotation `avp_ignore: "true"`.

##### Modifiers
By Default the plugin does not perform any transformation of the secrets in transit. So if you have plain text secrets in Vault, you will need to use the `stringData` field and if you have a base64 encoded secret in Vault, you will need to use the `data` field according to the [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/secret/).

However, as of now, we support one modifier. And that is `base64encode`. So if you have a plain text value in Vault and would like to Base64 encode it on the fly to inject into a Kubernetes secret you can do:

```
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
There are multiple ways to download and install argocd-vault-plugin depedning on your use case.

#### `curl` command
Use curl or wget to download the binary from Github and then move the binary to `/usr/local/bin` or another directory that is in you `PATH`

```
curl -L -o argocd-vault-plugin https://github.com/IBM/argocd-vault-plugin/releases/download/{version}/argocd-vault-plugin_{version}_{linux|darwin}_amd64

chmod +x argocd-vault-plugin

mv argocd-vault-plugin /usr/local/bin
```

#### Installing in ArgoCD

In order to use the plugin in ArgoCD you can add it to your Argo CD instance as a volume mount or build your own Argo CD image.

The Argo CD docs provide information on how to get started https://argoproj.github.io/argo-cd/operator-manual/custom_tools/.

*Note*: We have provided a Kustomize app that will install ArgoCD and configure the plugin [here](https://github.com/IBM/argocd-vault-plugin/blob/main/manifests/).

##### InitContainer
The first technique is to use an init container and a volumeMount to copy a different version of a tool into the repo-server container.
```
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
    - wget -O argocd-vault-plugin
      https://github.com/IBM/argocd-vault-plugin/releases/download/v1.0.0/argocd-vault-plugin_1.0.0_linux_amd64

      chmod +x argocd-vault-plugin && mv argocd-vault-plugin /custom-tools/
  volumeMounts:
    - mountPath: /custom-tools
      name: custom-tools
```

##### Custom Image
The following example builds an entirely customized repo-server from a Dockerfile, installing extra dependencies that may be needed for generating manifests.
```
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
RUN curl -L -o argocd-vault-plugin https://github.com/IBM/argocd-vault-plugin/releases/download/v1.0.0/argocd-vault-plugin_1.0.0_linux_amd64
RUN chmod +x argocd-vault-plugin
RUN mv argocd-vault-plugin /usr/local/bin

# Switch back to non-root user
USER argocd
```
After making the plugin available, you must then register the plugin, documentation can be found at https://argoproj.github.io/argo-cd/user-guide/config-management-plugins/#plugins on how to do that.

For this plugin, you would add this:
```
data:
  configManagementPlugins: |-
    - name: argocd-vault-plugin
      generate:
        command: ["argocd-vault-plugin"]
        args: ["generate", "./"]
```

If you want to use Helm along with argocd-vault-plugin add:
```
configManagementPlugins: |
  - name: argocd-vault-plugin-helm
    generate:
      command: ["sh", "-c"]
      args: ["helm template . > all.yaml && argocd-vault-plugin generate all.yaml"]
```

Or if you are using Kustomize:
```
configManagementPlugins: |
  - name: argocd-vault-plugin-helm
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

### ArgoCD
Before using the plugin in ArgoCD you must follow the [steps](#installing-in-argocd) to install the plugin to your ArgoCD instance. Once the plugin is installed, you can use it 3 ways.

1. Select your plugin via the UI by selecting `New App` and then changing `Directory` at the bottom of the form to be `argocd-vault-plugin`.

2. Apply a ArgoCD Application yaml that has `argocd-vault-plugin` as the plugin.
```
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

  1. Configuring ArgoCD

  You can either use your own Service Account or the default ArgoCD service account. To use the default ArgoCD service account all you need to do is set `automountServiceAccountToken` to true in the `argocd-repo-server`.

  ```
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
  `kubectl create serviceaccount your-service-account`

  And then you will update the `argocd-repo-server` to use that service account.

  ```
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

Once ArgoCD and Kubernetes are configured, you can then set the required environment variables for the plugin:
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
These are the required parameters parameters for AWS:
```
AVP_TYPE: awssecretmanager
AWS_ACCESS_KEY_ID: Your AWS Access Key ID
AWS_SECRET_ACCESS_KEY: Your AWS Secret Access Key
```

## Configuration
There are 3 different ways that parameters can be passed along to argocd-vault-plugin.

##### Kubernetes Secret
You can define a Secret in the `argocd` namespace of your Argo CD cluster with the Vault configuration. The keys of the secret's `data`
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
The configuration can be given in a file reachable from the plugin, in any Viper supported format (YAML, JSON, etc.):
```yaml
VAULT_ADDR: Zm9v
AVP_AUTH_TYPE: Zm9v
AVP_GITHUB_TOKEN: Zm9v
AVP_TYPE: Zm9v
```
You can use it like this: `argocd-vault-plugin generate /some/path -c /path/to/config/file.yaml`. This can be useful for usecases not involving Argo CD.

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
| AVP_KV_VERSION    | The vault secret engine  | Supported values: `1` and `2` (defaults to 2). KV_VERSION will be ignored if the `kv_version` annotation is present in a YAML resource.|
| AVP_AUTH_TYPE      | The type of authentication | Supported values: vault: `approle, github`   secretmanager: `iam` |
| AVP_GITHUB_TOKEN   | Github token               | Required with `AUTH_TYPE` of `github` |
| AVP_ROLE_ID        | Vault AppRole Role_ID      | Required with `AUTH_TYPE` of `approle` |
| AVP_SECRET_ID      | Vault AppRole Secret_ID    | Required with `AUTH_TYPE` of `approle` |
| AVP_K8S_MOUNT_PATH | Kuberentes Auth Mount PATH | Optional for `AUTH_TYPE` of `k8s` defaults to `auth/kubernetes` |
| AVP_K8S_ROLE       | Kuberentes Auth Role      | Required with `AUTH_TYPE` of `k8s` |
| AVP_K8S_TOKEN_PATH | Path to JWT for Kubernetes Auth  | Optional for `AUTH_TYPE` of `k8s` defaults to `/var/run/secrets/kubernetes.io/serviceaccount/token` |
| AVP_IBM_API_KEY    | IBM Cloud IAM API Key      | Required with `TYPE` of `secretmanager` and `AUTH_TYPE` of `iam` |
| AWS_ACCESS_KEY_ID    | AWS Access Key ID      | Required with `TYPE` of `awssecretsmanager` |
| AWS_SECRET_ACCESS_KEY | AWS Secret Access Key      | Required with `TYPE` of `awssecretsmanager` |
| AVP_AWS_REGION    | AWS Secrets Manager Region      | Only valid with `TYPE` `awssecretsmanager` |

### Full List of Supported Annotation
We support two different annotations that can be used inside a kubernetes resource. These annotations will override any corresponding configuration set via Environment Variable or Configuration File.

| Annotation | Description |  
| ---------- | ----------- |  
| avp_path | Path to the Vault Secret |
| avp_ignore | Boolean to tell the plugin whether or not to process the file. Invalid values translate to `false` |
| kv_version | Version of the KV Secret Engine |

## Notes
- The plugin tries to cache the Vault token obtained from logging into Vault on the `argocd-repo-server`'s container's disk, at `~/.avp/config.json` for the duration of the token's lifetime. This of course requires that the container user is able to write to that path. Some environments, like Openshift 4, will force a random user for containers to run with; therefore this feature will not work, and the plugin will attempt to login to Vault on every run. This can be fixed by ensuring the `argocd-repo-server`'s container runs with the user `argocd`.

## Contributing
Interested in contributing? Please read our contributing documentation [here](./CONTRIBUTING.md) to get started!
