package kube

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	k8yaml "k8s.io/apimachinery/pkg/util/yaml"
)

var placeholder, _ = regexp.Compile(`(?mU)<(.*)>`)
var indivPlacholder, _ = regexp.Compile(`(?mU)path:(.+?)\#(.+?)`)
var modifier, _ = regexp.Compile(`\|(.*)`)

// replaceInner recurses through the given map and replaces the placeholders by calling `replacerFunc`
// with the key, value, and map of keys to replacement values
func replaceInner(
	r *Resource,
	node *map[string]interface{},
	replacerFunc func(string, string, Resource) (interface{}, []error)) {
	obj := *node
	for key, value := range obj {
		valueType := reflect.ValueOf(value).Kind()

		// Recurse through nested maps
		if valueType == reflect.Map {
			inner, ok := value.(map[string]interface{})
			if !ok {
				continue
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
						replacement, err := replacerFunc(key, elm.(string), *r)
						if len(err) != 0 {
							r.replacementErrors = append(r.replacementErrors, err...)
						}
						value.([]interface{})[idx] = replacement
					}
				}
			}
		} else if valueType == reflect.String {
			// Base case, replace templated strings
			replacement, err := replacerFunc(key, value.(string), *r)
			if len(err) != 0 {
				r.replacementErrors = append(r.replacementErrors, err...)
			}

			obj[key] = replacement
		}
	}
}

func genericReplacement(key, value string, resource Resource) (_ interface{}, err []error) {
	var nonStringReplacement interface{}

	res := placeholder.ReplaceAllFunc([]byte(value), func(match []byte) []byte {
		placeholder := strings.Trim(string(match), "<>")

		var base64modifier bool
		if modifier.MatchString(placeholder) {
			modifierMatches := modifier.FindStringSubmatch(placeholder)
			base64modifier = strings.TrimSpace(string(modifierMatches[1])) == "base64encode"
			placeholder = strings.TrimSpace(strings.Split(placeholder, "|")[0])
		}

		var secretValue interface{}
		// check to see if should call out to get individual secret
		if indivPlacholder.Match([]byte(placeholder)) {
			indivSecretMatches := indivPlacholder.FindSubmatch([]byte(placeholder))
			secrets, secretErr := resource.Backend.GetSecrets(string(indivSecretMatches[1]), resource.Annotations)
			if secretErr != nil {
				err = append(err, secretErr)
				return match
			}
			secretKey := strings.TrimSpace(string(indivSecretMatches[2]))
			secretValue = secrets[secretKey]
		} else {
			secretValue = resource.Data[string(placeholder)]
		}

		// check for value in data
		if secretValue != nil {
			switch secretValue.(type) {
			case string:
				{
					if base64modifier {
						nonStringReplacement = []byte(stringify(secretValue))
						return match
					}
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

func configReplacement(key, value string, resource Resource) (interface{}, []error) {
	res, err := genericReplacement(key, value, resource)
	if err != nil {
		return nil, err
	}

	// configMap data values must be strings
	return stringify(res), err
}

func secretReplacement(key, value string, resource Resource) (interface{}, []error) {
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err == nil && placeholder.Match(decoded) {
		return genericReplacement(key, string(decoded), resource)
	}

	return genericReplacement(key, value, resource)
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
	case json.Number:
		{
			return string(input.(json.Number))
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
