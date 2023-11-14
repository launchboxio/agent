package watcher

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	launchbox "github.com/launchboxio/launchbox-go-sdk/config"
	"github.com/launchboxio/launchbox-go-sdk/service/project"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"sync"
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

func (w *Watcher) Run() error {
	var wg sync.WaitGroup
	wg.Add(2)
	go w.WatchProjects(&wg)
	go w.WatchAddons(&wg)

	wg.Wait()
	return nil
}

func (w *Watcher) WatchProjects(wg *sync.WaitGroup) error {
	defer wg.Done()
	changes, err := w.Kube.Resource(schema.GroupVersionResource{
		Resource: "projects",
		Group:    "core.launchboxhq.io",
		Version:  "v1alpha1",
	}).Watch(context.Background(), metav1.ListOptions{})
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

		if event.Type == watch.Added || event.Type == watch.Modified {
			_, addonStatusesFound, err := unstructured.NestedMap(probject.UnstructuredContent(), "status", "addons")
			if err != nil {
				w.Logger.Error(err, "Failed getting addonstatus field")
				continue
			}
			if !addonStatusesFound {
				continue
			}

			projectAddons, projectAddonsFound, err := unstructured.NestedSlice(probject.UnstructuredContent(), "spec", "addons")
			if err != nil {
				w.Logger.Error(err, "Failed getting configured addons field")
				continue
			}
			if !projectAddonsFound {
				continue
			}
			fmt.Println(projectAddons)
			// TODO: For each addon, we want to get the subscription ID, and post
			// that data back to HQ to propagate further
		}
	}
	return nil
}

func (w *Watcher) WatchAddons(wg *sync.WaitGroup) error {
	defer wg.Done()
	changes, err := w.Kube.Resource(schema.GroupVersionResource{
		Resource: "addons",
		Group:    "core.launchboxhq.io",
		Version:  "v1alpha1",
	}).Watch(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	// Start watching the events, and log the change.
	// TODO: We'll somehow want to transmit any success / errors back to HQ
	// This is more complicated for addons, because while projects have a single
	// installation, an addon is currently deployed once to each cluster
	for event := range changes.ResultChan() {
		fmt.Println(event)
	}
	return nil
}
