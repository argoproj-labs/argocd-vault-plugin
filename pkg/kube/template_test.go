package kube

import (
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
)

type MockVault struct {
	GetSecretsCalled bool
}

func (v *MockVault) Login() error {
	return nil
}
func (v *MockVault) GetSecrets(path, kvVersion string) (map[string]interface{}, error) {
	v.GetSecretsCalled = true
	return map[string]interface{}{
		"foo": "bar-value",
		"baz": "baz-value",
	}, nil
}

func TestNewTemplate(t *testing.T) {
	t.Run("will not GetSecrets for non-placeholder'd YAML", func(t *testing.T) {
		mv := MockVault{}
		template, _ := NewTemplate(map[string]interface{}{
			"kind":       "Service",
			"apiVersion": "v1",
			"metadata": map[string]interface{}{
				"annotations": map[string]interface{}{
					"kv_version": "1",
					"avp_path":   "path/to/secret",
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
		}, &mv, "string")
		if template.Resource.replaceable {
			t.Fatalf("template does not contain <placeholders> and shouldn't be replaced")
		}
		if mv.GetSecretsCalled {
			t.Fatalf("template does not contain <placeholders> so GetSecrets shouldn't be called")
		}
	})
	t.Run("will GetSecrets for placeholder'd YAML", func(t *testing.T) {
		mv := MockVault{}
		template, _ := NewTemplate(map[string]interface{}{
			"kind":       "Service",
			"apiVersion": "v1",
			"metadata": map[string]interface{}{
				"namespace": "default",
				"name":      "my-app",
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
		}, &mv, "string")
		if !template.Resource.replaceable {
			t.Fatalf("template does contain <placeholders> and should be replaced")
		}
		if !mv.GetSecretsCalled {
			t.Fatalf("template does contain <placeholders> so GetSecrets should be called")
		}
	})
	t.Run("will GetSecrets only for YAMLs w/o avp_ignore: True", func(t *testing.T) {
		mv := MockVault{}
		template, _ := NewTemplate(map[string]interface{}{
			"kind":       "Service",
			"apiVersion": "v1",
			"metadata": map[string]interface{}{
				"namespace": "default",
				"name":      "my-app",
				"annotations": map[string]interface{}{
					"avp_ignore": "True",
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
		}, &mv, "string")
		if template.Resource.replaceable {
			t.Fatalf("template contains avp_ignore:True and should NOT be replaced")
		}
		if mv.GetSecretsCalled {
			t.Fatalf("template contains avp_ignore:True so GetSecrets should NOT be called")
		}
	})

	t.Run("will auto template all 'data' keys when avp_all_keys_to: data", func(t *testing.T) {
		mv := MockVault{}
		template, _ := NewTemplate(map[string]interface{}{
			"kind":       "Secret",
			"apiVersion": "v1",
			"metadata": map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Secret",
				"annotations": map[string]interface{}{
					"avp_all_keys_to": "data",
				},
			},
		}, &mv, "")

		if !template.Resource.replaceable {
			t.Fatalf("template contains avp_all_keys_to:data and should be replaced")
		}

		if !mv.GetSecretsCalled {
			t.Fatalf("template contains avp_all_keys_to:data so GetSecrets should be called")
		}

		data := template.TemplateData["data"].(map[string]interface{})
		if data["foo"].(string) != "<foo>" {
			t.Fatalf("template contains avp_all_keys_to:data and should have a single key matching the secret")
		}

		if data["baz"].(string) != "<baz>" {
			t.Fatalf("template contains avp_all_keys_to:data and should have a single key matching the secret")
		}
	})

	t.Run("will auto template all 'stringData' keys when avp_all_keys_to: stringData", func(t *testing.T) {
		mv := MockVault{}
		template, _ := NewTemplate(map[string]interface{}{
			"kind":       "Secret",
			"apiVersion": "v1",
			"metadata": map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Secret",
				"annotations": map[string]interface{}{
					"avp_all_keys_to": "stringData",
				},
			},
		}, &mv, "")

		if !template.Resource.replaceable {
			t.Fatalf("template contains avp_all_keys_to:data and should be replaced")
		}

		if !mv.GetSecretsCalled {
			t.Fatalf("template contains avp_all_keys_to:data so GetSecrets should be called")
		}

		stringData := template.TemplateData["stringData"].(map[string]interface{})
		if stringData["foo"].(string) != "<foo>" {
			t.Fatalf("template contains avp_all_keys_to:data and should have a single key matching the secret")
		}

		if stringData["baz"].(string) != "<baz>" {
			t.Fatalf("template contains avp_all_keys_to:data and should have a single key matching the secret")
		}
	})

	t.Run("Cannot use avp_all_keys_to: stringData with ConfigMap", func(t *testing.T) {
		mv := MockVault{}
		_, err := NewTemplate(map[string]interface{}{
			"kind":       "ConfigMap",
			"apiVersion": "v1",
			"metadata": map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Secret",
				"annotations": map[string]interface{}{
					"avp_all_keys_to": "stringData",
				},
			},
		}, &mv, "")

		expected := "stringData can only be used with Secret! Got configmap"
		actual := err.Error()

		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("expected: %s, got: %s.", expected, err.Error())
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
			replaceable: true,
			VaultData: map[string]interface{}{
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
			replaceable: true,
			VaultData: map[string]interface{}{
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
					"namespace": "default",
					"name":      "<name>",
				},
				"data": map[string]interface{}{
					"MY_SECRET_STRING": "<string>",
					"MY_SECRET_NUM":    "<num>",
				},
			},
			replaceable: true,
			VaultData: map[string]interface{}{
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

func TestToYAML_Secret_HardcodedData(t *testing.T) {
	d := Template{
		Resource{
			Kind: "Secret",
			TemplateData: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Secret",
				"metadata": map[string]interface{}{
					"namespace": "default",
					"name":      "my-app",
				},
				"data": map[string]interface{}{
					"MY_LEAKED_SECRET": "cGFzc3dvcmQ=",
				},
			},
			replaceable: true,
			VaultData:   map[string]interface{}{},
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
					"namespace": "default",
					"name":      "<name>",
				},
				"data": map[string]interface{}{
					"MY_SECRET_STRING": "<string>",
					"MY_SECRET_NUM":    "<num>",
					"MY_LEAKED_SECRET": "cGFzc3dvcmQ=",
				},
			},
			replaceable: true,
			VaultData: map[string]interface{}{
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
					"namespace": "default",
					"name":      "<name>",
				},
				"stringData": map[string]interface{}{
					"MY_SECRET_STRING": "<string>",
					"MY_SECRET_NUM":    "<num>",
				},
			},
			replaceable: true,
			VaultData: map[string]interface{}{
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
					"namespace": "default",
					"name":      "<name>",
				},
				"data": map[string]interface{}{
					"MY_NONSECRET_STRING": "<string>",
					"MY_NONSECRET_NUM":    "<num>",
				},
			},
			replaceable: true,
			VaultData: map[string]interface{}{
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
			replaceable: true,
			VaultData: map[string]interface{}{
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
			replaceable: true,
			VaultData: map[string]interface{}{
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
			replaceable: true,
			VaultData: map[string]interface{}{
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

func TestToYAML_DeploymentBad(t *testing.T) {
	d := Template{
		Resource{
			Kind: "Deployment",
			TemplateData: map[string]interface{}{
				"metadata": map[string]interface{}{
					"namespace": "default",
					"name":      "<name>",
				},
				"spec": map[string]interface{}{
					"replicas":        "<replicas>",
					"minReadySeconds": "<minReadySeconds>",
					"template": map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app": "<name>",
							},
						},
					},
				},
			},
			replaceable: true,
			VaultData: map[string]interface{}{
				"replicas":        3,
				"minReadySeconds": "one hundred",
				"name":            "!!@#@.---.",
			},
		},
	}

	err := d.Replace()
	if err != nil {
		t.Fatalf(err.Error())
	}

	actual, err := d.ToYAML()
	if err == nil {
		t.Fatalf("Expected ToYAML error but got %s", actual)
	}

	if !strings.Contains(err.Error(), "Object 'Kind' is missing") {
		t.Fatalf("Expected error 'Object 'Kind' is missing', got: %s", err.Error())
	}
}
