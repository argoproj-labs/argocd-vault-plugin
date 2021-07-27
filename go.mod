module github.com/IBM/argocd-vault-plugin

go 1.16

require (
	github.com/aws/aws-sdk-go v1.40.9
	github.com/hashicorp/vault v1.7.3
	github.com/hashicorp/vault/api v1.1.1
	github.com/hashicorp/vault/sdk v0.2.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.8.1
	k8s.io/apimachinery v0.21.3
	k8s.io/client-go v0.21.3
	sigs.k8s.io/yaml v1.2.0
)
