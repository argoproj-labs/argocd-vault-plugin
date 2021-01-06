package client

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	k8yaml "sigs.k8s.io/yaml"
)

type Client struct {
	client *kubernetes.Clientset
}

func NewClient() (*Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &Client{
		client: clientset,
	}, nil
}

func (c *Client) ReadSecret(name string) ([]byte, error) {
	s, err := c.client.CoreV1().Secrets("argocd").Get(name, metav1.GetOptions{})
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
