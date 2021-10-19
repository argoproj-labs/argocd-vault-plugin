package kube

import (
	"fmt"
	"reflect"
	"testing"
)

func assertErrorEqual(t *testing.T, expected error, actual error) {
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("expected error [%v], got [%v]", expected, actual)
	}
}

func assertResultEqual(t *testing.T, expected interface{}, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("expected error [%v], got [%v]", expected, actual)
	}
}

func TestJsonPath_invalidParams(t *testing.T) {
	var data interface{} = map[string]interface{}{
		"data": map[string]interface{}{
			"subkey": "secret",
		},
	}
	expectedErr := fmt.Errorf("invalid parameters")
	err := jsonPath([]string{}, &data)
	assertErrorEqual(t, expectedErr, err)
}

func TestJsonPath_missingPath(t *testing.T) {
	var data interface{} = map[string]interface{}{
		"data": map[string]interface{}{
			"subkey": "secret",
		},
	}
	expectedErr := fmt.Errorf("missing is not found")
	err := jsonPath([]string{".data.missing"}, &data)
	assertErrorEqual(t, expectedErr, err)
}

func TestJsonPath_invalidPath(t *testing.T) {
	var data interface{} = map[string]interface{}{
		"data": map[string]interface{}{
			"subkey": "secret",
		},
	}
	expectedErr := fmt.Errorf("unrecognized character in action: U+0021 '!'")
	err := jsonPath([]string{"!invalidPath"}, &data)
	assertErrorEqual(t, expectedErr, err)
}

func TestJsonPath_emptyPath(t *testing.T) {
	var data interface{} = map[string]interface{}{
		"data": map[string]interface{}{
			"subkey": nil,
		},
	}
	expectedErr := fmt.Errorf("empty results")
	err := jsonPath([]string{".data.subkey"}, &data)
	assertErrorEqual(t, expectedErr, err)
}

func TestJsonPath_succcess(t *testing.T) {
	var data interface{} = map[string]interface{}{
		"data": map[string]interface{}{
			"subkey": "secret",
		},
	}
	var expected interface{} = "secret"
	err := jsonPath([]string{".data.subkey"}, &data)
	assertErrorEqual(t, nil, err)
	assertResultEqual(t, expected, data)
}

func TestBase64Encode_invalidParams(t *testing.T) {
	var data interface{} = "mysecret"
	expectedErr := fmt.Errorf("invalid parameters")
	err := base64encode([]string{"astring"}, &data)
	assertErrorEqual(t, expectedErr, err)
}

func TestBase64Encode_invalidDataType(t *testing.T) {
	var data interface{} = map[string]interface{}{
		"data": map[string]interface{}{
			"subkey": "secret",
		},
	}
	expectedErr := fmt.Errorf("invalid datatype *interface {}")
	err := base64encode([]string{}, &data)
	assertErrorEqual(t, expectedErr, err)
}

func TestBase64Encode_success(t *testing.T) {
	var data interface{} = "mysecret"
	var expected interface{} = "bXlzZWNyZXQ="
	err := base64encode([]string{}, &data)
	assertErrorEqual(t, nil, err)
	assertResultEqual(t, expected, data)
}
