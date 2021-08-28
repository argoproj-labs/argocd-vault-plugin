module github.com/IBM/argocd-vault-plugin

go 1.16

require (
	cloud.google.com/go v0.81.0
	github.com/IBM/go-sdk-core/v5 v5.5.0
	github.com/IBM/secrets-manager-go-sdk v1.0.24
	github.com/aws/aws-sdk-go v1.40.9
	github.com/googleapis/gax-go/v2 v2.0.5
	github.com/hashicorp/go-hclog v0.16.1
	github.com/hashicorp/vault v1.7.3
	github.com/hashicorp/vault-plugin-secrets-kv v0.8.0
	github.com/hashicorp/vault/api v1.1.1
	github.com/hashicorp/vault/sdk v0.2.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.8.1
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e
	google.golang.org/genproto v0.0.0-20210602131652-f16073e35f0c
	k8s.io/apimachinery v0.21.3
	k8s.io/client-go v0.21.3
	sigs.k8s.io/yaml v1.2.0
)
