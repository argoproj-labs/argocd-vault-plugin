module github.com/IBM/argocd-vault-plugin

go 1.14

require (
	github.com/hashicorp/vault/api v1.0.4
	github.com/spf13/cobra v1.0.0
	gopkg.in/yaml.v2 v2.3.0 // indirect
	k8s.io/api v0.19.3
	k8s.io/apimachinery v0.19.4
	sigs.k8s.io/yaml v1.2.0
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.3.0
)
