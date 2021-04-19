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

// Encodes a placeholder from an input string
func encodePlaceholder(input string) string {
	encoded := fmt.Sprintf("<%s>", input)
	return encoded
}

// replaceableInner recurses through the given map and returns true if _any_ value is a `<placeholder>` string
func replaceableInner(node *map[string]interface{}) bool {
	obj := *node

	for _, value := range obj {
		valueType := reflect.ValueOf(value).Kind()
		if valueType == reflect.Map {
			inner, ok := value.(map[string]interface{})
			if !ok {
				panic(fmt.Sprintf("Deserialized YAML node is non map[string]interface{}"))
			}
			if replaceableInner(&inner) {
				return true
			}
		} else if valueType == reflect.Slice {
			for _, elm := range value.([]interface{}) {
				switch elm.(type) {
				case map[string]interface{}:
					{
						inner := elm.(map[string]interface{})
						if replaceableInner(&inner) {
							return true
						}
					}
				case string:
					{
						placeHolderValue, _ := tryDecode(elm.(string))
						if placeholder.Match([]byte(placeHolderValue)) {
							return true
						}
					}
				default:
					{
						panic(fmt.Sprintf("Deserialized YAML list node is non map[string]interface{} nor string"))
					}
				}
			}
		} else if valueType == reflect.String {
			placeHolderValue, _ := tryDecode(value.(string))
			if placeholder.Match([]byte(placeHolderValue)) {
				return true
			}
		}
	}
	return false
}

// replaceInner recurses through the given map and replaces the placeholders by calling `replacerFunc`
// with the key, value, and map of keys to replacement values
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
						elmString, decoded := tryDecode(elm.(string))
						replacement, err := replacerFunc(key, elmString, r.VaultData)
						if len(err) != 0 {
							r.replacementErrors = append(r.replacementErrors, err...)
						}

						// if we decoded the input value, make an effort to put it back in the
						// format we found it.
						result := replacement
						if decoded {
							result = tryEncode(replacement)
						}

						value.([]interface{})[idx] = result
					}
				default:
					{
						panic(fmt.Sprintf("Deserialized YAML list node is non map[string]interface{} nor string"))
					}
				}
			}
		} else if valueType == reflect.String {
			// Base case, replace templated strings
			valueString, decoded := tryDecode(value.(string))
			replacement, err := replacerFunc(key, valueString, r.VaultData)
			if len(err) != 0 {
				r.replacementErrors = append(r.replacementErrors, err...)
			}

			// if we decoded the input value, make an effort to put it back in the
			// format we found it.
			result := replacement
			if decoded {
				result = tryEncode(replacement)
			}

			obj[key] = result
		}
	}
}

func genericReplacement(key, value string, vaultData map[string]interface{}) (_ interface{}, err []error) {
	var nonStringReplacement interface{}

	res := placeholder.ReplaceAllFunc([]byte(value), func(match []byte) []byte {
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

func tryDecode(s string) (string, bool) {
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err == nil {
		return string(decoded[:]), true
	}

	return s, false
}

func tryEncode(s interface{}) interface{} {
	switch s.(type) {
	case []byte:
		{
			byteString := s.([]byte)
			encoded := base64.StdEncoding.EncodeToString(byteString)
			return encoded
		}
	case string:
		{
			byteString := []byte(s.(string))
			encoded := base64.StdEncoding.EncodeToString(byteString)
			return encoded
		}
	default:
		{
			return s
		}
	}
}
