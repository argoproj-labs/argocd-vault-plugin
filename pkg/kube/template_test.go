package kube

import (
	"io/ioutil"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/helpers"
	"github.com/argoproj-labs/argocd-vault-plugin/pkg/types"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestToYAML_Missing_Placeholders(t *testing.T) {
	d := Template{
		Resource{
			Kind: "Secret",
			Annotations: map[string]string{
				(types.AVPPathAnnotation): "",
			},
			TemplateData: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Secret",
				"metadata": map[string]interface{}{
					"namespace": "default",
					"name":      "some-resource",
				},
				"stringData": map[string]interface{}{
					"MY_SECRET_STRING": "<string>",
				},
			},
			Data: map[string]interface{}{},
		},
	}

	expectedErr := "Replace: could not replace all placeholders in Template:\nreplaceString: missing Vault value for placeholder string in string MY_SECRET_STRING: <string>"

	err := d.Replace()
	if err == nil {
		t.Fatalf("expected error %s but got success", expectedErr)
	}

	if expectedErr != err.Error() {
		t.Fatalf("expected error \n%s but got error \n%s", expectedErr, err.Error())
	}
}

func TestToYAML_Missing_PlaceholdersSpecificPath(t *testing.T) {
	mv := helpers.MockVault{}
	mv.LoadData(map[string]interface{}{
		"different-placeholder": "string",
	})

	d := Template{
		Resource{
			Kind:        "Secret",
			Annotations: map[string]string{},
			TemplateData: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Secret",
				"metadata": map[string]interface{}{
					"namespace": "default",
					"name":      "some-resource",
				},
				"stringData": map[string]interface{}{
					"MY_SECRET_STRING": "<path:somewhere#string>",
				},
			},
			Backend: &mv,
			Data: map[string]interface{}{
				"string": "this-wont-be-used",
			},
		},
	}

	expectedErr := "Replace: could not replace all placeholders in Template:\nreplaceString: missing Vault value for placeholder path:somewhere#string in string MY_SECRET_STRING: <path:somewhere#string>"

	err := d.Replace()
	if err == nil {
		t.Fatalf("expected error %s but got success", expectedErr)
	}

	if expectedErr != err.Error() {
		t.Fatalf("expected error \n%s but got error \n%s", expectedErr, err.Error())
	}
}

func TestToYAML_RemoveMissing(t *testing.T) {
	mv := helpers.MockVault{}

	d := Template{
		Resource{
			Kind: "Secret",
			Annotations: map[string]string{
				types.AVPPathAnnotation:          "path/to/secret",
				types.AVPRemoveMissingAnnotation: "true",
			},
			TemplateData: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Secret",
				"metadata": map[string]interface{}{
					"namespace": "default",
					"name":      "my-app",
					"annotations": map[string]interface{}{
						types.AVPPathAnnotation:          "path/to/secret",
						types.AVPRemoveMissingAnnotation: "true",
					},
				},
				"data": map[string]interface{}{
					"MY_SECRET_STRING": "<string>",
					"MY_SECRET_NUM":    "<num>",
				},
			},
			Backend: &mv,
			Data: map[string]interface{}{
				"num": "NQ==",
			},
		},
	}

	err := d.Replace()
	if err != nil {
		t.Fatalf(err.Error())
	}

	expectedData, err := ioutil.ReadFile("../../fixtures/output/secret-remove-missing.yaml")
	if err != nil {
		t.Fatalf(err.Error())
	}

	expected := string(expectedData)
	actual, err := d.ToYAML()
	if err != nil {
		t.Fatalf(err.Error())
	}

	if !strings.Contains(actual, expected) {
		t.Fatalf("expected YAML:\n%s\nbut got:\n%s\n", expected, actual)
	}
}

func TestToYAML_RemoveMissingInvalidResource(t *testing.T) {
	mv := helpers.MockVault{}

	d := Template{
		Resource{
			Kind: "Service",
			Annotations: map[string]string{
				types.AVPRemoveMissingAnnotation: "true",
				types.AVPPathAnnotation:          "path/to/secret",
			},
			TemplateData: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Service",
				"metadata": map[string]interface{}{
					"namespace": "default",
					"name":      "<name>",
					"annotations": map[string]interface{}{
						types.AVPRemoveMissingAnnotation: "true",
						types.AVPPathAnnotation:          "path/to/secret",
					},
				},
				"spec": map[string]interface{}{
					"selector": map[string]interface{}{
						"app": "<name>",
					},
					"ports": []interface{}{
						map[string]interface{}{
							"port": "<port>",
						},
					},
				},
			},
			Backend: &mv,
			Data: map[string]interface{}{
				"name": "my-app",
			},
		},
	}

	expectedErr := "Replace: could not replace all placeholders in Template:\navp.kubernetes.io/remove-missing annotation can only be used on Secret or ConfigMap resources"

	err := d.Replace()
	if err == nil {
		t.Fatalf("expected error %s but got success", expectedErr)
	}

	if expectedErr != err.Error() {
		t.Fatalf("expected error \n%s but got error \n%s", expectedErr, err.Error())
	}
}

func TestNewTemplate(t *testing.T) {

	t.Run("will GetSecrets for placeholder'd YAML", func(t *testing.T) {
		mv := helpers.MockVault{}

		template, _ := NewTemplate(unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       "Service",
				"apiVersion": "v1",
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						types.VaultKVVersionAnnotation: "1",
						types.AVPPathAnnotation:        "path/to/secret",
					},
					"namespace": "default",
					"name":      "my-app",
				},
				"spec": map[string]interface{}{
					"selector": map[string]interface{}{
						"app": "my-app",
					},
					"ports": []interface{}{
						map[string]interface{}{
							"port": "3000",
						},
					},
				},
			},
		}, &mv, nil)
		if template.Resource.Kind != "Service" {
			t.Fatalf("template should have Kind of %s, instead it has %s", "Service", template.Resource.Kind)
		}

		if !mv.GetSecretsCalled {
			t.Fatalf("template does contain <placeholders> so GetSecrets should be called")
		}
	})

	t.Run("will GetSecrets only for YAMLs w/o avp.kubernetes.io/ignore: True", func(t *testing.T) {
		mv := helpers.MockVault{}
		NewTemplate(unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       "Service",
				"apiVersion": "v1",
				"metadata": map[string]interface{}{
					"namespace": "default",
					"name":      "my-app",
					"annotations": map[string]interface{}{
						types.AVPIgnoreAnnotation: "True",
					},
				},
				"spec": map[string]interface{}{
					"selector": map[string]interface{}{
						"app": "my-app",
					},
					"ports": []interface{}{
						map[string]interface{}{
							"port": "<port>",
						},
					},
				},
			},
		}, &mv, nil)
		if mv.GetSecretsCalled {
			t.Fatalf("template contains avp.kubernetes.io/ignore:True so GetSecrets should NOT be called")
		}
	})

	t.Run("will GetSecrets with version given in avp.kubernetes.io/secret-version", func(t *testing.T) {
		mv := helpers.MockVault{}

		mv.LoadData(map[string]interface{}{
			"password": "original-value",
		})
		mv.LoadData(map[string]interface{}{
			"password": "changed-value",
		})

		template, _ := NewTemplate(unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       "Secret",
				"apiVersion": "v1",
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						types.AVPSecretVersionAnnotation: "1",
						types.AVPPathAnnotation:          "path/to/secret",
					},
					"namespace": "default",
					"name":      "my-app",
				},
				"data": map[string]interface{}{
					"new-value": "<path:/path/to/secret#password#2>",
					"old-value": "<password>",
				},
			},
		}, &mv, nil)

		if template.Resource.Kind != "Secret" {
			t.Fatalf("template should have Kind of %s, instead it has %s", "Secret", template.Resource.Kind)
		}

		if !mv.GetSecretsCalled {
			t.Fatalf("template does contain <placeholders> so GetSecrets should be called")
		}

		err := template.Replace()
		if err != nil {
			t.Fatalf(err.Error())
		}

		expected := map[string]interface{}{
			"kind":       "Secret",
			"apiVersion": "v1",
			"metadata": map[string]interface{}{
				"annotations": map[string]interface{}{
					types.AVPSecretVersionAnnotation: "1",
					types.AVPPathAnnotation:          "path/to/secret",
				},
				"namespace": "default",
				"name":      "my-app",
			},
			"data": map[string]interface{}{
				"new-value": "changed-value",
				"old-value": "original-value",
			},
		}

		if !reflect.DeepEqual(expected, template.TemplateData) {
			t.Fatalf("expected %s got %s", expected, template.TemplateData)
		}
	})

	t.Run("will GetSecrets with latest version by default", func(t *testing.T) {
		mv := helpers.MockVault{}

		mv.LoadData(map[string]interface{}{
			"password": "original-value",
		})
		mv.LoadData(map[string]interface{}{
			"password": "changed-value",
		})

		template, _ := NewTemplate(unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       "Secret",
				"apiVersion": "v1",
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						types.AVPPathAnnotation: "path/to/secret",
					},
					"namespace": "default",
					"name":      "my-app",
				},
				"data": map[string]interface{}{
					"old-value": "<path:/path/to/secret#password#1>",
					"new-value": "<password>",
				},
			},
		}, &mv, nil)

		if template.Resource.Kind != "Secret" {
			t.Fatalf("template should have Kind of %s, instead it has %s", "Secret", template.Resource.Kind)
		}

		if !mv.GetSecretsCalled {
			t.Fatalf("template does contain <placeholders> so GetSecrets should be called")
		}

		err := template.Replace()
		if err != nil {
			t.Fatalf(err.Error())
		}

		expected := map[string]interface{}{
			"kind":       "Secret",
			"apiVersion": "v1",
			"metadata": map[string]interface{}{
				"annotations": map[string]interface{}{
					types.AVPPathAnnotation: "path/to/secret",
				},
				"namespace": "default",
				"name":      "my-app",
			},
			"data": map[string]interface{}{
				"new-value": "changed-value",
				"old-value": "original-value",
			},
		}

		if !reflect.DeepEqual(expected, template.TemplateData) {
			t.Fatalf("expected %s got %s", expected, template.TemplateData)
		}
	})

	t.Run("GetSecrets with path validation and invalid path", func(t *testing.T) {
		mv := helpers.MockVault{}

		mv.LoadData(map[string]interface{}{
			"password": "original-value",
		})
		mv.LoadData(map[string]interface{}{
			"password": "changed-value",
		})

		template, err := NewTemplate(unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       "Secret",
				"apiVersion": "v1",
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						types.AVPPathAnnotation: "path/to/secret",
					},
					"namespace": "default",
					"name":      "my-app",
				},
				"data": map[string]interface{}{
					"old-value": "<path:/path/to/secret#password#1>",
					"new-value": "<password>",
				},
			},
		}, &mv, regexp.MustCompile(`/[A-Z]/`))

		if template != nil {
			t.Fatalf("expected template to be nil got %s", template)
		}
		if err == nil {
			t.Fatalf("expected error got nil")
		}

		expected := "the path path/to/secret is disallowed by AVP_PATH_VALIDATION restriction"
		if err.Error() != expected {
			t.Fatalf("expected %s got %s", expected, err.Error())
		}
	})

	t.Run("will GetSecrets with latest version by default and path validation regexp", func(t *testing.T) {
		mv := helpers.MockVault{}

		mv.LoadData(map[string]interface{}{
			"password": "original-value",
		})
		mv.LoadData(map[string]interface{}{
			"password": "changed-value",
		})

		template, _ := NewTemplate(unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       "Secret",
				"apiVersion": "v1",
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						types.AVPPathAnnotation: "path/to/secret",
					},
					"namespace": "default",
					"name":      "my-app",
				},
				"data": map[string]interface{}{
					"old-value": "<path:/path/to/secret#password#1>",
					"new-value": "<password>",
				},
			},
		}, &mv, regexp.MustCompile(`^([A-Za-z/]*)$`))

		if template.Resource.Kind != "Secret" {
			t.Fatalf("template should have Kind of %s, instead it has %s", "Secret", template.Resource.Kind)
		}

		if !mv.GetSecretsCalled {
			t.Fatalf("template does contain <placeholders> so GetSecrets should be called")
		}

		err := template.Replace()
		if err != nil {
			t.Fatalf(err.Error())
		}

		expected := map[string]interface{}{
			"kind":       "Secret",
			"apiVersion": "v1",
			"metadata": map[string]interface{}{
				"annotations": map[string]interface{}{
					types.AVPPathAnnotation: "path/to/secret",
				},
				"namespace": "default",
				"name":      "my-app",
			},
			"data": map[string]interface{}{
				"new-value": "changed-value",
				"old-value": "original-value",
			},
		}

		if !reflect.DeepEqual(expected, template.TemplateData) {
			t.Fatalf("expected %s got %s", expected, template.TemplateData)
		}
	})
}

func TestToYAML_Deployment(t *testing.T) {
	d := Template{
		Resource{
			Kind: "Deployment",
			Annotations: map[string]string{
				(types.AVPPathAnnotation): "",
			},
			TemplateData: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						types.AVPPathAnnotation: "path",
					},
					"namespace": "default",
					"name":      "<name>",
				},
				"spec": map[string]interface{}{
					"replicas": "<replicas>",
					"template": map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app": "<name>",
							},
						},
					},
				},
			},
			Data: map[string]interface{}{
				"replicas": 3,
				"name":     "my-app",
			},
		},
	}

	err := d.Replace()
	if err != nil {
		t.Fatalf(err.Error())
	}

	expectedData, err := ioutil.ReadFile("../../fixtures/output/small-deployment.yaml")
	if err != nil {
		t.Fatalf(err.Error())
	}

	expected := string(expectedData)
	actual, err := d.ToYAML()
	if err != nil {
		t.Fatalf(err.Error())
	}

	if !strings.Contains(actual, expected) {
		t.Fatalf("expected YAML:\n%s\nbut got:\n%s\n", expected, actual)
	}
}

func TestToYAML_Service(t *testing.T) {
	d := Template{
		Resource{
			Kind: "Service",
			Annotations: map[string]string{
				(types.AVPPathAnnotation): "",
			},
			TemplateData: map[string]interface{}{
				"kind":       "Service",
				"apiVersion": "v1",
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						types.AVPPathAnnotation: "path",
					},
					"namespace": "default",
					"name":      "<name>",
				},
				"spec": map[string]interface{}{
					"selector": map[string]interface{}{
						"app": "<name>",
					},
					"ports": []interface{}{
						map[string]interface{}{
							"port": "<port>",
						},
					},
				},
			},
			Data: map[string]interface{}{
				"name": "my-app",
				"port": 8080,
			},
		},
	}

	err := d.Replace()
	if err != nil {
		t.Fatalf(err.Error())
	}

	expectedData, err := ioutil.ReadFile("../../fixtures/output/small-service.yaml")
	if err != nil {
		t.Fatalf(err.Error())
	}

	expected := string(expectedData)
	actual, err := d.ToYAML()
	if err != nil {
		t.Fatalf(err.Error())
	}

	if !strings.Contains(actual, expected) {
		t.Fatalf("expected YAML:\n%s\nbut got:\n%s\n", expected, actual)
	}
}

func TestToYAML_Secret_PlaceholderedData(t *testing.T) {
	d := Template{
		Resource{
			Kind: "Secret",
			Annotations: map[string]string{
				(types.AVPPathAnnotation): "",
			},
			TemplateData: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Secret",
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						types.AVPPathAnnotation:        "path",
						types.VaultKVVersionAnnotation: "1",
					},
					"namespace": "default",
					"name":      "<name>",
				},
				"data": map[string]interface{}{
					"MY_SECRET_STRING": "<string>",
					"MY_SECRET_NUM":    "<num>",
				},
			},
			Data: map[string]interface{}{
				"name":   "my-app",
				"string": "Zm9v",
				"num":    "NQ==",
			},
		},
	}

	err := d.Replace()
	if err != nil {
		t.Fatalf(err.Error())
	}

	expectedData, err := ioutil.ReadFile("../../fixtures/output/small-secret.yaml")
	if err != nil {
		t.Fatalf(err.Error())
	}

	expected := string(expectedData)
	actual, err := d.ToYAML()
	if err != nil {
		t.Fatalf(err.Error())
	}

	if !strings.Contains(actual, expected) {
		t.Fatalf("expected YAML:\n%s\nbut got:\n%s\n", expected, actual)
	}
}

func TestToYAML_CRD_PlaceholderedData(t *testing.T) {
	d := Template{
		Resource{
			Kind: "SomeCustomResource",
			Annotations: map[string]string{
				(types.AVPPathAnnotation): "",
			},
			TemplateData: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "SomeCustomResource",
				"metadata": map[string]interface{}{
					"namespace": "default",
					"name":      "<name>",
				},
				"data": map[string]interface{}{
					"A_SEQUENCE": []interface{}{
						1,
						"<two>",
					},
					"A_YAML":         "username: <username>\npassword: <password>",
					"A_SHELL_SCRIPT": "bx login --apikey <apikey>",
				},
			},
			Data: map[string]interface{}{
				"name":     "my-app",
				"two":      "two",
				"username": "user",
				"password": "pass",
				"apikey":   "123",
			},
		},
	}

	err := d.Replace()
	if err != nil {
		t.Fatalf(err.Error())
	}

	expectedData, err := ioutil.ReadFile("../../fixtures/output/small-custom-resource.yaml")
	if err != nil {
		t.Fatalf(err.Error())
	}

	expected := string(expectedData)
	actual, err := d.ToYAML()
	if err != nil {
		t.Fatalf(err.Error())
	}

	if !strings.Contains(actual, expected) {
		t.Fatalf("expected YAML:\n%s\nbut got:\n%s\n", expected, actual)
	}
}
func TestToYAML_CRD_FakePlaceholders(t *testing.T) {
	mv := helpers.MockVault{}
	mv.LoadData(map[string]interface{}{
		"apikey": "123",
	})

	d := Template{
		Resource{
			Kind: "SomeCustomResource",
			TemplateData: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "SomeCustomResource",
				"metadata": map[string]interface{}{
					"namespace": "default",
					"name":      "some-resource",
				},
				"data": map[string]interface{}{
					"description":    "supported options: <beep>, <boop>",
					"A_SHELL_SCRIPT": "bx login --apikey <path:a/path#apikey>",
				},
			},
			Backend: &mv,
			Data:    map[string]interface{}{},
		},
	}

	err := d.Replace()
	if err != nil {
		t.Fatalf(err.Error())
	}

	expectedData, err := ioutil.ReadFile("../../fixtures/output/small-custom-resource-fake-placeholders.yaml")
	if err != nil {
		t.Fatalf(err.Error())
	}

	expected := string(expectedData)
	actual, err := d.ToYAML()
	if err != nil {
		t.Fatalf(err.Error())
	}

	if !strings.Contains(actual, expected) {
		t.Fatalf("expected YAML:\n%s\nbut got:\n%s\n", expected, actual)
	}
}

func TestToYAML_Secret_HardcodedData(t *testing.T) {
	d := Template{
		Resource{
			Kind: "Secret",
			Annotations: map[string]string{
				(types.AVPPathAnnotation): "",
			},
			TemplateData: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Secret",
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						types.AVPPathAnnotation: "path",
					},
					"namespace": "default",
					"name":      "my-app",
				},
				"data": map[string]interface{}{
					"MY_LEAKED_SECRET": "cGFzc3dvcmQ=",
				},
			},
			Data: map[string]interface{}{},
		},
	}

	err := d.Replace()
	if err != nil {
		t.Fatalf(err.Error())
	}

	expectedData, err := ioutil.ReadFile("../../fixtures/output/small-secret4.yaml")
	if err != nil {
		t.Fatalf(err.Error())
	}

	expected := string(expectedData)
	actual, err := d.ToYAML()
	if err != nil {
		t.Fatalf(err.Error())
	}

	if !strings.Contains(actual, expected) {
		t.Fatalf("expected YAML:\n%s\nbut got:\n%s\n", expected, actual)
	}
}
func TestToYAML_Secret_MixedData(t *testing.T) {
	d := Template{
		Resource{
			Kind: "Secret",
			Annotations: map[string]string{
				(types.AVPPathAnnotation): "",
			},
			TemplateData: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Secret",
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						types.AVPPathAnnotation: "path",
					},
					"namespace": "default",
					"name":      "<name>",
				},
				"data": map[string]interface{}{
					"MY_SECRET_STRING": "<string>",
					"MY_SECRET_NUM":    "<num>",
					"MY_LEAKED_SECRET": "cGFzc3dvcmQ=",
				},
			},
			Data: map[string]interface{}{
				"name":   "my-app",
				"string": "Zm9v",
				"num":    "NQ==",
			},
		},
	}

	err := d.Replace()
	if err != nil {
		t.Fatalf(err.Error())
	}

	expectedData, err := ioutil.ReadFile("../../fixtures/output/small-secret3.yaml")
	if err != nil {
		t.Fatalf(err.Error())
	}

	expected := string(expectedData)
	actual, err := d.ToYAML()
	if err != nil {
		t.Fatalf(err.Error())
	}

	if !strings.Contains(actual, expected) {
		t.Fatalf("expected YAML:\n%s\nbut got:\n%s\n", expected, actual)
	}
}

func TestToYAML_Secret_PlaceholderedStringData(t *testing.T) {
	d := Template{
		Resource{
			Kind: "Secret",
			Annotations: map[string]string{
				(types.AVPPathAnnotation): "",
			},
			TemplateData: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Secret",
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						types.AVPPathAnnotation: "path",
					},
					"namespace": "default",
					"name":      "<name>",
				},
				"stringData": map[string]interface{}{
					"MY_SECRET_STRING": "<string>",
					"MY_SECRET_NUM":    "<num>",
				},
			},
			Data: map[string]interface{}{
				"name":   "my-app",
				"string": "foo",
				"num":    5,
			},
		},
	}

	err := d.Replace()
	if err != nil {
		t.Fatalf(err.Error())
	}

	expectedData, err := ioutil.ReadFile("../../fixtures/output/small-secret2.yaml")
	if err != nil {
		t.Fatalf(err.Error())
	}

	expected := string(expectedData)
	actual, err := d.ToYAML()
	if err != nil {
		t.Fatalf(err.Error())
	}

	if !strings.Contains(actual, expected) {
		t.Fatalf("expected YAML:\n%s\nbut got:\n%s\n", expected, actual)
	}
}

func TestToYAML_ConfigMap(t *testing.T) {
	d := Template{
		Resource{
			Kind: "ConfigMap",
			Annotations: map[string]string{
				(types.AVPPathAnnotation): "",
			},
			TemplateData: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						types.AVPPathAnnotation: "path",
					},
					"namespace": "default",
					"name":      "<name>",
				},
				"data": map[string]interface{}{
					"MY_NONSECRET_STRING": "<string>",
					"MY_NONSECRET_NUM":    "<num>",
				},
			},
			Data: map[string]interface{}{
				"name":   "my-app",
				"string": "foo",
				"num":    5,
			},
		},
	}

	err := d.Replace()
	if err != nil {
		t.Fatalf(err.Error())
	}

	expectedData, err := ioutil.ReadFile("../../fixtures/output/small-configmap.yaml")
	if err != nil {
		t.Fatalf(err.Error())
	}

	expected := string(expectedData)
	actual, err := d.ToYAML()
	if err != nil {
		t.Fatalf(err.Error())
	}

	if !strings.Contains(actual, expected) {
		t.Fatalf("expected YAML:\n%s\nbut got:\n%s\n", expected, actual)
	}
}

func TestToYAML_Ingress(t *testing.T) {
	d := Template{
		Resource{
			Kind: "Ingress",
			Annotations: map[string]string{
				(types.AVPPathAnnotation): "",
			},
			TemplateData: map[string]interface{}{
				"apiVersion": "networking.k8s.io/v1",
				"kind":       "Ingress",
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						types.AVPPathAnnotation: "path",
					},
					"namespace": "default",
					"name":      "<name>",
				},
				"spec": map[string]interface{}{
					"tls": []interface{}{
						map[string]interface{}{
							"hosts": []interface{}{
								"mysubdomain.<host>",
							},
							"secretName": "<secret>",
						},
					},
				},
			},
			Data: map[string]interface{}{
				"name":   "my-app",
				"host":   "foo.com",
				"secret": "foo-secret",
			},
		},
	}

	err := d.Replace()
	if err != nil {
		t.Fatalf(err.Error())
	}

	expectedData, err := ioutil.ReadFile("../../fixtures/output/small-ingress.yaml")
	if err != nil {
		t.Fatalf(err.Error())
	}

	expected := string(expectedData)
	actual, err := d.ToYAML()
	if err != nil {
		t.Fatalf(err.Error())
	}

	if !strings.Contains(actual, expected) {
		t.Fatalf("expected YAML:\n%s\nbut got:\n%s\n", expected, actual)
	}
}

func TestToYAML_CronJob(t *testing.T) {
	d := Template{
		Resource{
			Kind: "CronJob",
			Annotations: map[string]string{
				(types.AVPPathAnnotation): "",
			},
			TemplateData: map[string]interface{}{
				"apiVersion": "batch/v1beta1",
				"kind":       "CronJob",
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						types.AVPPathAnnotation: "path",
					},
					"name": "<name>",
				},
				"spec": map[string]interface{}{
					"schedule": "0 0 0 0 0",
					"jobTemplate": map[string]interface{}{
						"spec": map[string]interface{}{
							"template": map[string]interface{}{
								"spec": map[string]interface{}{
									"containers": []interface{}{
										map[string]interface{}{
											"image": "<name>:<tag>",
											"name":  "<name>",
										},
									},
								},
							},
						},
					},
				},
			},
			Data: map[string]interface{}{
				"name": "my-app",
				"tag":  "latest",
			},
		},
	}

	err := d.Replace()
	if err != nil {
		t.Fatalf(err.Error())
	}

	expectedData, err := ioutil.ReadFile("../../fixtures/output/small-cronjob.yaml")
	if err != nil {
		t.Fatalf(err.Error())
	}

	expected := string(expectedData)
	actual, err := d.ToYAML()
	if err != nil {
		t.Fatalf(err.Error())
	}

	if !strings.Contains(actual, expected) {
		t.Fatalf("expected YAML:\n%s\nbut got:\n%s\n", expected, actual)
	}
}

func TestToYAML_Job(t *testing.T) {
	d := Template{
		Resource{
			Kind: "Job",
			Annotations: map[string]string{
				(types.AVPPathAnnotation): "",
			},
			TemplateData: map[string]interface{}{
				"apiVersion": "batch/v1",
				"kind":       "Job",
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						types.AVPPathAnnotation: "path",
					},
					"name": "<name>",
				},
				"spec": map[string]interface{}{
					"template": map[string]interface{}{
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"image": "<name>:<tag>",
									"name":  "<name>",
								},
							},
						},
					},
				},
			},
			Data: map[string]interface{}{
				"name": "my-app",
				"tag":  "latest",
			},
		},
	}

	err := d.Replace()
	if err != nil {
		t.Fatalf(err.Error())
	}

	expectedData, err := ioutil.ReadFile("../../fixtures/output/small-job.yaml")
	if err != nil {
		t.Fatalf(err.Error())
	}

	expected := string(expectedData)
	actual, err := d.ToYAML()
	if err != nil {
		t.Fatalf(err.Error())
	}

	if !strings.Contains(actual, expected) {
		t.Fatalf("expected YAML:\n%s\nbut got:\n%s\n", expected, actual)
	}
}
