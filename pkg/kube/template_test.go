package kube

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/IBM/argocd-vault-plugin/pkg/helpers"
	"github.com/IBM/argocd-vault-plugin/pkg/types"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestToYAML_Missing_Placeholders(t *testing.T) {
	d := Template{
		Resource{
			Kind: "Secret",
			TemplateData: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Secret",
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						types.AVPPathAnnotation: "path",
					},
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
		}, &mv)
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
		}, &mv)
		if mv.GetSecretsCalled {
			t.Fatalf("template contains avp.kubernetes.io/ignore:True so GetSecrets should NOT be called")
		}
	})
}

func TestToYAML_Deployment(t *testing.T) {
	d := Template{
		Resource{
			Kind: "Deployment",
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
