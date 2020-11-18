package kube

import (
	"io/ioutil"
	"strings"
	"testing"
)

func TestToYAML_Deployment(t *testing.T) {
	d := DeploymentTemplate{
		Resource{
			templateData: map[string]interface{}{
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
			vaultData: map[string]interface{}{
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

func TestToYAML_Secret(t *testing.T) {
	d := SecretTemplate{
		Resource{
			templateData: map[string]interface{}{
				"metadata": map[string]interface{}{
					"namespace": "default",
					"name":      "<name>",
				},
				"data": map[string]interface{}{
					"MY_SECRET_STRING": "<string>",
					"MY_SECRET_NUM":    "<num>",
				},
			},
			vaultData: map[string]interface{}{
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

func TestToYAML_ConfigMap(t *testing.T) {
	d := ConfigMapTemplate{
		Resource{
			templateData: map[string]interface{}{
				"metadata": map[string]interface{}{
					"namespace": "default",
					"name":      "<name>",
				},
				"data": map[string]interface{}{
					"MY_NONSECRET_STRING": "<string>",
					"MY_NONSECRET_NUM":    "<num>",
				},
			},
			vaultData: map[string]interface{}{
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
