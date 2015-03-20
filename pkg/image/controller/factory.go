package controller

import (
	"time"

	kapi "github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/client/cache"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/fields"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/runtime"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/util"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/watch"

	"github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/controller"
	"github.com/openshift/origin/pkg/dockerregistry"
	"github.com/openshift/origin/pkg/image/api"
)

// ImportControllerFactory can create an ImportController.
type ImportControllerFactory struct {
	Client client.Interface
}

// Create creates an ImportController.
func (f *ImportControllerFactory) Create() controller.RunnableController {
	lw := &cache.ListWatch{
		ListFunc: func() (runtime.Object, error) {
			return f.Client.ImageRepositories(kapi.NamespaceAll).List(labels.Everything(), fields.Everything())
		},
		WatchFunc: func(resourceVersion string) (watch.Interface, error) {
			return f.Client.ImageRepositories(kapi.NamespaceAll).Watch(labels.Everything(), fields.Everything(), resourceVersion)
		},
	}
	q := cache.NewFIFO(cache.MetaNamespaceKeyFunc)
	cache.NewReflector(lw, &api.ImageRepository{}, q, 2*time.Minute).Run()

	c := &ImportController{
		client:       dockerregistry.NewClient(),
		repositories: f.Client,
		mappings:     f.Client,
	}

	return &controller.RetryController{
		Queue: q,
		RetryManager: controller.NewQueueRetryManager(
			q,
			cache.MetaNamespaceKeyFunc,
			func(obj interface{}, err error, count int) bool {
				util.HandleError(err)
				return count < 5
			},
		),
		Handle: func(obj interface{}) error {
			r := obj.(*api.ImageRepository)
			return c.Next(r)
		},
	}
}
