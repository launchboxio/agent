package evaluator

import (
	"context"
	"errors"
	"fmt"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"os"
)

type Evaluator struct {
	KubeClient *kubernetes.Clientset
}

type Evaluation struct {
	KubernetesVersion string
	Region            string
	Provider          string
	AgentIdentifier   string
}

func New(client *kubernetes.Clientset) *Evaluator {
	return &Evaluator{
		KubeClient: client,
	}
}

func (e *Evaluator) Evaluate() (*Evaluation, error) {
	result := &Evaluation{
		Provider:          "launchbox",
		Region:            "",
		KubernetesVersion: "",
		AgentIdentifier:   fmt.Sprintf("%s/%s", os.Getenv("POD_NAME"), os.Getenv("POD_NAMESPACE")),
	}
	version, err := e.KubeClient.ServerVersion()
	if err != nil {
		return nil, err
	}
	result.KubernetesVersion = version.String()

	nodes, err := e.KubeClient.CoreV1().Nodes().List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	if len(nodes.Items) == 0 {
		return nil, errors.New("Evaluation failed: No nodes running for cluster")
	}

	node := nodes.Items[0]

	for label, value := range node.Labels {
		if label == "k8s.io/cloud-provider-aws" {
			result.Provider = "aws"
		}

		if label == "topology.kubernetes.io/region" {
			result.Region = value
		}
	}

	return result, nil
}
