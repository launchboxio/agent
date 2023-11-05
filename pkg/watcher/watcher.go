package watcher

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/launchboxio/agent/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
)

type Watcher struct {
	Sdk    *client.Client
	Kube   *dynamic.DynamicClient
	Logger logr.Logger
}

const DefaultProjectStatus = "provisioning"

func New(clientset *dynamic.DynamicClient, sdk *client.Client, logger logr.Logger) *Watcher {
	return &Watcher{
		Sdk:    sdk,
		Kube:   clientset,
		Logger: logger,
	}
}

func (w *Watcher) Run(resource schema.GroupVersionResource) error {
	changes, err := w.Kube.Resource(resource).Watch(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	projectClient := w.Sdk.Projects()
	for event := range changes.ResultChan() {
		project := event.Object.(*unstructured.Unstructured)

		update := &client.ProjectUpdatePayload{}
		projectId, projectIdFound, err := unstructured.NestedInt64(project.UnstructuredContent(), "spec", "id")
		fmt.Println(projectId)
		if err != nil {
			w.Logger.Error(err, "Failed getting project ID")
			continue
		}
		if !projectIdFound {
			w.Logger.Error(err, "Resource didn't have an ID...")
			continue
		}
		status, statusFound, err := unstructured.NestedString(project.UnstructuredContent(), "status", "status")
		fmt.Println(status)
		if err != nil {
			w.Logger.Error(err, "Failed getting status field")
			continue
		}
		if statusFound {
			update.Status = status
		} else {
			update.Status = DefaultProjectStatus
		}

		caCert, caCertFound, err := unstructured.NestedString(project.UnstructuredContent(), "status", "caCertificate")
		fmt.Println(caCert)
		if err != nil {
			w.Logger.Error(err, "Failed getting caCertificate field")
			continue
		}

		if caCertFound {
			update.CaCertificate = caCert
		}

		request := &client.ProjectUpdateRequest{Project: update}
		switch event.Type {
		case watch.Added:
			w.Logger.Info("Updating project", "ID", projectId)
			if _, err := projectClient.Update(int(projectId), request); err != nil {
				w.Logger.Error(err, "Failed updating project")
			}
		case watch.Modified:
			w.Logger.Info("Updating project", "ID", projectId)
			if _, err := projectClient.Update(int(projectId), request); err != nil {
				w.Logger.Error(err, "Failed updating project")
			}
		case watch.Deleted:
			w.Logger.Info("Project deleted, not sure what to do from here...")
		}
	}
	return nil
}
