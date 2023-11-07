package watcher

import (
	"context"
	"github.com/go-logr/logr"
	launchbox "github.com/launchboxio/launchbox-go-sdk/config"
	"github.com/launchboxio/launchbox-go-sdk/service/project"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
)

type Watcher struct {
	Sdk    *launchbox.Config
	Kube   *dynamic.DynamicClient
	Logger logr.Logger
}

const DefaultProjectStatus = "provisioning"

func New(clientset *dynamic.DynamicClient, sdk *launchbox.Config, logger logr.Logger) *Watcher {
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

	projectClient := project.New(w.Sdk)
	for event := range changes.ResultChan() {
		probject := event.Object.(*unstructured.Unstructured)

		projectId, projectIdFound, err := unstructured.NestedInt64(probject.UnstructuredContent(), "spec", "id")
		if err != nil {
			w.Logger.Error(err, "Failed getting project ID")
			continue
		}
		if !projectIdFound {
			w.Logger.Error(err, "Resource didn't have an ID...")
			continue
		}

		update := &project.UpdateProjectInput{
			ProjectId: int(projectId),
		}

		status, statusFound, err := unstructured.NestedString(probject.UnstructuredContent(), "status", "status")
		if err != nil {
			w.Logger.Error(err, "Failed getting status field")
			continue
		}
		if statusFound {
			update.Status = status
		} else {
			update.Status = DefaultProjectStatus
		}

		caCert, caCertFound, err := unstructured.NestedString(probject.UnstructuredContent(), "status", "caCertificate")
		if err != nil {
			w.Logger.Error(err, "Failed getting caCertificate field")
			continue
		}

		if caCertFound {
			update.CaCertificate = caCert
		}

		switch event.Type {
		case watch.Added:
			w.Logger.Info("Updating project", "ID", projectId)
			if _, err := projectClient.Update(update); err != nil {
				w.Logger.Error(err, "Failed updating project")
			}
		case watch.Modified:
			w.Logger.Info("Updating project", "ID", projectId)
			if _, err := projectClient.Update(update); err != nil {
				w.Logger.Error(err, "Failed updating project")
			}
		case watch.Deleted:
			w.Logger.Info("Project deleted, not sure what to do from here...")
		}
	}
	return nil
}
