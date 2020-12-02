package kube

import (
	"io/ioutil"
	"strings"
	"testing"
)

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

func TestToYAML_Secret(t *testing.T) {
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
