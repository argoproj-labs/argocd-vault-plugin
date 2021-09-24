package types

const (
	// Environment Variable Constants
	EnvAvpType             = "AVP_TYPE"
	EnvAvpRoleID           = "AVP_ROLE_ID"
	EnvAvpSecretID         = "AVP_SECRET_ID"
	EnvAvpAuthType         = "AVP_AUTH_TYPE"
	EnvAvpGithubToken      = "AVP_GITHUB_TOKEN"
	EnvAvpK8sRole          = "AVP_K8S_ROLE"
	EnvAvpK8sMountPath     = "AVP_K8S_MOUNT_PATH"
	EnvAvpK8sTokenPath     = "AVP_K8S_TOKEN_PATH"
	EnvAvpIBMAPIKey        = "AVP_IBM_API_KEY"
	EnvAvpIBMInstanceURL   = "AVP_IBM_INSTANCE_URL"
	EnvAvpKvVersion        = "AVP_KV_VERSION"
	EnvAvpPathPrefix       = "AVP_PATH_PREFIX"
	EnvAWSRegion           = "AWS_REGION"
	EnvVaultAddress        = "VAULT_ADDR"
	EnvAvpThycoticURL      = "AVP_THYCOTIC_URL"
	EnvAvpThycoticUser     = "AVP_THYCOTIC_USER"
	EnvAvpThycoticPassword = "AVP_THYCOTIC_PASSWORD"

	// Backend and Auth Constants
	VaultBackend                = "vault"
	IBMSecretsManagerbackend    = "ibmsecretsmanager"
	AWSSecretsManagerbackend    = "awssecretsmanager"
	GCPSecretManagerbackend     = "gcpsecretmanager"
	AzureKeyVaultbackend        = "azurekeyvault"
	ThycoticSecretServerbackend = "thycoticsecretserver"
	K8sAuth                     = "k8s"
	ApproleAuth                 = "approle"
	GithubAuth                  = "github"
	TokenAuth                   = "token"
	IAMAuth                     = "iam"
	AwsDefaultRegion            = "us-east-2"
	GCPCurrentSecretVersion     = "latest"
	IBMMaxRetries               = 3
	IBMRetryIntervalSeconds     = 20
	IBMMaxPerPage               = 200

	// Supported annotations
	AVPPathAnnotation          = "avp.kubernetes.io/path"
	AVPIgnoreAnnotation        = "avp.kubernetes.io/ignore"
	AVPRemoveMissingAnnotation = "avp.kubernetes.io/remove-missing"
	AVPSecretVersionAnnotation = "avp.kubernetes.io/secret-version"
	VaultKVVersionAnnotation   = "avp.kubernetes.io/kv-version"
)
