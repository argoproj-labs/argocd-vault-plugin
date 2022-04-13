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

func TestJsonPath_unmarshal_error(t *testing.T) {
	var data interface{} = "hi, how are you"
	_, err := jsonPath([]string{"{.data.subkey}"}, data)
	if _, ok := err.(*json.SyntaxError); !ok {
		t.Fatalf("expected error [%s], got [%T]", "json.SyntaxError", err)
	}
}

func TestJsonPath_unmarshal_succcess(t *testing.T) {
	var data interface{} = "{\"data\": {\"subkey\": \"secret\"}}"
	var expected interface{} = "secret"
	res, err := jsonPath([]string{"{.data.subkey}"}, data)
	assertErrorEqual(t, nil, err)
	assertResultEqual(t, expected, res)
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

func TestBase64Decode_invalidParams(t *testing.T) {
	var data interface{} = "bXlzZWNyZXQ="
	expectedErr := fmt.Errorf("invalid parameters")
	_, err := base64decode([]string{"astring"}, data)
	assertErrorEqual(t, expectedErr, err)
}

func TestBase64Decode_invalidDataType(t *testing.T) {
	var data interface{} = map[string]interface{}{
		"data": map[string]interface{}{
			"subkey": "secret",
		},
	}
	expectedErr := fmt.Errorf("invalid datatype map[string]interface {}")
	_, err := base64decode([]string{}, data)
	assertErrorEqual(t, expectedErr, err)
}

func TestBase64Decode_success(t *testing.T) {
	var data interface{} = "bXlzZWNyZXQ="
	var expected interface{} = "mysecret"
	res, err := base64decode([]string{}, data)
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

func TestYAMLToJSON_success(t *testing.T) {
	var data interface{} = "data: secret"
	var expected interface{} = map[string]interface{}{
		"data": "secret",
	}
	res, err := yamlParse([]string{}, data)
	assertErrorEqual(t, nil, err)
	assertResultEqual(t, expected, res)
}

func TestYamlJsonPath_unmarshal_succcess(t *testing.T) {
	var data interface{} = "---\nkey1: secret1\nkey2: secret2\nkey3: secret3"
	var expected interface{} = "secret2"
	jsonData, err1 := yamlParse([]string{}, data)
	res, err2 := jsonPath([]string{"{.key2}"}, jsonData)
	assertErrorEqual(t, nil, err1)
	assertErrorEqual(t, nil, err2)
	assertResultEqual(t, expected, res)
}

func TestIndent_multiline(t *testing.T) {
	var data interface{} = "a\nmulti-line\nstring"
	var expected interface{} = "a\n   multi-line\n   string"
	res, err := indent([]string{"3"}, data)
	assertErrorEqual(t, nil, err)
	assertResultEqual(t, expected, res)
}

func TestIndent_singleline(t *testing.T) {
	var data interface{} = "a-string"
	var expected interface{} = "a-string"
	res, err := indent([]string{"10"}, data)
	assertErrorEqual(t, nil, err)
	assertResultEqual(t, expected, res)
}

func TestIndent_invalid(t *testing.T) {
	var data interface{} = "a-string"
	var expected interface{} = nil
	res, err := indent([]string{"ten"}, data)
	assertErrorEqual(t, fmt.Errorf("invalid indentation level"), err)
	assertResultEqual(t, expected, res)

	res, err = indent([]string{"10", "10"}, data)
	assertErrorEqual(t, fmt.Errorf("invalid parameters"), err)
	assertResultEqual(t, expected, res)
}

func TestSha256Sum_invalidParams(t *testing.T) {
	var data interface{} = "mysecret"
	expectedErr := fmt.Errorf("invalid parameters")
	_, err := sha256sum([]string{"astring"}, data)
	assertErrorEqual(t, expectedErr, err)
}

func TestSha256Sum_invalidDataType(t *testing.T) {
	var data interface{} = map[string]interface{}{
		"data": map[string]interface{}{
			"subkey": "secret",
		},
	}
	expectedErr := fmt.Errorf("invalid datatype map[string]interface {}, expected string")
	_, err := sha256sum([]string{}, data)
	assertErrorEqual(t, expectedErr, err)
}

func TestSha256Sum_success(t *testing.T) {
	var data interface{} = "mysecret"
	var expected interface{} = "652c7dc687d98c9889304ed2e408c74b611e86a40caa51c4b43f1dd5913c5cd0"
	res, err := sha256sum([]string{}, data)
	assertErrorEqual(t, nil, err)
	assertResultEqual(t, expected, res)
}
