package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	k8yamldecoder "k8s.io/apimachinery/pkg/util/yaml"
	k8yaml "sigs.k8s.io/yaml"
)

type Deployment struct {
	templateData      map[string]interface{}
	replacementErrors []error
	vaultData         map[string]string
}

func NewDeployment(data map[string]interface{}) *Deployment {
	d := Deployment{
		templateData: data,
		vaultData: map[string]string{
			"namespace": "default",
			"name":      "inspector",
			"tag":       "123",
		},
	}

	return &d
}

func (d *Deployment) Replace() error {
	d.replaceInner(&d.templateData)
	if len(d.replacementErrors) != 0 {

		// TODO format multiple errors nicely
		return fmt.Errorf("could not replace placeholders in template: %s", d.replacementErrors)
	}
	return nil
}

func (d *Deployment) replaceInner(node *map[string]interface{}) {
	obj := *node
	for key, value := range obj {
		valueType := reflect.ValueOf(value).Kind()

		// Recurse through nested maps
		if valueType == reflect.Map {
			inner := value.(map[string]interface{})
			d.replaceInner(&inner)
		} else if valueType == reflect.Slice {

			// Iterate and recurse through maps in a slice
			for _, elm := range value.([]interface{}) {
				inner := elm.(map[string]interface{})
				d.replaceInner(&inner)
			}
		} else if valueType == reflect.String {

			// Base case, replace templated strings
			replacement, err := d.replaceString(key, value.(string))
			if len(err) != 0 {
				d.replacementErrors = append(d.replacementErrors, err...)
			}
			obj[key] = replacement
		}
	}
}

func (d *Deployment) replaceString(key string, value string) (string, []error) {
	re, _ := regexp.Compile(`(?mU)<(.*)>`)
	var err []error

	res := re.ReplaceAllFunc([]byte(value), func(match []byte) []byte {
		placeholder := strings.Trim(string(match), "<>")
		secretValue, ok := d.vaultData[string(placeholder)]
		if ok {
			return []byte(secretValue)
		} else {
			err = append(err, fmt.Errorf("missing Vault value for placeholder %s in string %s: %s", placeholder, key, value))
		}
		return match
	})

	return string(res), err
}

func (d *Deployment) toKubeResource() *appsv1.Deployment {
	jsondata, _ := json.Marshal(d.templateData)

	decoder := k8yamldecoder.NewYAMLOrJSONDecoder(bytes.NewReader(jsondata), 1000)
	var deployment appsv1.Deployment
	err := decoder.Decode(&deployment)
	if err != nil {

	}
	return &deployment
}

func (d *Deployment) ToYAML() string {
	res, err := k8yaml.Marshal(d.templateData)
	if err != nil {

	}
	return string(res)
}
