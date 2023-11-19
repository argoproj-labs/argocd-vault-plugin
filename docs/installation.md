## Installing in Argo CD

In order to use the plugin in Argo CD you have 4 distinct options:

- Installation via `argocd-cm` ConfigMap

    - Download AVP in a volume and control everything as Kubernetes manifests
        - Available as a pre-built Kustomize app: <https://github.com/argoproj-labs/argocd-vault-plugin/blob/main/manifests/cmp-configmap>

    - Create a custom `argocd-repo-server` image with AVP and supporting tools pre-installed

- Installation via a sidecar container [(new, starting with Argo CD v2.4.0)](https://argo-cd.readthedocs.io/en/stable/user-guide/config-management-plugins/#installing-a-cmp)

    - Download AVP and supporting tools into a volume and control everything as Kubernetes manifests, using an off-the-shelf sidecar image

        - Available as a pre-built Kustomize app: <https://github.com/argoproj-labs/argocd-vault-plugin/blob/main/manifests/cmp-sidecar>

    - Create a custom sidecar image with AVP and supporting tools pre-installed

### Explaining your options

First, the Argo CD docs provide valuable information on how to extend the `argocd-repo-server` with additonal tools or a custom built image: <https://argoproj.github.io/argo-cd/operator-manual/custom_tools/>.

Before version 2.4.0 of Argo CD, the only way to install AVP was as an additional binary that ran inside the `argocd-repo-server` container when specifically told by including the following YAML in an Application mainfest:
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
spec:
  ... other fields
  plugin:
    name: argocd-vault-plugin
```
This is a perfectly fine method and will continue to work as long as Argo CD supports it. 

However, the Argo CD project has another method of using custom plugins which involves defining a [sidecar container](https://kubernetes.io/docs/concepts/workloads/pods/#workload-resources-for-managing-pods) for each individual plugin (this is a different container from the `argocd-repo-server` and will be the context in which the plugin runs), and having Argo CD decide which plugin to use based on the plugin definition:
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
spec:
  ... other fields
  # No need to define `plugin` since Argo CD will figure it out!
```
There are some [security benefits to running this way](https://github.com/argoproj/argo-cd/issues/9083#issuecomment-1098517762), it may be [future proof](https://github.com/argoproj/argo-cd/issues/8117), and you don't have to explicitly tell Argo CD which plugin to use: it will auto-detect it, like it does for Helm or Kustomize based applications. On the other hand, it adds a bit more complexity and can make some argocd-vault-plugin integrations a bit trickier - see the [caveats section of the Usage page](../usage#running-argocd-vault-plugin-in-a-sidecar-container) for details.

### InitContainer and configuration via argocd-cm ConfigMap
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
            value: "1.18.0"
        args:
          - >-
            OS="$(uname | tr '[:upper:]' '[:lower:]')" && [ "$(uname -m)" = "aarch64" ] && ARCH="arm64" || ARCH="amd64" &&
            wget -O argocd-vault-plugin
            https://github.com/argoproj-labs/argocd-vault-plugin/releases/download/v${AVP_VERSION}/argocd-vault-plugin_${AVP_VERSION}_${OS}_${ARCH} &&
            chmod +x argocd-vault-plugin &&
            mv argocd-vault-plugin /custom-tools/
        volumeMounts:
          - mountPath: /custom-tools
            name: custom-tools

      # Not strictly necessary, but required for passing AVP configuration from a secret and for using Kubernetes auth to Hashicorp Vault
      automountServiceAccountToken: true
```

### Custom Image and configuration via argocd-cm ConfigMap
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
ENV AVP_VERSION=1.18.0
ENV BIN=argocd-vault-plugin
RUN curl -L -o ${BIN} https://github.com/argoproj-labs/argocd-vault-plugin/releases/download/v${AVP_VERSION}/argocd-vault-plugin_${AVP_VERSION}_linux_amd64
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

### InitContainer and configuration via sidecar

Define the plugin in a ConfigMap that will be mounted in the sidecar container
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cmp-plugin
data:
  avp.yaml: |
    apiVersion: argoproj.io/v1alpha1
    kind: ConfigManagementPlugin
    metadata:
      name: argocd-vault-plugin
    spec:
      allowConcurrency: true
      discover:
        find:
          command:
            - sh
            - "-c"
            - "find . -name '*.yaml' | xargs -I {} grep \"<path\\|avp\\.kubernetes\\.io\" {} | grep ."
      generate:
        command:
          - argocd-vault-plugin
          - generate
          - "."
      lockRepo: false
---
```

Patch the argocd-repo-server to add an initContainer to download argocd-vault-plugin and define the sidecar. You can change the image from `registry.access.redhat.com/ubi8` to whatever is desired, so long as it [contains the needed binaries](../usage#running-argocd-vault-plugin-in-a-sidecar-container)
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: argocd-repo-server
spec:
  template:
    spec:
      automountServiceAccountToken: true
      volumes:
        - configMap:
            name: cmp-plugin
          name: cmp-plugin
        - name: custom-tools
          emptyDir: {}
      initContainers:
      - name: download-tools
        image: registry.access.redhat.com/ubi8
        env:
          - name: AVP_VERSION
            value: 1.18.0
        command: [sh, -c]
        args:
          - >-
            curl -L https://github.com/argoproj-labs/argocd-vault-plugin/releases/download/v$(AVP_VERSION)/argocd-vault-plugin_$(AVP_VERSION)_linux_amd64 -o argocd-vault-plugin &&
            chmod +x argocd-vault-plugin &&
            mv argocd-vault-plugin /custom-tools/
        volumeMounts:
          - mountPath: /custom-tools
            name: custom-tools
      containers:
      - name: avp
        command: [/var/run/argocd/argocd-cmp-server]
        image: registry.access.redhat.com/ubi8
        securityContext:
          runAsNonRoot: true
          runAsUser: 999
        volumeMounts:
          - mountPath: /var/run/argocd
            name: var-files
          - mountPath: /home/argocd/cmp-server/plugins
            name: plugins
          - mountPath: /tmp
            name: tmp
          
          # Register plugins into sidecar
          - mountPath: /home/argocd/cmp-server/config/plugin.yaml
            subPath: avp.yaml
            name: cmp-plugin

          # Important: Mount tools into $PATH
          - name: custom-tools
            subPath: argocd-vault-plugin
            mountPath: /usr/local/bin/argocd-vault-plugin
```

### Custom Image and configuration via sidecar
Define the plugin in a ConfigMap that will be mounted in the sidecar container
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cmp-plugin
data:
  avp.yaml: |
    apiVersion: argoproj.io/v1alpha1
    kind: ConfigManagementPlugin
    metadata:
      name: argocd-vault-plugin
    spec:
      allowConcurrency: true
      discover:
        find:
          command:
            - sh
            - "-c"
            - "find . -name '*.yaml' | xargs -I {} grep \"<path\\|avp\\.kubernetes\\.io\" {} | grep ."
      generate:
        command:
          - argocd-vault-plugin
          - generate
          - "."
      lockRepo: false
---
```

Define a sidecar image from a suitable base
```Dockerfile
FROM registry.access.redhat.com/ubi8

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
ENV AVP_VERSION=1.18.0
ENV BIN=argocd-vault-plugin
RUN curl -L -o ${BIN} https://github.com/argoproj-labs/argocd-vault-plugin/releases/download/v${AVP_VERSION}/argocd-vault-plugin_${AVP_VERSION}_linux_amd64
RUN chmod +x ${BIN}
RUN mv ${BIN} /usr/local/bin

# Switch back to non-root user
USER 999
```

Patch the argocd-repo-server to define the sidecar
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: argocd-repo-server
spec:
  template:
    spec:
      automountServiceAccountToken: true
      volumes:
        - configMap:
            name: cmp-plugin
          name: cmp-plugin
      containers:
      - name: avp
        command: [/var/run/argocd/argocd-cmp-server]
        image: your-container-registry/your-custom-image
        securityContext:
          runAsNonRoot: true
          runAsUser: 999
        volumeMounts:
          - mountPath: /var/run/argocd
            name: var-files
          - mountPath: /home/argocd/cmp-server/plugins
            name: plugins
          - mountPath: /tmp
            name: tmp
          
          # Register plugins into sidecar
          - mountPath: /home/argocd/cmp-server/config/plugin.yaml
            subPath: avp.yaml
            name: cmp-plugin
```

## Installing locally
### On Linux or macOS via Curl
```
curl -Lo argocd-vault-plugin https://github.com/argoproj-labs/argocd-vault-plugin/releases/download/{version}/argocd-vault-plugin_{version}_{linux|darwin}_{amd64|arm64|s390x}

chmod +x argocd-vault-plugin

mv argocd-vault-plugin /usr/local/bin
```

### On macOS via Homebrew

```
brew install argocd-vault-plugin
```

## Security considerations

The Argo CD Vault Plugin injects secrets into Kubernetes manifests generated inside the Argo CD repo-server component.
Those manifests, and the secrets they contain, are cached in the Redis instance used by Argo CD. So they are available
to anyone with direct access to the Redis instance. The manifests are also accessible to anyone with direct access to 
the repo-server.

Mitigations:
1. Set up network policies to prevent direct access to Argo CD components (Redis and the repo-server). Make sure your 
   cluster supports those network policies and can actually enforce them.
2. Consider running Argo CD on its own cluster, with no other applications running on it.
3. [Enable password authentication on the Redis instance](https://github.com/argoproj/argo-cd/issues/3130) (currently
   only supported for non-HA Argo CD installations).
