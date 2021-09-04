package kube

import (
	"context"
	"fmt"

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

// ReadSecret reads the specified Secret from the specified namespace
// and returns a YAML []byte containing its data, decoded from base64
func (c *Client) ReadSecret(name string, namespace string) ([]byte, error) {
	s, err := c.client.CoreV1().Secrets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	decoded := make(map[string]string)
	for key, value := range s.Data {
		decoded[key] = string(value)
	}
	res, err := k8yaml.Marshal(&decoded)
	if err != nil {
		return nil, err
	}
	return res, nil
}
