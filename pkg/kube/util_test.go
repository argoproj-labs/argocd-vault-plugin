package kube

import (
	"encoding/json"
	"errors"
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
		Annotations: map[string]string{
			(types.AVPPathAnnotation): "",
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
		Annotations: map[string]string{
			(types.AVPPathAnnotation): "",
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
