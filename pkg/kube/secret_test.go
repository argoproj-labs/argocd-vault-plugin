package kube

// import (
// 	"testing"
// )
//
// func TestSecretReplacement_string(t *testing.T) {
// 	secret := SecretTemplate{
// 		Resource{
// 			templateData: map[string]interface{}{
// 				"metadata": map[string]interface{}{
// 					"namespace": "<namespace>",
// 				},
// 				"data": map[string]interface{}{
// 					"PASSWORD": "<password>",
// 				},
// 			},
// 			vaultData: map[string]interface{}{
// 				"namespace": "default",
// 				"password":  "foo",
// 			},
// 		},
// 	}
//
// 	err := secret.Replace()
// 	if err != nil {
// 		t.Fatalf(err.Error())
// 	}
//
// 	expected := Resource{
// 		templateData: map[string]interface{}{
// 			"metadata": map[string]interface{}{
// 				"namespace": "default",
// 			},
// 			"data": map[string]interface{}{
// 				"PASSWORD": []byte("foo"),
// 			},
// 		},
// 		vaultData: map[string]interface{}{
// 			"namespace": "default",
// 			"password":  "foo",
// 		},
// 		replacementErrors: []error{},
// 	}
//
// 	assertSuccessfulReplacement(&secret.Resource, &expected, t)
// }
//
// func TestSecretReplacement_int(t *testing.T) {
// 	secret := SecretTemplate{
// 		Resource{
// 			templateData: map[string]interface{}{
// 				"metadata": map[string]interface{}{
// 					"namespace": "<namespace>",
// 				},
// 				"data": map[string]interface{}{
// 					"PORT": "<port>",
// 				},
// 			},
// 			vaultData: map[string]interface{}{
// 				"namespace": "default",
// 				"port":      5,
// 			},
// 		},
// 	}
//
// 	err := secret.Replace()
// 	if err != nil {
// 		t.Fatalf(err.Error())
// 	}
//
// 	expected := Resource{
// 		templateData: map[string]interface{}{
// 			"metadata": map[string]interface{}{
// 				"namespace": "default",
// 			},
// 			"data": map[string]interface{}{
// 				"PORT": []byte("5"),
// 			},
// 		},
// 		vaultData: map[string]interface{}{
// 			"namespace": "default",
// 			"port":      5,
// 		},
// 		replacementErrors: []error{},
// 	}
//
// 	assertSuccessfulReplacement(&secret.Resource, &expected, t)
// }
//
// func TestSecretReplacement_bool(t *testing.T) {
// 	secret := SecretTemplate{
// 		Resource{
// 			templateData: map[string]interface{}{
// 				"metadata": map[string]interface{}{
// 					"namespace": "<namespace>",
// 				},
// 				"data": map[string]interface{}{
// 					"SECRET_FLAG": "<flag>",
// 				},
// 			},
// 			vaultData: map[string]interface{}{
// 				"namespace": "default",
// 				"flag":      true,
// 			},
// 		},
// 	}
//
// 	err := secret.Replace()
// 	if err != nil {
// 		t.Fatalf(err.Error())
// 	}
//
// 	expected := Resource{
// 		templateData: map[string]interface{}{
// 			"metadata": map[string]interface{}{
// 				"namespace": "default",
// 			},
// 			"data": map[string]interface{}{
// 				"SECRET_FLAG": []byte("true"),
// 			},
// 		},
// 		vaultData: map[string]interface{}{
// 			"namespace": "default",
// 			"flag":      true,
// 		},
// 		replacementErrors: []error{},
// 	}
//
// 	assertSuccessfulReplacement(&secret.Resource, &expected, t)
// }
