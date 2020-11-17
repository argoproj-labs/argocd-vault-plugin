package kube

// Template is created from YAML files for Kubernetes resources (manifests) that contain
// <placeholders>'s. Templates can be replaced by replacing the <placeholders>
// with values from Vault. They can be serialized back to YAML for usage by Kubernetes.
type Template interface {
	Replace() error
	ToYAML() (string, error)
}

// A Resource is the basis for all Templates
type Resource struct {
	templateData      map[string]interface{} // The template as read from YAML
	replacementErrors []error                // Any errors encountered in performing replacements
	vaultData         map[string]interface{} // The data to replace with, from Vault
}
