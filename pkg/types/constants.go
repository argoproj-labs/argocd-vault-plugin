package types

const (
	// Environment Variable Prefix
	EnvArgoCDPrefix = "ARGOCD_ENV"

	// Environment Variable Constants
	EnvAvpType             = "AVP_TYPE"
	EnvAvpRoleID           = "AVP_ROLE_ID"
	EnvAvpSecretID         = "AVP_SECRET_ID"
	EnvAvpAuthType         = "AVP_AUTH_TYPE"
	EnvAvpGithubToken      = "AVP_GITHUB_TOKEN"
	EnvAvpK8sRole          = "AVP_K8S_ROLE"
	EnvAvpK8sMountPath     = "AVP_K8S_MOUNT_PATH"
	EnvAvpMountPath        = "AVP_MOUNT_PATH"
	EnvAvpK8sTokenPath     = "AVP_K8S_TOKEN_PATH"
	EnvAvpIBMAPIKey        = "AVP_IBM_API_KEY"
	EnvAvpIBMInstanceURL   = "AVP_IBM_INSTANCE_URL"
	EnvAvpKvVersion        = "AVP_KV_VERSION"
	EnvAvpPathPrefix       = "AVP_PATH_PREFIX"
	EnvAWSRegion           = "AWS_REGION"
	EnvVaultAddress        = "VAULT_ADDR"
	EnvYCLKeyID            = "AVP_YCL_KEY_ID"
	EnvYCLServiceAccountID = "AVP_YCL_SERVICE_ACCOUNT_ID"
	EnvYCLPrivateKey       = "AVP_YCL_PRIVATE_KEY"

	// Backend and Auth Constants
	VaultBackend              = "vault"
	IBMSecretsManagerbackend  = "ibmsecretsmanager"
	AWSSecretsManagerbackend  = "awssecretsmanager"
	GCPSecretManagerbackend   = "gcpsecretmanager"
	AzureKeyVaultbackend      = "azurekeyvault"
	Sopsbackend               = "sops"
	YandexCloudLockboxbackend = "yandexcloudlockbox"
	OnePasswordConnect        = "1passwordconnect"
	K8sAuth                   = "k8s"
	ApproleAuth               = "approle"
	GithubAuth                = "github"
	TokenAuth                 = "token"
	IAMAuth                   = "iam"
	AwsDefaultRegion          = "us-east-2"
	GCPCurrentSecretVersion   = "latest"
	IBMMaxRetries             = 3
	IBMRetryIntervalSeconds   = 20
	IBMMaxPerPage             = 200
	IBMIAMCredentialsType     = "iam_credentials"
	IBMImportedCertType       = "imported_cert"
	IBMPublicCertType         = "public_cert"

	// Supported annotations
	AVPPathAnnotation          = "avp.kubernetes.io/path"
	AVPIgnoreAnnotation        = "avp.kubernetes.io/ignore"
	AVPRemoveMissingAnnotation = "avp.kubernetes.io/remove-missing"
	AVPSecretVersionAnnotation = "avp.kubernetes.io/secret-version"
	VaultKVVersionAnnotation   = "avp.kubernetes.io/kv-version"

	// Kube Constants
	ArgoCDNamespace = "argocd"
)
