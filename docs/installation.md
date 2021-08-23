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
      args: ["helm template $ARGOCD_APP_NAME . > all.yaml && argocd-vault-plugin generate all.yaml"]
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
      args: ["helm template $ARGOCD_APP_NAME ${helm_args} . > all.yaml && argocd-vault-plugin generate all.yaml"]
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

If you are using Kustomize add:
```yaml
configManagementPlugins: |
  - name: argocd-vault-plugin-kustomize
    generate:
      command: ["sh", "-c"]
      args: ["kustomize build . > all.yaml && argocd-vault-plugin generate all.yaml"]
```

to the `argocd-cm` configMap.

Or if you are using Jsonnet add:
```yaml
configManagementPlugins: |
  - name: argocd-vault-plugin-jsonnet
    generate:
      command: ["sh", "-c"]
      args: ["jsonnet . | argocd-vault-plugin generate -"]
```

to the `argocd-cm` configMap.