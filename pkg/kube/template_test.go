package kube

import (
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
	return map[string]interface{}{}, nil
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
}
