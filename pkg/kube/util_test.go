package kube

import (
	"errors"
	"reflect"
	"testing"
)

func assertSuccessfulReplacement(actual, expected *Resource, t *testing.T) {
	if len(actual.replacementErrors) != 0 {
		t.Fatalf("expected 0 errors but got: %s", actual.replacementErrors)
	}

	if !reflect.DeepEqual(actual.templateData, expected.templateData) {
		t.Fatalf("expected replaced template to look like %s\n but got: %s", expected.templateData, actual.templateData)
	}

	if !reflect.DeepEqual(actual.vaultData, expected.vaultData) {
		t.Fatalf("expected Vault map to look like %s\n but got: %s", expected.vaultData, actual.vaultData)
	}
}

func assertFailedReplacement(actual, expected *Resource, t *testing.T) {
	if !reflect.DeepEqual(actual.replacementErrors, expected.replacementErrors) {
		t.Fatalf("expected replacementErrors: %s but got %s", expected.replacementErrors, actual.replacementErrors)
	}

	if !reflect.DeepEqual(actual.templateData, expected.templateData) {
		t.Fatalf("expected replaced template to look like %s\n but got: %s", expected.templateData, actual.templateData)
	}

	if !reflect.DeepEqual(actual.vaultData, expected.vaultData) {
		t.Fatalf("expected Vault map to look like %s\n but got: %s", expected.vaultData, actual.vaultData)
	}
}

func TestGenericReplacement_simpleString(t *testing.T) {
	dummyResource := Resource{
		templateData: map[string]interface{}{
			"namespace": "<namespace>",
		},
		vaultData: map[string]interface{}{
			"namespace": "default",
		},
	}

	replaceInner(&dummyResource, &dummyResource.templateData, genericReplacement)

	expected := Resource{
		templateData: map[string]interface{}{
			"namespace": "default",
		},
		vaultData: map[string]interface{}{
			"namespace": "default",
		},
		replacementErrors: []error{},
	}

	assertSuccessfulReplacement(&dummyResource, &expected, t)
}

func TestGenericReplacement_nestedString(t *testing.T) {
	dummyResource := Resource{
		templateData: map[string]interface{}{
			"namespace": "<namespace>",
			"spec": map[string]interface{}{
				"selector": map[string]interface{}{
					"app": "<name>",
				},
			},
		},
		vaultData: map[string]interface{}{
			"namespace": "default",
			"name":      "foo",
		},
	}

	replaceInner(&dummyResource, &dummyResource.templateData, genericReplacement)

	expected := Resource{
		templateData: map[string]interface{}{
			"namespace": "default",
			"spec": map[string]interface{}{
				"selector": map[string]interface{}{
					"app": "foo",
				},
			},
		},
		vaultData: map[string]interface{}{
			"namespace": "default",
			"name":      "foo",
		},
		replacementErrors: []error{},
	}

	assertSuccessfulReplacement(&dummyResource, &expected, t)
}

func TestGenericReplacement_int(t *testing.T) {
	dummyResource := Resource{
		templateData: map[string]interface{}{
			"namespace": "<namespace>",
			"spec": map[string]interface{}{
				"replicas": "<replicas>",
			},
		},
		vaultData: map[string]interface{}{
			"namespace": "default",
			"replicas":  1,
		},
	}

	replaceInner(&dummyResource, &dummyResource.templateData, genericReplacement)

	expected := Resource{
		templateData: map[string]interface{}{
			"namespace": "default",
			"spec": map[string]interface{}{
				"replicas": 1,
			},
		},
		vaultData: map[string]interface{}{
			"namespace": "default",
			"replicas":  1,
		},
		replacementErrors: []error{},
	}

	assertSuccessfulReplacement(&dummyResource, &expected, t)
}

func TestGenericReplacement_missingValue(t *testing.T) {
	dummyResource := Resource{
		templateData: map[string]interface{}{
			"namespace": "<namespace>",
			"spec": map[string]interface{}{
				"replicas": "<replicas>",
			},
		},
		vaultData: map[string]interface{}{
			"namespace": "default",
		},
	}

	replaceInner(&dummyResource, &dummyResource.templateData, genericReplacement)

	expected := Resource{
		templateData: map[string]interface{}{
			"namespace": "default",
			"spec": map[string]interface{}{
				"replicas": "<replicas>",
			},
		},
		vaultData: map[string]interface{}{
			"namespace": "default",
		},
		replacementErrors: []error{
			errors.New("replaceString: missing Vault value for placeholder replicas in string replicas: <replicas>"),
		},
	}

	assertFailedReplacement(&dummyResource, &expected, t)
}
