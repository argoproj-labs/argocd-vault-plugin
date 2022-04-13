package kube

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	k8jsonpath "k8s.io/client-go/util/jsonpath"
	k8yaml "sigs.k8s.io/yaml"
)

var modifiers = map[string]func([]string, interface{}) (interface{}, error){
	"base64encode": base64encode,
	"base64decode": base64decode,
	"jsonPath":     jsonPath,
	"jsonParse":    jsonParse,
	"yamlParse":    yamlParse,
	"indent":       indent,
	"sha256sum":    sha256sum,
}

func indent(params []string, input interface{}) (interface{}, error) {
	if len(params) != 1 {
		return nil, fmt.Errorf("invalid parameters")
	}

	level, err := strconv.ParseInt(params[0], 10, 0)
	if err != nil {
		return nil, fmt.Errorf("invalid indentation level")
	}

	switch input.(type) {
	case string:
		{
			lines := strings.Split(input.(string), "\n")
			var builder strings.Builder
			builder.WriteString(strings.TrimSpace(lines[0]))
			if len(lines) > 1 {
				builder.WriteString("\n")
			}

			for i := 1; i < len(lines); i += 1 {
				if len(lines[i]) > 0 {
					for j := 0; int64(j) < level; j += 1 {
						builder.WriteString(" ")
					}
					builder.WriteString(strings.TrimSpace(lines[i]))
					if i < len(lines)-1 {
						builder.WriteString("\n")
					}
				}
			}

			return builder.String(), nil
		}
	default:
		return nil, fmt.Errorf("invalid datatype %v", reflect.TypeOf(input))
	}
}

func base64encode(params []string, input interface{}) (interface{}, error) {
	if len(params) > 0 {
		return nil, fmt.Errorf("invalid parameters")
	}
	switch input.(type) {
	case string:
		{
			s := base64.StdEncoding.EncodeToString([]byte(input.(string)))
			return s, nil
		}
	default:
		return nil, fmt.Errorf("invalid datatype %v", reflect.TypeOf(input))
	}
}

func base64decode(params []string, input interface{}) (interface{}, error) {
	if len(params) > 0 {
		return nil, fmt.Errorf("invalid parameters")
	}
	switch input.(type) {
	case string:
		{
			s, _ := base64.StdEncoding.DecodeString(input.(string))
			return string(s), nil
		}
	default:
		return nil, fmt.Errorf("invalid datatype %v", reflect.TypeOf(input))
	}
}

func jsonPath(params []string, input interface{}) (interface{}, error) {
	if len(params) < 1 {
		return nil, fmt.Errorf("invalid parameters")
	}

	jp := k8jsonpath.New("AVPJsonPath")
	jp.AllowMissingKeys(false)
	err := jp.Parse(strings.Join(params, " "))
	if err != nil {
		return nil, err
	}

	// Auto-unmarshal strings
	obj := input
	if reflect.ValueOf(input).Kind() == reflect.String {
		err := json.Unmarshal([]byte(input.(string)), &obj)
		if err != nil {
			return nil, err
		}
	}

	var buf bytes.Buffer
	err = jp.Execute(&buf, obj)
	if err != nil {
		return nil, err
	}
	return buf.String(), nil
}

func jsonParse(params []string, input interface{}) (interface{}, error) {
	if len(params) > 0 {
		return nil, fmt.Errorf("invalid parameters")
	}
	switch input.(type) {
	case string:
		{
			var obj interface{}
			err := json.Unmarshal([]byte(input.(string)), &obj)
			if err != nil {
				return nil, err
			}
			return obj, nil
		}
	default:
		return nil, fmt.Errorf("invalid datatype %v", reflect.TypeOf(input))
	}
}

func yamlParse(params []string, input interface{}) (interface{}, error) {
	if len(params) > 0 {
		return nil, fmt.Errorf("invalid parameters")
	}
	switch input.(type) {
	case string:
		{
			var obj interface{}
			err := k8yaml.Unmarshal([]byte(input.(string)), &obj)
			if err != nil {
				return nil, err
			}
			return obj, nil
		}
	default:
		return nil, fmt.Errorf("invalid datatype %v", reflect.TypeOf(input))
	}
}

func sha256sum(params []string, input interface{}) (interface{}, error) {
	if len(params) > 0 {
		return nil, fmt.Errorf("invalid parameters")
	}
	if reflect.ValueOf(input).Kind() == reflect.String {
		sum := sha256.Sum256([]byte(input.(string)))
		return hex.EncodeToString(sum[:]), nil
	} else {
		return nil, fmt.Errorf("invalid datatype %v, expected string", reflect.TypeOf(input))
	}

}
