package kube

type Template interface {
	Replace() error
	ToYAML() (string, error)
}

type Resource struct {
	templateData      map[string]interface{}
	replacementErrors []error
	vaultData         map[string]interface{}
}
