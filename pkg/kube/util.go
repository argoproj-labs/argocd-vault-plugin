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
	"text/template"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/types"
	k8yaml "k8s.io/apimachinery/pkg/util/yaml"
)

type missingKeyError struct {
	s string
}

func (e *missingKeyError) Error() string {
	return e.s
}

var genericPlaceholder, _ = regexp.Compile(`(?mU)<(.*)>`)
var specificPathPlaceholder, _ = regexp.Compile(`(?mU)<path:([^#]+)#([^#]+)(?:#([^#]+))?>`)
var indivPlaceholderSyntax, _ = regexp.Compile(`(?mU)path:(?P<path>[^#]+?)#(?P<key>[^#]+?)(?:#(?P<version>.+?))??`)

// replaceInner recurses through the given map and replaces the placeholders by calling `replacerFunc`
// with the key, value, and map of keys to replacement values
func replaceInner(
	r *Resource,
	node *map[string]interface{},
	replacerFunc func(string, string, Resource) (interface{}, []error)) {
	removeMissing, _ := strconv.ParseBool(r.Annotations[types.AVPRemoveMissingAnnotation])
	if removeMissing && (r.Kind != "Secret" && r.Kind != "ConfigMap") {
		invalidRemoveMissingErr := fmt.Errorf("%s annotation can only be used on Secret or ConfigMap resources", types.AVPRemoveMissingAnnotation)
		r.replacementErrors = append(r.replacementErrors, invalidRemoveMissingErr)
		return
	}

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

			removeKey := false
			replacement, err := replacerFunc(key, value.(string), *r)
			if len(err) != 0 {
				if removeMissing {
					var filteredErr []error
					for _, e := range err {
						if _, ok := e.(*missingKeyError); ok {
							removeKey = true
						} else {
							filteredErr = append(filteredErr, e)
						}
					}

					err = filteredErr
				}

				r.replacementErrors = append(r.replacementErrors, err...)
			}

			if removeKey {
				delete(obj, key)
			} else {
				obj[key] = replacement
			}
		}
	}
}

func genericReplacement(key, value string, resource Resource) (_ interface{}, err []error) {
	var nonStringReplacement interface{}
	var placeholderRegex = specificPathPlaceholder

	// If the Vault path annotation is present, there may be placeholders with/without an explicit path
	// so we look for those. Only if the annotation is absent do we narrow the search to placeholders with
	// explicit paths, to prevent catching <things> that aren't placeholders
	// See https://github.com/argoproj-labs/argocd-vault-plugin/issues/130
	if _, pathAnnotationPresent := resource.Annotations[types.AVPPathAnnotation]; pathAnnotationPresent {
		placeholderRegex = genericPlaceholder
	}

	res := placeholderRegex.ReplaceAllFunc([]byte(value), func(match []byte) []byte {
		placeholder := strings.Trim(string(match), "<>")

		// Split modifiers from placeholder
		pipelineFields := strings.Split(placeholder, "|")
		placeholder = strings.Trim(pipelineFields[0], " ")

		var secretValue interface{}
		var secretErr error
		// Check to see if should call out to get individual secret (inline-path in placeholder)
		// This can include an optional version argument - if unspecified, the latest version is retrieved
		if indivPlaceholderSyntax.Match([]byte(placeholder)) {
			indivSecretMatches := indivPlaceholderSyntax.FindStringSubmatch(placeholder)
			path := indivSecretMatches[indivPlaceholderSyntax.SubexpIndex("path")]
			key := indivSecretMatches[indivPlaceholderSyntax.SubexpIndex("key")]
			version := indivSecretMatches[indivPlaceholderSyntax.SubexpIndex("version")]

			secretValue, secretErr = resource.Backend.GetIndividualSecret(path, strings.TrimSpace(key), version, resource.Annotations)
			if secretErr != nil {
				err = append(err, secretErr)
				return match
			}
		} else {
			secretValue = resource.Data[placeholder]
		}

		if secretValue != nil {
			// Process modifiers
			for _, stmt := range pipelineFields[1:] {
				fields := strings.Fields(stmt)
				functionName := strings.Trim(fields[0], " ")
				if _, ok := modifiers[functionName]; !ok {
					e := fmt.Errorf("invalid modifier: %s for placeholder %s in string %s: %s", functionName, placeholder, key, value)
					err = append(err, e)
					return match
				}
				var modErr error
				secretValue, modErr = modifiers[functionName](fields[1:], secretValue)
				if modErr != nil {
					e := fmt.Errorf("%s: %s for placeholder %s in string %s: %s", functionName, modErr.Error(), placeholder, key, value)
					err = append(err, e)
					return match
				}
			}

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
			missingKeyErr := &missingKeyError{
				s: fmt.Sprintf("replaceString: missing Vault value for placeholder %s in string %s: %s", placeholder, key, value),
			}
			err = append(err, missingKeyErr)
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
	if err == nil && genericPlaceholder.Match(decoded) {
		res, err := genericReplacement(key, string(decoded), resource)
		return base64.StdEncoding.EncodeToString([]byte(stringify(res))), err
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
	case []byte:
		{
			return string(input.([]byte))
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

func secretNamespaceName(input string) (string, string) {
	var secretNamespace, secretName string
	nameFields := strings.Split(input, ":")
	if len(nameFields) == 2 {
		secretNamespace = nameFields[0]
		secretName = nameFields[1]
	} else {
		secretNamespace = types.ArgoCDNamespace
		secretName = nameFields[0]
	}
	return secretNamespace, secretName
}

var dataTemplateFunctionContext template.FuncMap = func() template.FuncMap {
	funcMap := template.FuncMap{}
	for field, modifier := range modifiers {
		modifier := modifier
		funcMap[field] = func(input ...interface{}) (interface{}, error) {
			paramsLen := len(input) - 1
			if paramsLen < 0 {
				paramsLen = 0
			}
			params := make([]string, paramsLen)
			for i, param := range input[:paramsLen] {
				switch param.(type) {
				case string:
					{
						params[i] = param.(string)
					}
				default:
					{
						return nil, fmt.Errorf("invalid datatype %v", reflect.TypeOf(param))
					}
				}
			}
			return modifier(params, input[paramsLen])
		}
	}
	return funcMap
}()

func replaceFromTemplate(r *Resource, rawTemplate string) error {
	context := struct {
		Data map[string]interface{}
	}{
		Data: r.Data,
	}

	t, err := template.New(types.AVPDataTemplateAnnotation).Funcs(dataTemplateFunctionContext).Parse(rawTemplate)
	if err != nil {
		return err
	}

	var renderedTemplate bytes.Buffer
	if err := t.Execute(&renderedTemplate, context); err != nil {
		return err
	}

	renderedTemplateAsJson, err := k8yaml.ToJSON(renderedTemplate.Bytes())
	if err != nil {
		return err
	}

	var newData interface{}
	if err := json.Unmarshal(renderedTemplateAsJson, &newData); err != nil {
		return err
	}
	r.TemplateData["data"] = newData

	return nil
}
