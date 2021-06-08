package kube

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func assertSuccessfulReplacement(actual, expected *Resource, t *testing.T) {
	if len(actual.replacementErrors) != 0 {
		t.Fatalf("expected 0 errors but got: %s", actual.replacementErrors)
	}

	if !reflect.DeepEqual(actual.TemplateData, expected.TemplateData) {
		t.Fatalf("expected replaced template to look like %s\n but got: %s", expected.TemplateData, actual.TemplateData)
	}

	if !reflect.DeepEqual(actual.Data, expected.Data) {
		t.Fatalf("expected Vault map to look like %s\n but got: %s", expected.Data, actual.Data)
	}
}

func assertFailedReplacement(actual, expected *Resource, t *testing.T) {
	if !reflect.DeepEqual(actual.replacementErrors, expected.replacementErrors) {
		t.Fatalf("expected replacementErrors: %s but got %s", expected.replacementErrors, actual.replacementErrors)
	}

	if !reflect.DeepEqual(actual.TemplateData, expected.TemplateData) {
		t.Fatalf("expected replaced template to look like %s\n but got: %s", expected.TemplateData, actual.TemplateData)
	}

	if !reflect.DeepEqual(actual.Data, expected.Data) {
		t.Fatalf("expected Vault map to look like %s\n but got: %s", expected.Data, actual.Data)
	}
}

func TestGenericReplacement_simpleString(t *testing.T) {
	dummyResource := Resource{
		TemplateData: map[string]interface{}{
			"namespace": "<namespace>",
		},
		Data: map[string]interface{}{
			"namespace": "default",
		},
	}

	replaceInner(&dummyResource, &dummyResource.TemplateData, genericReplacement)

	expected := Resource{
		TemplateData: map[string]interface{}{
			"namespace": "default",
		},
		Data: map[string]interface{}{
			"namespace": "default",
		},
		replacementErrors: []error{},
	}

	assertSuccessfulReplacement(&dummyResource, &expected, t)
}

func TestGenericReplacement_multiString(t *testing.T) {
	dummyResource := Resource{
		TemplateData: map[string]interface{}{
			"namespace": "<namespace>",
			"image":     "foo.io/<name>:<tag>",
		},
		Data: map[string]interface{}{
			"namespace": "default",
			"name":      "app",
			"tag":       "latest",
		},
	}

	replaceInner(&dummyResource, &dummyResource.TemplateData, genericReplacement)

	expected := Resource{
		TemplateData: map[string]interface{}{
			"namespace": "default",
			"image":     "foo.io/app:latest",
		},
		Data: map[string]interface{}{
			"namespace": "default",
			"name":      "app",
			"tag":       "latest",
		},
		replacementErrors: []error{},
	}

	assertSuccessfulReplacement(&dummyResource, &expected, t)
}

func TestGenericReplacement_Base64(t *testing.T) {
	dummyResource := Resource{
		TemplateData: map[string]interface{}{
			"namespace": "<namespace | base64encode>",
			"image":     "foo.io/<name>:<tag>",
		},
		Data: map[string]interface{}{
			"namespace": "default",
			"name":      "app",
			"tag":       "latest",
		},
	}

	replaceInner(&dummyResource, &dummyResource.TemplateData, genericReplacement)

	expected := Resource{
		TemplateData: map[string]interface{}{
			"namespace": []uint8("default"),
			"image":     "foo.io/app:latest",
		},
		Data: map[string]interface{}{
			"namespace": "default",
			"name":      "app",
			"tag":       "latest",
		},
		replacementErrors: []error{},
	}

	assertSuccessfulReplacement(&dummyResource, &expected, t)
}

func TestGenericReplacement_nestedString(t *testing.T) {
	dummyResource := Resource{
		TemplateData: map[string]interface{}{
			"namespace": "<namespace>",
			"spec": map[string]interface{}{
				"selector": map[string]interface{}{
					"app": "<name>",
				},
			},
		},
		Data: map[string]interface{}{
			"namespace": "default",
			"name":      "foo",
		},
	}

	replaceInner(&dummyResource, &dummyResource.TemplateData, genericReplacement)

	expected := Resource{
		TemplateData: map[string]interface{}{
			"namespace": "default",
			"spec": map[string]interface{}{
				"selector": map[string]interface{}{
					"app": "foo",
				},
			},
		},
		Data: map[string]interface{}{
			"namespace": "default",
			"name":      "foo",
		},
		replacementErrors: []error{},
	}

	assertSuccessfulReplacement(&dummyResource, &expected, t)
}

func TestGenericReplacement_int(t *testing.T) {
	dummyResource := Resource{
		TemplateData: map[string]interface{}{
			"namespace": "<namespace>",
			"spec": map[string]interface{}{
				"replicas": "<replicas>",
			},
		},
		Data: map[string]interface{}{
			"namespace": "default",
			"replicas":  1,
		},
	}

	replaceInner(&dummyResource, &dummyResource.TemplateData, genericReplacement)

	expected := Resource{
		TemplateData: map[string]interface{}{
			"namespace": "default",
			"spec": map[string]interface{}{
				"replicas": 1,
			},
		},
		Data: map[string]interface{}{
			"namespace": "default",
			"replicas":  1,
		},
		replacementErrors: []error{},
	}

	assertSuccessfulReplacement(&dummyResource, &expected, t)
}

func TestGenericReplacement_missingValue(t *testing.T) {
	dummyResource := Resource{
		TemplateData: map[string]interface{}{
			"namespace": "<namespace>",
			"spec": map[string]interface{}{
				"replicas": "<replicas>",
			},
		},
		Data: map[string]interface{}{
			"namespace": "default",
		},
	}

	replaceInner(&dummyResource, &dummyResource.TemplateData, genericReplacement)

	expected := Resource{
		TemplateData: map[string]interface{}{
			"namespace": "default",
			"spec": map[string]interface{}{
				"replicas": "<replicas>",
			},
		},
		Data: map[string]interface{}{
			"namespace": "default",
		},
		replacementErrors: []error{
			errors.New("replaceString: missing Vault value for placeholder replicas in string replicas: <replicas>"),
		},
	}

	assertFailedReplacement(&dummyResource, &expected, t)
}

func TestSecretReplacement(t *testing.T) {
	dummyResource := Resource{
		TemplateData: map[string]interface{}{
			"namespace": "<namespace | base64encode>",
			"image":     "foo.io/<name>:<tag>",
		},
		Data: map[string]interface{}{
			"namespace": "default",
			"name":      "app",
			"tag":       "latest",
		},
	}

	replaceInner(&dummyResource, &dummyResource.TemplateData, secretReplacement)

	expected := Resource{
		TemplateData: map[string]interface{}{
			"namespace": []uint8("default"),
			"image":     "foo.io/app:latest",
		},
		Data: map[string]interface{}{
			"namespace": "default",
			"name":      "app",
			"tag":       "latest",
		},
		replacementErrors: []error{},
	}

	assertSuccessfulReplacement(&dummyResource, &expected, t)
}

func TestSecretReplacement_Base64(t *testing.T) {
	dummyResource := Resource{
		TemplateData: map[string]interface{}{
			"namespace": "PG5hbWVzcGFjZSB8IGJhc2U2NGVuY29kZT4=",
			"image":     "foo.io/<name>:<tag>",
		},
		Data: map[string]interface{}{
			"namespace": "default",
			"name":      "app",
			"tag":       "latest",
		},
	}

	replaceInner(&dummyResource, &dummyResource.TemplateData, secretReplacement)

	expected := Resource{
		TemplateData: map[string]interface{}{
			"namespace": []uint8("default"),
			"image":     "foo.io/app:latest",
		},
		Data: map[string]interface{}{
			"namespace": "default",
			"name":      "app",
			"tag":       "latest",
		},
		replacementErrors: []error{},
	}

	assertSuccessfulReplacement(&dummyResource, &expected, t)
}

func TestStringify(t *testing.T) {
	testCases := []struct {
		input    interface{}
		expected string
	}{
		{
			"thing",
			"thing",
		},
		{
			123,
			"123",
		},
		{
			true,
			"true",
		},
		{
			json.Number("123"),
			"123",
		},
	}

	for _, tc := range testCases {
		out := stringify(tc.input)
		if out != tc.expected {
			t.Errorf("expected: %s, got: %s.", tc.expected, out)
		}
	}
}
