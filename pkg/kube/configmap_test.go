package kube

// import (
// 	"testing"
// )
//
// func TestConfigMapReplacement_string(t *testing.T) {
// 	configmap := ConfigMapTemplate{
// 		Resource{
// 			templateData: map[string]interface{}{
// 				"metadata": map[string]interface{}{
// 					"namespace": "<namespace>",
// 				},
// 				"data": map[string]interface{}{
// 					"SOME_ENV_VAR": "<env-var>",
// 				},
// 			},
// 			vaultData: map[string]interface{}{
// 				"namespace": "default",
// 				"env-var":   "foo",
// 			},
// 		},
// 	}
//
// 	err := configmap.Replace()
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
// 				"SOME_ENV_VAR": "foo",
// 			},
// 		},
// 		vaultData: map[string]interface{}{
// 			"namespace": "default",
// 			"env-var":   "foo",
// 		},
// 		replacementErrors: []error{},
// 	}
//
// 	assertSuccessfulReplacement(&configmap.Resource, &expected, t)
// }
//
// func TestConfigMapReplacement_int(t *testing.T) {
// 	configmap := ConfigMapTemplate{
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
// 	err := configmap.Replace()
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
// 				"PORT": "5",
// 			},
// 		},
// 		vaultData: map[string]interface{}{
// 			"namespace": "default",
// 			"port":      5,
// 		},
// 		replacementErrors: []error{},
// 	}
//
// 	assertSuccessfulReplacement(&configmap.Resource, &expected, t)
// }
//
// func TestConfigMapReplacement_bool(t *testing.T) {
// 	configmap := ConfigMapTemplate{
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
// 	err := configmap.Replace()
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
// 				"SECRET_FLAG": "true",
// 			},
// 		},
// 		vaultData: map[string]interface{}{
// 			"namespace": "default",
// 			"flag":      true,
// 		},
// 		replacementErrors: []error{},
// 	}
//
// 	assertSuccessfulReplacement(&configmap.Resource, &expected, t)
// }
