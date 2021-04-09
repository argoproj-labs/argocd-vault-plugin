package types

const (
	// Environment Variable Constants
	EnvAvpType         = "AVP_TYPE"
	EnvAvpRoleID       = "AVP_ROLE_ID"
	EnvAvpSecretID     = "AVP_SECRET_ID"
	EnvAvpAuthType     = "AVP_AUTH_TYPE"
	EnvAvpGithubToken  = "AVP_GITHUB_TOKEN"
	EnvAvpK8sRole      = "AVP_K8S_ROLE"
	EnvAvpK8sMountPath = "AVP_K8S_MOUNT_PATH"
	EnvAvpK8sTokenPath = "AVP_K8S_TOKEN_PATH"
	EnvAvpIbmAPIKey    = "AVP_IBM_API_KEY"
	EnvAvpKvVersion    = "AVP_KV_VERSION"
	EnvAvpPathPrefix   = "AVP_PATH_PREFIX"

	// Backend and Auth Constants
	VaultBackend            = "vault"
	IbmSecretManagerbackend = "ibmsecretmanager"
	K8sAuth                 = "k8s"
	ApproleAuth             = "approle"
	GithubAuth              = "github"
	IamAuth                 = "iam"
)
