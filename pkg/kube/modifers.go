package kube

import (
	"encoding/base64"
	"fmt"
	"reflect"

	k8jsonpath "k8s.io/client-go/util/jsonpath"
)

var modifiers = map[string]func([]string, interface{}) error{
	"base64encode": base64encode,
	"jsonPath":     jsonPath,
}

func base64encode(params []string, input interface{}) error {
	if len(params) > 0 {
		return fmt.Errorf("invalid parameters")
	}
	ref := reflect.ValueOf(input).Elem()
	switch ref.Interface().(type) {
	case string:
		{
			s := base64.StdEncoding.EncodeToString([]byte(ref.Interface().(string)))
			ref.Set(reflect.ValueOf(s))
			return nil
		}
	default:
		return fmt.Errorf("invalid datatype %v", reflect.TypeOf(input))
	}
}

func jsonPath(params []string, input interface{}) error {
	if len(params) != 1 {
		return fmt.Errorf("invalid parameters")
	}
	jp := k8jsonpath.New("AVPJsonPath")
	err := jp.Parse(fmt.Sprintf("{%s}", params[0]))
	if err != nil {
		return err
	}
	res, err := jp.FindResults(input)
	if err != nil {
		return err
	}
	if len(res) > 0 && len(res[0]) > 0 && !res[0][0].IsNil() {
		// set input to jsonPath output
		ref := reflect.ValueOf(input).Elem()
		ref.Set(reflect.ValueOf(res[0][0].Interface()))
		return nil
	}
	return fmt.Errorf("empty results")
}
