package kube

import (
	"encoding/json"
	"fmt"
	"os"
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
		t.Fatalf("expected result [%v], got [%v]", expected, actual)
	}
}

func TestJsonPath_invalidParams(t *testing.T) {
	var data interface{} = map[string]interface{}{
		"data": map[string]interface{}{
			"subkey": "secret",
		},
	}
	expectedErr := fmt.Errorf("invalid parameters")
	_, err := jsonPath([]string{}, data)
	assertErrorEqual(t, expectedErr, err)
}

func TestJsonPath_missingPath(t *testing.T) {
	var data interface{} = map[string]interface{}{
		"data": map[string]interface{}{
			"subkey": "secret",
		},
	}
	expectedErr := fmt.Errorf("missingPath is not found")
	v, err := jsonPath([]string{"{.missingPath}"}, data)
	fmt.Fprintf(os.Stderr, "data: %v\n", data)
	fmt.Fprintf(os.Stderr, "v: %v\n", v)
	assertErrorEqual(t, expectedErr, err)
}

func TestJsonPath_invalidPath(t *testing.T) {
	var data interface{} = map[string]interface{}{
		"data": map[string]interface{}{
			"subkey": "secret",
		},
	}
	expectedErr := fmt.Errorf("unrecognized character in action: U+0021 '!'")
	_, err := jsonPath([]string{"{!invalidPath}"}, data)
	assertErrorEqual(t, expectedErr, err)
}

func TestJsonPath_succcess(t *testing.T) {
	var data interface{} = map[string]interface{}{
		"data": map[string]interface{}{
			"subkey": "secret",
		},
	}
	var expected interface{} = "secret"
	res, err := jsonPath([]string{"{.data.subkey}"}, data)
	assertErrorEqual(t, nil, err)
	assertResultEqual(t, expected, res)
}

func TestBase64Encode_invalidParams(t *testing.T) {
	var data interface{} = "mysecret"
	expectedErr := fmt.Errorf("invalid parameters")
	_, err := base64encode([]string{"astring"}, data)
	assertErrorEqual(t, expectedErr, err)
}

func TestBase64Encode_invalidDataType(t *testing.T) {
	var data interface{} = map[string]interface{}{
		"data": map[string]interface{}{
			"subkey": "secret",
		},
	}
	expectedErr := fmt.Errorf("invalid datatype map[string]interface {}")
	_, err := base64encode([]string{}, data)
	assertErrorEqual(t, expectedErr, err)
}

func TestBase64Encode_success(t *testing.T) {
	var data interface{} = "mysecret"
	var expected interface{} = "bXlzZWNyZXQ="
	res, err := base64encode([]string{}, data)
	assertErrorEqual(t, nil, err)
	assertResultEqual(t, expected, res)
}

func TestJsonParse_invalidParams(t *testing.T) {
	var data interface{} = "mysecret"
	expectedErr := fmt.Errorf("invalid parameters")
	_, err := jsonParse([]string{"astring"}, data)
	assertErrorEqual(t, expectedErr, err)
}

func TestJsonParse_invalidDataType(t *testing.T) {
	var data interface{} = map[string]interface{}{
		"data": map[string]interface{}{
			"subkey": "secret",
		},
	}
	expectedErr := fmt.Errorf("invalid datatype map[string]interface {}")
	_, err := jsonParse([]string{}, data)
	assertErrorEqual(t, expectedErr, err)
}

func TestJsonParse_invalidJSON(t *testing.T) {
	var data interface{} = "hello"
	_, err := jsonParse([]string{}, data)
	if _, ok := err.(*json.SyntaxError); !ok {
		t.Fatalf("expected error [%s], got [%T]", "json.SyntaxError", err)
	}
}

func TestJsonParse_success(t *testing.T) {
	var data interface{} = "{\"data\": {\"subkey\": \"secret\"}}"
	var expected interface{} = map[string]interface{}{
		"data": map[string]interface{}{
			"subkey": "secret",
		},
	}
	res, err := jsonParse([]string{}, data)
	assertErrorEqual(t, nil, err)
	assertResultEqual(t, expected, res)
}
