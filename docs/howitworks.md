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

#### Modifiers
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
