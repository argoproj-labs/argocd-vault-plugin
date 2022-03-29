There are multiple ways to download and install argocd-vault-plugin depending on your use case.

#### On Linux or macOS via Curl
```
curl -Lo argocd-vault-plugin https://github.com/argoproj-labs/argocd-vault-plugin/releases/download/{version}/argocd-vault-plugin_{version}_{linux|darwin}_{amd64|arm64|s390x}

chmod +x argocd-vault-plugin

mv argocd-vault-plugin /usr/local/bin
```

#### On macOS via Homebrew

```
brew install argocd-vault-plugin
```

#### Installing in Argo CD

In order to use the plugin in Argo CD you can add it to your Argo CD instance as a volume mount or build your own Argo CD image.

The Argo CD docs provide information on how to get started <https://argoproj.github.io/argo-cd/operator-manual/custom_tools/>.

*Note*: We have provided a Kustomize app that will install Argo CD and configure the plugin [here](https://github.com/argoproj-labs/argocd-vault-plugin/blob/main/manifests/).

##### InitContainer
The first technique is to use an init container and a volumeMount to copy a different version of a tool into the repo-server container.
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: argocd-repo-server
spec:
  template:
    spec:
      containers:
      - name: argocd-repo-server
        volumeMounts:
        - name: custom-tools
          mountPath: /usr/local/bin/argocd-vault-plugin
          subPath: argocd-vault-plugin

        # Note: AVP config (for the secret manager, etc) can be passed in several ways. This is just one example
        # https://argocd-vault-plugin.readthedocs.io/en/stable/config/
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

        # Don't forget to update this to whatever the stable release version is
        # Note the lack of the `v` prefix unlike the git tag
        env:
          - name: AVP_VERSION
            value: "1.7.0"
        args:
          - >-
            wget -O argocd-vault-plugin
            https://github.com/argoproj-labs/argocd-vault-plugin/releases/download/v${AVP_VERSION}/argocd-vault-plugin_${AVP_VERSION}_linux_amd64 &&
            chmod +x argocd-vault-plugin &&
            mv argocd-vault-plugin /custom-tools/
        volumeMounts:
          - mountPath: /custom-tools
            name: custom-tools

      # Not strictly necessary, but required for passing AVP configuration from a secret and for using Kubernetes auth to Hashicorp Vault
      automountServiceAccountToken: true
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
ENV AVP_VERSION=0.2.2
ENV BIN=argocd-vault-plugin
RUN curl -L -o ${BIN} https://github.com/IBM/argocd-vault-plugin/releases/download/v${AVP_VERSION}/argocd-vault-plugin_${AVP_VERSION}_linux_amd64
RUN chmod +x ${BIN}
RUN mv ${BIN} /usr/local/bin

# Switch back to non-root user
USER 999
```
After making the plugin available, you must then register the plugin, documentation can be found at <https://argoproj.github.io/argo-cd/user-guide/config-management-plugins/#plugins> on how to do that.

For this plugin, you would add this:
```yaml
data:
  configManagementPlugins: |-
    - name: argocd-vault-plugin
      generate:
        command: ["argocd-vault-plugin"]
        args: ["generate", "./"]
```

You can use ArgoCD Vault Plugin along with other Kubernetes configuration tools (Helm, Kustomize, etc). The general method is to have your configuration tool output YAMLs that are ready to apply to a cluster except for containing `<placeholder>`s, and then run the plugin on this output to fill in the secrets. See the [Usage page](../usage) for examples.
