package kube

import (
	"context"
	"fmt"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	k8yaml "sigs.k8s.io/yaml"
)

// Client contains the Kubernetes client for connecting
// to the cluster hosting ArgoCD
type Client struct {
	client *kubernetes.Clientset
}

// NewClient returns a Client ready to call the local Kubernetes API
func NewClient() (*Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("could not connect to local Kubernetes cluster to read Secret: %s", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &Client{
		client: clientset,
	}, nil
}

// ReadSecretData reads the specified Secret from the defined namespace, otherwise defaults to `argocd`
// and returns map[string][]byte containing its data
func (c *Client) ReadSecretData(name string) (map[string][]byte, error) {
	secretNamespace, secretName := secretNamespaceName(name)

	utils.VerboseToStdErr("parsed secret name as %s from namespace %s", secretName, secretNamespace)

	s, err := c.client.CoreV1().Secrets(secretNamespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return s.Data, nil
}

// ReadSecret reads the specified Secret from the defined namespace, otherwise defaults to `argocd`
// and returns a YAML []byte containing its data, decoded from base64
func (c *Client) ReadSecret(name string) ([]byte, error) {
	data, err := c.ReadSecretData(name)
	if err != nil {
		return nil, err
	}

	decoded := make(map[string]string)
	for key, value := range data {
		decoded[key] = string(value)
	}
	res, err := k8yaml.Marshal(&decoded)
	if err != nil {
		return nil, err
	}
	return res, nil
}
