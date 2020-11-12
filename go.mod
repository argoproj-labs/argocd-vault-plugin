module github.com/IBM/argocd-vault-plugin

go 1.14

require (
	github.com/beego/goyaml2 v0.0.0-20130207012346-5545475820dd
	github.com/ghodss/yaml v1.0.0
	github.com/hashicorp/vault/api v1.0.4
	github.com/imdario/mergo v0.3.11
	github.com/mitchellh/mapstructure v1.3.3
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/wendal/errors v0.0.0-20181209125328-7f31f4b264ec // indirect
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.19.3
	k8s.io/apimachinery v0.19.3
)
