package types

const (
	// Environment Variable Constants
	EnvAvpType            = "AVP_TYPE"
	EnvAvpRoleID          = "AVP_ROLE_ID"
	EnvAvpSecretID        = "AVP_SECRET_ID"
	EnvAvpAuthType        = "AVP_AUTH_TYPE"
	EnvAvpGithubToken     = "AVP_GITHUB_TOKEN"
	EnvAvpK8sRole         = "AVP_K8S_ROLE"
	EnvAvpK8sMountPath    = "AVP_K8S_MOUNT_PATH"
	EnvAvpK8sTokenPath    = "AVP_K8S_TOKEN_PATH"
	EnvAvpIBMAPIKey       = "AVP_IBM_API_KEY"
	EnvAvpKvVersion       = "AVP_KV_VERSION"
	EnvAvpPathPrefix      = "AVP_PATH_PREFIX"
	EnvAWSAccessKey       = "AVP_AWS_ACCESS_KEY_ID"
	EnvAWSSecretAccessKey = "AVP_AWS_SECRET_ACCESS_KEY"
	EnvAWSRegion          = "AVP_AWS_REGION"

	// Backend and Auth Constants
	VaultBackend             = "vault"
	IBMSecretsManagerbackend = "ibmsecretsmanager"
	AWSSecretsManagerbackend = "awssecretsmanager"
	K8sAuth                  = "k8s"
	ApproleAuth              = "approle"
	GithubAuth               = "github"
	IAMAuth                  = "iam"
)
