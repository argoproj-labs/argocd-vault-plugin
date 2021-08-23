package kube

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"

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

// ReadSecret reads the specified Secret from the `argocd` namespace
// and returns a YAML []byte containing its data, decoded from base64
func (c *Client) ReadSecret(name string) ([]byte, error) {

	var namespace, secretname string

	// parse `namespace/secret` or get namespace from the service account.
	if strings.Contains(name, "/") {
		split := strings.Split(name, "/")
		namespace = split[0]
		secretname = split[1]
	} else {
		namespace_bytes, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
		if err != nil {
			return nil, fmt.Errorf("could not get namespace for serviceaccount: %s", err)
		}
		namespace = strings.TrimSpace(string(namespace_bytes))
		secretname = name
	}

	s, err := c.client.CoreV1().Secrets(namespace).Get(context.TODO(), secretname, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not get Secret %s/%s: %s", namespace, secretname, err)
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
