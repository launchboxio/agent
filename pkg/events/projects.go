package events

import (
  "context"
  "errors"
  "github.com/go-logr/logr"
  launchbox "github.com/launchboxio/launchbox-go-sdk/config"
  "github.com/launchboxio/launchbox-go-sdk/service/project"
  "github.com/launchboxio/operator/api/v1alpha1"
  apierrors "k8s.io/apimachinery/pkg/api/errors"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  "k8s.io/apimachinery/pkg/types"
  "sigs.k8s.io/controller-runtime/pkg/client"
)

type ProjectEventPayload struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Slug  string `json:"slug"`
	Users []struct {
		Email       string `json:"email"`
		ClusterRole string `json:"clusterRole"`
	} `json:"users"`
	Cpu    int `json:"cpu"`
	Memory int `json:"memory"`
	Disk   int `json:"disk"`
}

type ProjectPayload struct {
}

type ProjectHandler struct {
	Logger logr.Logger
	Client client.Client
	Sdk    *launchbox.Config
}

func (ph *ProjectHandler) syncProjectResource(event *LaunchboxEvent) error {
	resource, err := ph.projectFromPayload(event)
	if err != nil {
		return err
	}

	projectCr := &v1alpha1.Project{}
	if err := ph.Client.Get(context.TODO(), types.NamespacedName{
		Name:      resource.ObjectMeta.Name,
		Namespace: resource.ObjectMeta.Namespace,
	}, projectCr); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		ph.Logger.Info("Creating new project resource")
		return ph.Client.Create(context.TODO(), resource)
	}
	ph.Logger.Info("Updating existing project resource")
	projectCr.Spec = resource.Spec
	return ph.Client.Update(context.TODO(), projectCr)
}

func (ph *ProjectHandler) Create(event *LaunchboxEvent) error {
	return ph.syncProjectResource(event)
}

func (ph *ProjectHandler) Update(event *LaunchboxEvent) error {
	return ph.syncProjectResource(event)
}

func (ph *ProjectHandler) Delete(event *LaunchboxEvent) error {
	resource, err := ph.projectFromPayload(event)
	if err != nil {
		return err
	}
	project := &v1alpha1.Project{}
	if err := ph.Client.Get(context.TODO(), client.ObjectKey{
		Name:      resource.ObjectMeta.Name,
		Namespace: resource.ObjectMeta.Namespace,
	}, project); err != nil {
		return err
	}

	return ph.Client.Delete(context.TODO(), project)
}

func (ph *ProjectHandler) Pause(event *LaunchboxEvent) error {
	resource, err := ph.projectFromPayload(event)
	if err != nil {
		return err
	}
	project := &v1alpha1.Project{}
	if err := ph.Client.Get(context.TODO(), client.ObjectKey{
		Name:      resource.ObjectMeta.Name,
		Namespace: resource.ObjectMeta.Namespace,
	}, project); err != nil {
		return err
	}
	project.Spec.Paused = true

	return ph.Client.Update(context.TODO(), project)
}

func (ph *ProjectHandler) Resume(event *LaunchboxEvent) error {
	resource, err := ph.projectFromPayload(event)
	if err != nil {
		return err
	}
	project := &v1alpha1.Project{}
	if err := ph.Client.Get(context.TODO(), client.ObjectKey{
		Name:      resource.ObjectMeta.Name,
		Namespace: resource.ObjectMeta.Namespace,
	}, project); err != nil {
		return err
	}
	project.Spec.Paused = false
	return ph.Client.Update(context.TODO(), project)
}

// TODO: We should instead just query the projects directly using the SDK
func (ph *ProjectHandler) projectFromPayload(event *LaunchboxEvent) (*v1alpha1.Project, error) {
	if _, ok := event.Payload["id"]; !ok {
		return nil, errors.New("invalid payload: no ID field found")
	}
	projectId, ok := event.Payload["id"].(float64)
	if !ok {
		return nil, errors.New("invalid payload: unable to cast ID")
	}

	projectSdk := project.New(ph.Sdk)
	output, err := projectSdk.GetManifest(&project.GetProjectManifestInput{
		ProjectId: int(projectId),
	})
	if err != nil {
		return nil, err
	}
	users := []v1alpha1.ProjectUser{
		{Email: output.Manifest.User.Email, ClusterRole: "cluster-admin"},
	}

	projectCr := &v1alpha1.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name:      output.Manifest.Slug,
			Namespace: "lbx-system",
		},
		Spec: v1alpha1.ProjectSpec{
			Slug: output.Manifest.Slug,
			Id:   output.Manifest.Id,
			// TODO: Pull this from the event payload
			Crossplane: v1alpha1.ProjectCrossplaneSpec{
				Providers: []string{},
			},
			Resources: v1alpha1.Resources{
				Cpu:    int32(output.Manifest.Cpu),
				Memory: int32(output.Manifest.Memory),
				Disk:   int32(output.Manifest.Disk),
			},
			KubernetesVersion: output.Manifest.KubernetesVersion,
			Users:             users,
		},
	}

	if projectCr.Spec.Addons == nil {
		projectCr.Spec.Addons = []v1alpha1.ProjectAddonSpec{}
	}

	for _, addonSub := range output.Manifest.AddonSubscriptions {
		projectCr.Spec.Addons = append(projectCr.Spec.Addons, v1alpha1.ProjectAddonSpec{
			AddonName:        addonSub.Addon.Name,
			InstallationName: addonSub.Name,
			Resource:         addonSub.Addon.DefaultVersion.ClaimName,
			Group:            addonSub.Addon.DefaultVersion.Group,
			Version:          addonSub.Addon.DefaultVersion.Version,
		})
	}

	return projectCr, nil
}
