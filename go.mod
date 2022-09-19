module github.com/IBM/argocd-vault-plugin

go 1.16

require (
	cloud.google.com/go v0.81.0
	github.com/Azure/azure-sdk-for-go v57.0.0+incompatible
	github.com/IBM/go-sdk-core/v5 v5.5.0
	github.com/IBM/secrets-manager-go-sdk v1.0.24
	github.com/aws/aws-sdk-go v1.40.9
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/googleapis/gax-go/v2 v2.0.5
	github.com/hashicorp/go-hclog v0.16.1
	github.com/hashicorp/vault v1.7.4
	github.com/hashicorp/vault-plugin-secrets-kv v0.8.0
	github.com/hashicorp/vault/api v1.1.1
	github.com/hashicorp/vault/sdk v0.2.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.8.1
	github.com/thycotic/tss-sdk-go v1.0.0
	golang.org/x/net v0.0.0-20210805182204-aaa1db679c0d
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/tools v0.1.6-0.20210908190839-cf92b39a962c // indirect
	google.golang.org/genproto v0.0.0-20210602131652-f16073e35f0c
	k8s.io/apimachinery v0.21.3
	k8s.io/client-go v0.21.3
	sigs.k8s.io/yaml v1.2.0
)
