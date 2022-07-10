### Summary

The argocd-vault-plugin works by taking a directory of YAML or JSON files that have been templated out using the pattern of `<placeholder>` where you would want a value from Vault to go. The inside of the `<>` would be the actual key in Vault.

An annotation can be used to specify exactly where the plugin should look for the vault values. The annotation needs to be in the format `avp.kubernetes.io/path: "path/to/secret"`.

For example, if you have a secret with the key `password-vault-key` that you would want to pull from vault, you might have a yaml that looks something like the below code. In this yaml, the plugin will pull the value of the _latest version_ of the secret at `path/to/secret/password-vault-key` and inject it into the Secret.

As YAML:
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

As JSON:
```json
{
  "kind": "Secret",
  "apiVersion": "v1",
  "metadata": {
    "name": "example-secret",
    "annotations": {
      "avp.kubernetes.io/path": "path/to/secret"
    }
  },
  "type": "Opaque",
  "data": {
    "password": "<password-vault-key>"
  }
}
```

And then once the plugin is done doing the substitutions, it outputs the manifest as YAML to standard out to then be applied by Argo CD. The resulting YAML would look like:
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

### Replacement behavior
By default the plugin does not perform any transformation of the secrets in transit. So if you have plain text secrets in Vault, you will need to use the `stringData` field and if you have a base64 encoded secret in Vault, you will need to use the `data` field according to the [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/secret/).

There are 2 exceptions to this:

- Placeholders that are in base64 format - see [Base64 placeholders](#base64-placeholders) for details

- Modifiers - see [Modifiers](#modifiers) for details

### Types of placeholders

#### Generic placeholders
The example in the Summary uses a _generic_ placeholder, which is just the name of the _key_ of the secret in the secrets manager you want to inject. All placeholders have to be keys in the _same_ secret in the secrets manager.

Valid examples:

- `<placeholder>`

##### Specifying the path of a secret
The only way to specify the path of a secret for generic placeholders is to use the `avp.kubernetes.io/path` annotation like this:
```yaml
kind: Secret
apiVersion: v1
metadata:
  name: example-secret
  annotations:
    avp.kubernetes.io/path: "path/to/secret"
```

##### Specifying the version of a secret
The only way to specify the version of a secret for generic placeholders is to use the `avp.kubernetes.io/secret-version` annotation like this:
```yaml
kind: Secret
apiVersion: v1
metadata:
  name: example-secret
  annotations:
    avp.kubernetes.io/secret-version: "2" # Requires at least 2 revisions to exist to work
```
**Note**: This ignored for secret managers that don't allow versioning, meaning the latest version is returned

#### Inline-path placeholders
An inline-path placeholder allows you to specify the path, key, and optionally, the version to use for a specific placeholder. This means you can inject values from _multiple distinct_ secrets in your secrets manager into the same YAML. 

Valid examples:

- `<path:some/path#secret-key>`
- `<path:some/path#secret-key#version>`

If the `version` is omitted (first example), the latest version of the secret is retrieved. 

##### Specifying the path of a secret
The only way to specify the path is in the placeholder itself: the string `path:` followed by the path in your secret manager to the secret. The `avp.kubernetes.io/path` annotation has _no effect_ on these placeholders.

##### Specifying the version of a secret
The only way to specify the version is in the placeholder itself: the string following the last `#` in the placeholder should be the ID of the version of the secret in your secret manager. The `avp.kubernetes.io/secret-version` annotation has _no effect_ on these placeholders.

**Note**: This ignored for secret managers that don't allow versioning, meaning the latest version is returned

#### Special behavior

##### Base64 placeholders
Some tools like Kustomize secret generator will create Secrets with `data` fields containing base64 encoded strings from the source files. If you try to use `<placeholder>`s in the source files, they will be output in a base64 format. 

The plugin can handle this case by finding any base64 encoded placeholders (either generic or inline-path), replace them, and re-base64 encode the result. 

For example, given this input:
```yaml
kind: Secret
apiVersion: v1
metadata:
  name: example-secret
  annotations:
    avp.kubernetes.io/path: "path/to/secret"
type: Opaque
data:
  # The base64 encoding of postgres://<username>:<password>@<host>:<port>/<database>?sslmode=require
  POSTGRES_URL: cG9zdGdyZXM6Ly88dXNlcm5hbWU+OjxwYXNzd29yZD5APGhvc3Q+Ojxwb3J0Pi88ZGF0YWJhc2U+P3NzbG1vZGU9cmVxdWlyZQ==
```

and these values for the secrets:
```
username: user
password: pass
host: host
port: 9443
database: my-db
```

the output is:
```yaml
kind: Secret
apiVersion: v1
metadata:
  name: example-secret
  annotations:
    avp.kubernetes.io/path: "path/to/secret"
type: Opaque
data:
  # The base64 encoding of postgres://user:pass@host:9443/my-db?sslmode=require
  POSTGRES_URL: cG9zdGdyZXM6Ly91c2VyOnBhc3NAaG9zdDo5NDQzL215LWRiP3NzbG1vZGU9cmVxdWlyZQ==
```

##### Automatically ignoring `<placeholder>` strings
The plugin tries to be helpful and will ignore strings in the format `<string>` if the `avp.kubernetes.io/path` annotation is missing, and only try to replace [inline-path placeholders](#inline-path-placeholders)

This can be very useful when using AVP with YAML/JSON that uses `<string>`'s for other purposes, for example in CRD's with usage information:
```yaml
kind: CustomResourceDefinition
apiVersion: v1
metadata:
  name: some-crd

  # Notice, no `avp.kubernetes.io/path` annotation here
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

##### Ignoring entire YAML/JSON files
The plugin will ignore any given YAML/JSON file outright with the `avp.kubernetes.io/ignore` annotation set to `"true"`:

```yaml
kind: CustomResourceDefinition
apiVersion: v1
metadata:
  name: some-crd

  # Notice, `avp.kubernetes.io/ignore` annotation is set
  annotations:
    avp.kubernetes.io/ignore: "true"
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
  some-credential: <path:somewhere/in/my/vault#credential>

  # Nor this
  some-credential: <path:somewhere/in/my/vault#credential#version3>
```

##### Removing keys with missing values
By default, AVP will return an error if there is a `<placeholder>` that has no matching key in the secrets manager. 

You can override this by using the annotation `avp.kubernetes.io/remove-missing`. This will remove keys whose values are missing from Vault from the entire YAML. 

For example, given this input:
```yaml
kind: Secret
apiVersion: v1
metadata:
  name: example-secret
  annotations:
    avp.kubernetes.io/remove-missing: "true"
stringData:
  username: <username>
  password: <pass>
```

And this data in the secrets manager:
```
username: user
```

output is:
```yaml
kind: Secret
apiVersion: v1
metadata:
  name: example-secret
  annotations:
    avp.kubernetes.io/remove-missing: "true"
stringData:
  username: user
```
This only works with _generic_ placeholders.

#### Modifiers

##### `base64encode`
<!-- By default the plugin does not perform any transformation of the secrets in transit. So if you have plain text secrets in Vault, you will need to use the `stringData` field and if you have a base64 encoded secret in Vault, you will need to use the `data` field according to the [Kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/secret/). -->

The base64encode modifier allows you to base64 encode a plain-text value retrieved from a secrets manager before injecting it into a Kubernetes secret.

Valid examples:

- `<username | base64encode>`

- `<path:secrets/data/my-db#username | base64encode>`

- `<path:secrets/data/my-db#username#version3 | base64encode>`

This can be used for both generic and inline-path placeholders.

##### `base64decode`

The base64decode modifier decodes base64 encoded values into plain-text.

Valid examples:

- `<b64_username | base64decode>`

- `<path:secrets/data/my-db#b64_username | base64decode>`

- `<path:secrets/data/my-db#b64_username#version3 | base64decode>`

##### `jsonPath`

The jsonPath modifier allows you use jsonpath to post-process objects or json, retrieved from a secrets manager, before injecting into a Kubernetes manifest.  The output is a string.  If your desired datatype is not a string, pass the output through jsonParse.

See the Kubernetes jsonPath documentation for more detail: [https://kubernetes.io/docs/reference/kubectl/jsonpath/](https://kubernetes.io/docs/reference/kubectl/jsonpath/)

Valid examples:

- `<credentials | jsonPath {.username}>`

- `<path:secrets/data/my-db#credentials | jsonPath {.username}{':'}{.password}>`

- `<path:secrets/data/my-db#credentials#version3 | jsonPath {.username} | base64encode>`

- `<path:secrets/data/my-db#config | jsonPath {.replicas} | jsonParse>`

##### `jsonParse`

The jsonParse modifier parses json strings into objects.

Valid examples:

- `<credentialsJson | jsonParse>`

- `<path:secrets/data/my-db#credentialsJson | jsonParse>`

- `<path:secrets/data/my-db#credentialsJson#version3 | jsonParse>`

##### `yamlParse`

The yamlParse modifier converts YAML data into JSON.

Valid examples:

- `<credentials_yaml | yamlParse | jsonPath {.username}>`

- `<path:secrets/data/db_yaml#yaml | yamlParse | jsonPath {.username}{':'}{.password}>`

- `<path:secrets/data/db_yaml#yaml#version2 | yamlParse | jsonPath {.username} | base64encode>`

##### `indent`

The indent modifier indents the secret data by the specified number of space characters (`0x20`), largely useful when injecting secrets into YAML strings embedded in YAML.

Valid examples:

- `<path:secrets/data/db#certs | jsonPath {.certificate} | indent 3>`

##### `sha256sum`

The sha256sum modifier computes the SHA256 checksum of the string. Can be used to detect changes in a secret.

Valid examples:

```yaml
kind: Deployment
spec:
  template:
    metadata:
      annotations:
        checksum/secret: <path:secrets/data/db#certs | sha256sum>
```

### Error Handling

#### Detecting errors in chained commands

By default argocd-vault-plugin will read valid kubernetes YAMLs and replace variables with values from Vault.
If a previous command failed and outputs nothing to stdout and AVP reads the input from stdin with
the `-` argument, AVP will forward an empty YAML output downstream. To catch and prevent accientental errors
in chained commands, please use the `-o pipefail` bash option like so:

```bash
$ sh -c '((>&2 echo "some error" && exit 1) | argocd-vault-plugin generate - | kubectl diff -f -); echo $?;'
some error
0

$ set -o pipefail
$ sh -c '((>&2 echo "some error" && exit 1) | argocd-vault-plugin generate - | kubectl diff -f -); echo $?;'
some error
1
```
