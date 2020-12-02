package kube

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	k8yaml "k8s.io/apimachinery/pkg/util/yaml"
)

func replaceInner(
	r *Resource,
	node *map[string]interface{},
	replacerFunc func(string, string, map[string]interface{}) (interface{}, []error)) {
	obj := *node
	for key, value := range obj {
		valueType := reflect.ValueOf(value).Kind()

		// Recurse through nested maps
		if valueType == reflect.Map {
			inner, ok := value.(map[string]interface{})
			if !ok {
				panic(fmt.Sprintf("Deserialized YAML node is non map[string]interface{}"))
			}
			replaceInner(r, &inner, replacerFunc)
		} else if valueType == reflect.Slice {
			for idx, elm := range value.([]interface{}) {
				switch elm.(type) {
				case map[string]interface{}:
					{
						inner := elm.(map[string]interface{})
						replaceInner(r, &inner, replacerFunc)
					}
				case string:
					{
						// Base case, replace templated strings
						replacement, err := replacerFunc(key, elm.(string), r.VaultData)
						if len(err) != 0 {
							r.replacementErrors = append(r.replacementErrors, err...)
						}
						value.([]interface{})[idx] = replacement
					}
				default:
					{
						panic(fmt.Sprintf("Deserialized YAML list node is non map[string]interface{} nor string"))
					}
				}
			}
		} else if valueType == reflect.String {

			// Base case, replace templated strings
			replacement, err := replacerFunc(key, value.(string), r.VaultData)
			if len(err) != 0 {
				r.replacementErrors = append(r.replacementErrors, err...)
			}

			obj[key] = replacement
		}
	}
}

func genericReplacement(key, value string, vaultData map[string]interface{}) (_ interface{}, err []error) {
	re, _ := regexp.Compile(`(?mU)<(.*)>`)
	var nonStringReplacement interface{}

	res := re.ReplaceAllFunc([]byte(value), func(match []byte) []byte {
		placeholder := strings.Trim(string(match), "<>")
		secretValue, ok := vaultData[string(placeholder)]
		if ok {
			switch secretValue.(type) {
			case string:
				{
					return []byte(secretValue.(string))
				}
			default:
				{
					nonStringReplacement = secretValue
					return match
				}
			}
		} else {
			err = append(err, fmt.Errorf("replaceString: missing Vault value for placeholder %s in string %s: %s", placeholder, key, value))
		}
		return match
	})

	// The above block can only replace <placeholder> strings with other strings
	// In the case where the value is a non-string, we insert it directly here.
	// Useful for cases like `replicas: <replicas>`
	if nonStringReplacement != nil {
		return nonStringReplacement, err
	}
	return string(res), err
}

func stringify(input interface{}) string {
	switch input.(type) {
	case int:
		{
			return strconv.Itoa(input.(int))
		}
	case bool:
		{
			return strconv.FormatBool(input.(bool))
		}
	default:
		{
			return input.(string)
		}
	}
}

func kubeResourceDecoder(data *map[string]interface{}) *k8yaml.YAMLToJSONDecoder {
	jsondata, _ := json.Marshal(data)
	decoder := k8yaml.NewYAMLToJSONDecoder(bytes.NewReader(jsondata))
	return decoder
}
