package watcher

import (
	"context"
	"fmt"
	"github.com/launchboxio/agent/pkg/client"
	"github.com/launchboxio/operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
)

type Watcher struct {
	Sdk  *client.Client
	Kube *dynamic.DynamicClient
}

func New(clientset *dynamic.DynamicClient, sdk *client.Client) *Watcher {
	return &Watcher{
		Sdk:  sdk,
		Kube: clientset,
	}
}

func (w *Watcher) Run(resource schema.GroupVersionResource) error {
	changes, err := w.Kube.Resource(resource).Watch(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for event := range changes.ResultChan() {
		svc := event.Object.(*v1alpha1.Project)

		switch event.Type {
		case watch.Added:
			fmt.Printf("Project %s/%s added", svc.ObjectMeta.Namespace, svc.ObjectMeta.Name)
		case watch.Modified:
			fmt.Printf("Project %s/%s modified", svc.ObjectMeta.Namespace, svc.ObjectMeta.Name)
		case watch.Deleted:
			fmt.Printf("Project %s/%s deleted", svc.ObjectMeta.Namespace, svc.ObjectMeta.Name)
		}
	}
	return nil
}
