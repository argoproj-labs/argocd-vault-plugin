package cmd

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	yaml2 "github.com/ghodss/yaml"
	"github.com/mitchellh/mapstructure"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Deployment struct {
	data      map[string]interface{}
	vaultdata map[string]string
}

func NewDeployment(data map[string]interface{}) *Deployment {
	d := Deployment{
		data: data,
		vaultdata: map[string]string{
			"namespace": "default",
			"name":      "inspector",
			"tag":       "123",
			"something": "else",
		},
	}

	return &d
}

func (d *Deployment) Replace() interface{} {
	fmt.Println("Originally had: ")
	fmt.Println(d.data)
	d.replaceinner(&d.data)
	fmt.Println("Struct has the following: ")
	fmt.Println(d.data)
	return nil
}

func (d *Deployment) replaceinner(block *map[string]interface{}) interface{} {
	// obj, _ := (*block).(map[interface{}]interface{})
	obj := (*block)
	for key, value := range obj {
		valuetype := reflect.ValueOf(value).Kind()

		if valuetype == reflect.Map {
			// fmt.Printf("key: %s, need to recurse since value is %s\n", key, value)
			inner := value.(map[string]interface{})
			d.replaceinner(&inner)
		} else if valuetype == reflect.Slice {
			fmt.Printf("key: %s, need to iterate and recurse since value is %s\n", key, value)
			for _, elm := range value.([]interface{}) {
				inner := elm.(map[string]interface{})
				d.replaceinner(&inner)
			}
		} else if valuetype == reflect.String {
			replacedvalue, _ := d.replacestring(value.(string))
			obj[key] = replacedvalue
		} else {
			// fmt.Printf("key: %s, need to skip since type is %s\n", key, valuetype)
		}
	}

	// fmt.Println(block)
	return block
}

func (d *Deployment) replacestring(value string) (string, error) {
	re, _ := regexp.Compile(`(?mU)<(.*)>`)

	res := re.ReplaceAllFunc([]byte(value), func(match []byte) []byte {
		placeholder := strings.Trim(string(match), "<>")
		fmt.Printf("Matching byte slice is: %s from %s\n", placeholder, value)
		secretvalue, ok := d.vaultdata[string(placeholder)]
		if ok {
			return []byte(secretvalue)
		} else {
			fmt.Printf("Error: Missing Vault value for placeholder %s in value %s\n", placeholder, value)
		}
		return []byte("")
	})

	// TODO: check for unreplaced `<>`'s to throw an error

	return string(res), nil
}

func (d *Deployment) toKubeResource() *appsv1.Deployment {

	metadata := d.data["metadata"].(map[string]interface{})
	spec := d.data["spec"].(map[string]interface{})
	var objectmeta metav1.ObjectMeta
	var deploymentspec appsv1.DeploymentSpec

	var specdecodemetata mapstructure.Metadata
	specdecodeconfig := mapstructure.DecoderConfig{
		ErrorUnused: true,
		TagName:     "json",
		Result:      &deploymentspec,
		Metadata:    &specdecodemetata,
	}
	specdecoder, _ := mapstructure.NewDecoder(&specdecodeconfig)

	mapstructure.Decode(metadata, &objectmeta)
	err := specdecoder.Decode(spec)
	if err != nil {
		fmt.Printf("Err: %s\n", err)
	}

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: d.data["apiVersion"].(string),
		},
		ObjectMeta: objectmeta,
		Spec:       appsv1.DeploymentSpec{},
	}
	return deployment
}

// func (d *Deployment) ToYAML() string {
// 	return ""
// }

// func (d *Deployment) ToYAML() string {
// 	kubedeployment := d.toKubeResource()
// 	fmt.Println(kubedeployment)
// 	jsonSecret, err := json.Marshal(&kubedeployment)
// 	if err != nil {
// 		panic(err)
// 	}

// 	yamlSecret, _ := yaml2.JSONToYAML(jsonSecret)
// 	fmt.Printf(string(yamlSecret))
// 	return string(yamlSecret)
// }

func (d *Deployment) ToYAML() string {
	kubedeployment := d.toKubeResource()
	jsondata, err := json.Marshal(&kubedeployment)
	yaml, err := yaml2.JSONToYAML(jsondata)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	fmt.Printf("YAML output is: \n%s", string(yaml))
	return string(yaml)
}
