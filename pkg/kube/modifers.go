package kube

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	k8jsonpath "k8s.io/client-go/util/jsonpath"
)

var modifiers = map[string]func([]string, interface{}) (interface{}, error){
	"base64encode": base64encode,
	"jsonPath":     jsonPath,
	"jsonParse":    jsonParse,
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

	var buf bytes.Buffer
	err = jp.Execute(&buf, input)
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
