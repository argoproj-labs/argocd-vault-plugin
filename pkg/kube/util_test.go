package kube

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/IBM/argocd-vault-plugin/pkg/helpers"
	"github.com/IBM/argocd-vault-plugin/pkg/types"
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
		Annotations: map[string]string{
			(types.AVPPathAnnotation): "",
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

func TestGenericReplacement_specificPath(t *testing.T) {
	// Test that the specific-path placeholder syntax is used to find/replace placeholders
	// along with the generic syntax, since the generic Vault path is defined
	mv := helpers.MockVault{}
	mv.LoadData(map[string]interface{}{
		"namespace": "default",
	})

	dummyResource := Resource{
		TemplateData: map[string]interface{}{
			"namespace": "<path:blah/blah#namespace>",
			"name":      "<name>",
		},
		Data: map[string]interface{}{
			"namespace": "something-else",
			"name":      "foo",
		},
		Backend: &mv,
		Annotations: map[string]string{
			(types.AVPPathAnnotation): "",
		},
	}

	replaceInner(&dummyResource, &dummyResource.TemplateData, genericReplacement)

	if !mv.GetSecretsCalled {
		t.Fatalf("expected GetSecrets to be called since placeholder contains explicit path so Vault lookup is neeed")
	}

	expected := Resource{
		TemplateData: map[string]interface{}{
			"namespace": "default",
			"name":      "foo",
		},
		Data: map[string]interface{}{
			"namespace": "something-else",
			"name":      "foo",
		},
		replacementErrors: []error{},
	}

	assertSuccessfulReplacement(&dummyResource, &expected, t)
}

func TestGenericReplacement_specificPathVersioned(t *testing.T) {
	// Test that the specific-path placeholder syntax with versioning is used to find/replace placeholders
	mv := helpers.MockVault{}
	mv.LoadData(map[string]interface{}{
		"version": "one",
	})
	mv.LoadData(map[string]interface{}{
		"version": "two",
	})
	mv.LoadData(map[string]interface{}{
		"version": "three",
	})

	dummyResource := Resource{
		TemplateData: map[string]interface{}{
			"first":  "<path:blah/blah#version#1>",
			"second": "<path:blah/blah#version#2>",
			"third":  "<path:blah/blah#version#3>",
			"latest": "<path:blah/blah#version>",
		},
		Data:    map[string]interface{}{},
		Backend: &mv,
		Annotations: map[string]string{
			(types.AVPPathAnnotation): "",
		},
	}

	replaceInner(&dummyResource, &dummyResource.TemplateData, genericReplacement)

	if !mv.GetSecretsCalled {
		t.Fatalf("expected GetSecrets to be called since placeholder contains explicit path so Vault lookup is neeed")
	}

	expected := Resource{
		TemplateData: map[string]interface{}{
			"first":  "one",
			"second": "two",
			"third":  "three",
			"latest": "three",
		},
		Data:              map[string]interface{}{},
		replacementErrors: []error{},
	}

	assertSuccessfulReplacement(&dummyResource, &expected, t)
}

func TestGenericReplacement_specificPathNoAnnotation(t *testing.T) {
	mv := helpers.MockVault{}
	mv.LoadData(map[string]interface{}{
		"namespace": "default",
	})

	// Test that the specific-path placeholder syntax is used to find/replace placeholders
	// and NOT the generic one, since the generic Vault path is undefined
	dummyResource := Resource{
		TemplateData: map[string]interface{}{
			"namespace":   "<path:blah/blah#namespace>",
			"description": "for example, write <key>",
		},
		Data: map[string]interface{}{
			"namespace": "something-else",
		},
		Backend:     &mv,
		Annotations: map[string]string{},
	}

	replaceInner(&dummyResource, &dummyResource.TemplateData, genericReplacement)

	if !mv.GetSecretsCalled {
		t.Fatalf("expected GetSecrets to be called since placeholder contains explicit path, was not")
	}

	expected := Resource{
		TemplateData: map[string]interface{}{
			"namespace":   "default",
			"description": "for example, write <key>",
		},
		Data: map[string]interface{}{
			"namespace": "something-else",
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
		Annotations: map[string]string{
			(types.AVPPathAnnotation): "",
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
		Annotations: map[string]string{
			(types.AVPPathAnnotation): "",
		},
	}

	replaceInner(&dummyResource, &dummyResource.TemplateData, genericReplacement)

	expected := Resource{
		TemplateData: map[string]interface{}{
			"namespace": base64.StdEncoding.EncodeToString([]byte("default")),
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
		Annotations: map[string]string{
			(types.AVPPathAnnotation): "",
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
		Annotations: map[string]string{
			(types.AVPPathAnnotation): "",
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
		Annotations: map[string]string{
			(types.AVPPathAnnotation): "",
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
			&missingKeyError{
				s: fmt.Sprint("replaceString: missing Vault value for placeholder replicas in string replicas: <replicas>"),
			},
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
		Annotations: map[string]string{
			(types.AVPPathAnnotation): "",
		},
	}

	replaceInner(&dummyResource, &dummyResource.TemplateData, secretReplacement)

	expected := Resource{
		TemplateData: map[string]interface{}{
			"namespace": base64.StdEncoding.EncodeToString([]byte("default")),
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
		Annotations: map[string]string{
			(types.AVPPathAnnotation): "",
		},
	}

	replaceInner(&dummyResource, &dummyResource.TemplateData, secretReplacement)

	expected := Resource{
		TemplateData: map[string]interface{}{
			"namespace": "WkdWbVlYVnNkQT09",
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

func TestSecretReplacement_Base64Substrings(t *testing.T) {
	dummyResource := Resource{
		TemplateData: map[string]interface{}{
			"data": map[string]interface{}{
				"credentials": `W2RlZmF1bHRdCmF3c19hY2Nlc3Nfa2V5X2lkPTxhY2Nlc3Nfa2V5X2lkPgphd3Nfc2VjcmV0X2FjY2Vzc19rZXk9PHNlY3JldF9hY2Nlc3Nfa2V5X2lkPgo=`,
			},
		},
		Data: map[string]interface{}{
			"access_key_id":        "testkey",
			"secret_access_key_id": "testsecret",
		},
		Annotations: map[string]string{
			(types.AVPPathAnnotation): "",
		},
	}

	replaceInner(&dummyResource, &dummyResource.TemplateData, secretReplacement)

	expected := Resource{
		TemplateData: map[string]interface{}{
			"data": map[string]interface{}{
				"credentials": `W2RlZmF1bHRdCmF3c19hY2Nlc3Nfa2V5X2lkPXRlc3RrZXkKYXdzX3NlY3JldF9hY2Nlc3Nfa2V5PXRlc3RzZWNyZXQK`,
			},
		},
		Data: map[string]interface{}{
			"access_key_id":        "testkey",
			"secret_access_key_id": "testsecret",
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
		{
			[]byte("bytes"),
			"bytes",
		},
	}

	for _, tc := range testCases {
		out := stringify(tc.input)
		if out != tc.expected {
			t.Errorf("expected: %s, got: %s.", tc.expected, out)
		}
	}
}
