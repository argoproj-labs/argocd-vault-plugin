package kube

import (
	"testing"
)

func TestSecretReplacement_string(t *testing.T) {
	secret := SecretTemplate{
		Resource{
			templateData: map[string]interface{}{
				"metadata": map[string]interface{}{
					"namespace": "<namespace>",
				},
				"data": map[string]interface{}{
					"PASSWORD": "<password>",
				},
			},
			vaultData: map[string]interface{}{
				"namespace": "default",
				"password":  "foo",
			},
		},
	}

	secret.Replace()

	expected := Resource{
		templateData: map[string]interface{}{
			"metadata": map[string]interface{}{
				"namespace": "default",
			},
			"data": map[string]interface{}{
				"PASSWORD": []byte("foo"),
			},
		},
		vaultData: map[string]interface{}{
			"namespace": "default",
			"password":  "foo",
		},
		replacementErrors: []error{},
	}

	assertSuccessfulReplacement(&secret.Resource, &expected, t)
}
