<a name="v0.6.0"></a>
## [v0.6.0](https://github.com/IBM/argocd-vault-plugin/compare/v0.5.1...v0.6.0) (2021-03-15)

### Fix

* Increase timeout on DefaultHttpClient to 60s ([#99](https://github.com/IBM/argocd-vault-plugin/issues/99))

### BREAKING CHANGE


Secrets will now have to be base64 encoded in Vault to use the data field of a Kubernetes secret

<a name="v0.5.1"></a>
## [v0.5.1](https://github.com/IBM/argocd-vault-plugin/compare/v0.5.0...v0.5.1) (2021-03-10)

### Fix

* skip empty manifests ([#89](https://github.com/IBM/argocd-vault-plugin/issues/89))

### Test

* Add e2e tests for generate command and test vault ([#90](https://github.com/IBM/argocd-vault-plugin/issues/90))

<a name="v0.5.0"></a>
## [v0.5.0](https://github.com/IBM/argocd-vault-plugin/compare/v0.4.1...v0.5.0) (2021-03-02)

### Docs

* Add kustomize app to easily spin up argocd with plugin ([#71](https://github.com/IBM/argocd-vault-plugin/issues/71))

### Feat

* Make k8s mount path optional ([#82](https://github.com/IBM/argocd-vault-plugin/issues/82))
* Add support for cacert, capath and skip verify variables ([#78](https://github.com/IBM/argocd-vault-plugin/issues/78))
* Add kustomize to argocd-cm kustomize patch ([#77](https://github.com/IBM/argocd-vault-plugin/issues/77))
* Support Vault Namespace ([#76](https://github.com/IBM/argocd-vault-plugin/issues/76))
* Support Kubernetes auth ([#65](https://github.com/IBM/argocd-vault-plugin/issues/65))

### Fix

* k8s auth type reversed args ([#81](https://github.com/IBM/argocd-vault-plugin/issues/81))
* Add check for empty map in secret manager logic ([#75](https://github.com/IBM/argocd-vault-plugin/issues/75))


<a name="v0.4.1"></a>
## [v0.4.1](https://github.com/IBM/argocd-vault-plugin/compare/v0.4.0...v0.4.1) (2021-02-24)

### Fix

* Change secretData to stringData ([#70](https://github.com/IBM/argocd-vault-plugin/issues/70))


<a name="v0.4.0"></a>
## [v0.4.0](https://github.com/IBM/argocd-vault-plugin/compare/v0.3.0...v0.4.0) (2021-02-18)

### Docs

* Update usage to have specific examples for each backend/auth ([#61](https://github.com/IBM/argocd-vault-plugin/issues/61))

### Fix

* Handle YAMLs with non-placeholder strings properly ([#62](https://github.com/IBM/argocd-vault-plugin/issues/62))
* Fix PR template typo
* Update usage to be argocd-vault-plugin ([#56](https://github.com/IBM/argocd-vault-plugin/issues/56))

### Refactor

* Update all TODO comments
* Move auth to own path, split out config and types
* Rename vault to backends, start using AuthType interface

### Tests

* Add more tests for utils


<a name="v0.3.0"></a>
## [v0.3.0](https://github.com/IBM/argocd-vault-plugin/compare/v0.2.2...v0.3.0) (2021-02-05)
### BREAKING CHANGE: KV v2 is now the default secret engine for the Vault backend

### Docs

* Minor readme tweaks ([#46](https://github.com/IBM/argocd-vault-plugin/issues/46))
* Update readme to include chmod of binary ([#48](https://github.com/IBM/argocd-vault-plugin/issues/48))
* Update contributing, add code of conduct ([#43](https://github.com/IBM/argocd-vault-plugin/issues/43))

### Feat

* Add support for kvv2 ([#49](https://github.com/IBM/argocd-vault-plugin/issues/49))

### Fix

* Include .yml extension and helpful kvv2 message ([#51](https://github.com/IBM/argocd-vault-plugin/issues/51))


<a name="v0.2.2"></a>
## [v0.2.2](https://github.com/IBM/argocd-vault-plugin/compare/v0.2.1...v0.2.2) (2021-01-25)
Resolves a bug that was introduced in 0.2.0 that broke the ability to print valid yaml if there was an issue writing file/directory based on permissions

### Fix

* Change some logic to avoid yaml issues ([#41](https://github.com/IBM/argocd-vault-plugin/issues/41))


<a name="v0.2.1"></a>
## [v0.2.1](https://github.com/IBM/argocd-vault-plugin/compare/v0.2.0...v0.2.1) (2021-01-22)

### Fix

* Set token before writing to file, silent fail if error ([#39](https://github.com/IBM/argocd-vault-plugin/issues/39))


<a name="v0.2.0"></a>
## [v0.2.0](https://github.com/IBM/argocd-vault-plugin/compare/v0.1.0...v0.2.0) (2021-01-13)

### Docs

* Update readme to include better usage instructions ([#28](https://github.com/IBM/argocd-vault-plugin/issues/28))
* Add Go report card to README

### Feat

* Read/Write vault token to/from file ([#35](https://github.com/IBM/argocd-vault-plugin/issues/35))
* Read configuration from files, Secrets, and env variables ([#33](https://github.com/IBM/argocd-vault-plugin/issues/33))
* Support IBM Secret Manager ([#29](https://github.com/IBM/argocd-vault-plugin/issues/29))
* Allow for all kubernetes kinds, accept path as annotation ([#26](https://github.com/IBM/argocd-vault-plugin/issues/26))


<a name="v0.1.0"></a>
## v0.1.0 (2020-11-19)

### Chore

* Skip Windows in tests
* go mod tidy
* Tidy modules
* Save point
* Save before refactor
* Improve error handling
* Add issue templates ([#12](https://github.com/IBM/argocd-vault-plugin/issues/12))
* Add pull request template ([#13](https://github.com/IBM/argocd-vault-plugin/issues/13))
* Add codeowners file
* Enable code scanning
* Initialize project ([#2](https://github.com/IBM/argocd-vault-plugin/issues/2))

### Docs

* Update logo to have less space
* Add logo to readme
* Add initial readme documentation ([#10](https://github.com/IBM/argocd-vault-plugin/issues/10))

### Feat

* Support Service and multi-YAML documents
* Use config for prefix, some logic changes, tests  ([#21](https://github.com/IBM/argocd-vault-plugin/issues/21))
* Support ConfigMap templates ([#18](https://github.com/IBM/argocd-vault-plugin/issues/18))
* Use new vault config
* Connect auth to manifest generation
* Add some initial auth and vault code
* Support Deployments and Secrets
* Proper find/replace for Deployments
* quick and dirty poc, must be refactored
* Add codecov status badge to readme
* Send code coverage to codecov ([#7](https://github.com/IBM/argocd-vault-plugin/issues/7))
* Add Github workflow status badges
* Add generate command, a sample test and some docs outline ([#6](https://github.com/IBM/argocd-vault-plugin/issues/6))

### Fix

* Add line endings for Windows?
* Support bools from Vault
* Typo and remove unused function
* Errors go to stderr
* Checkin vault util file
* Build binary in cwd
* Error handling file I/O

### Refactor

* Rebase and get secrets for Services
* Use YAML decoder since guaranteed YAML input
* Rename fixtures
* No panic, I/O to util

### Tests

* For Service and error-path Deployment
* ToYAML tests
* Tests for failure, secrets
* Testing generic replacement
* Simple tests for CLI
