package kube

//
// import (
// 	"fmt"
//
// 	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
// 	k8yaml "sigs.k8s.io/yaml"
// )
//
// // A Resource is the basis for all Templates
// type Resource struct {
// 	TemplateData      map[string]interface{} // The template as read from YAML
// 	replacementErrors []error                // Any errors encountered in performing replacements
// 	VaultData         map[string]interface{} // The data to replace with, from Vault
// }
//
// // Template is the template for Kubernetes
// type Template struct {
// 	Resource
// }
//
// // NewTemplate returns a *Template given the template's data, and a VaultType
// // func NewTemplate(template map[string]interface{}, vault vault.VaultType) (*Template, error) {
// // 	data, err := vault.GetSecrets("/config")
// // 	if err != nil {
// // 		return nil, err
// // 	}
// //
// // 	return &Template{
// // 		Resource{
// // 			TemplateData: template,
// // 			VaultData:    data,
// // 		},
// // 	}, nil
// // }
//
// // Replace will replace the <placeholders> in the template's data with values from Vault.
// // It will return an aggregrate of any errors encountered during the replacements
// func (t *Template) Replace() error {
// 	replaceInner(&t.Resource, &t.TemplateData, genericReplacement)
// 	if len(t.replacementErrors) != 0 {
// 		// TODO format multiple errors nicely
// 		return fmt.Errorf("Replace: could not replace all placeholders in Template: %s", t.replacementErrors)
// 	}
// 	return nil
// }
//
// func templateReplacement(key, value string, vaultData map[string]interface{}) (interface{}, []error) {
// 	res, err := genericReplacement(key, value, vaultData)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// configMap data values must be strings
// 	return stringify(res), err
// }
//
// // ToYAML seralizes the completed template into YAML to be consumed by Kubernetes
// func (t *Template) ToYAML() (string, error) {
// 	obj := &unstructured.Unstructured{}
// 	err := KubeResourceDecoder(&t.TemplateData).Decode(&obj)
// 	if err != nil {
// 		return "", fmt.Errorf("ToYAML: could not convert replaced template into %s: %s", obj.GetKind(), err)
// 	}
// 	res, err := k8yaml.Marshal(&obj)
// 	if err != nil {
// 		return "", fmt.Errorf("ToYAML: could not export %s into YAML: %s", obj.GetKind(), err)
// 	}
// 	return string(res), nil
// }
