package controller

import (
	"github.com/golang/glog"

	kapiv1 "k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/controller"
	sacontroller "k8s.io/kubernetes/pkg/controller/serviceaccount"
	"k8s.io/kubernetes/pkg/serviceaccount"

	"github.com/openshift/origin/pkg/cmd/server/bootstrappolicy"
	serviceaccountcontrollers "github.com/openshift/origin/pkg/serviceaccounts/controllers"
)

type ServiceAccountControllerOptions struct {
	ManagedNames []string
}

func (c *ServiceAccountControllerOptions) RunController(ctx ControllerContext) (bool, error) {
	if len(c.ManagedNames) == 0 {
		glog.Infof("Skipped starting Service Account Manager, no managed names specified")
		return false, nil
	}

	options := sacontroller.DefaultServiceAccountsControllerOptions()
	options.ServiceAccounts = []kapiv1.ServiceAccount{}

	for _, saName := range c.ManagedNames {
		sa := kapiv1.ServiceAccount{}
		sa.Name = saName

		options.ServiceAccounts = append(options.ServiceAccounts, sa)
	}

	go sacontroller.NewServiceAccountsController(
		ctx.DeprecatedOpenshiftInformers.KubernetesInformers().Core().V1().ServiceAccounts(),
		ctx.KubeControllerContext.InformerFactory.Core().V1().Namespaces(),
		ctx.ClientBuilder.ClientOrDie(bootstrappolicy.InfraServiceAccountControllerServiceAccountName),
		options).Run(3, ctx.Stop)

	return true, nil
}

type ServiceAccountTokensControllerOptions struct {
	RootCA           []byte
	ServiceServingCA []byte
	PrivateKey       interface{}

	RootClientBuilder controller.SimpleControllerClientBuilder
}

func (c *ServiceAccountTokensControllerOptions) RunController(ctx ControllerContext) (bool, error) {
	go sacontroller.NewTokensController(
		ctx.DeprecatedOpenshiftInformers.KubernetesInformers().Core().V1().ServiceAccounts(),
		ctx.DeprecatedOpenshiftInformers.KubernetesInformers().Core().V1().Secrets(),
		c.RootClientBuilder.ClientOrDie(bootstrappolicy.InfraServiceAccountTokensControllerServiceAccountName),
		sacontroller.TokensControllerOptions{
			TokenGenerator:   serviceaccount.JWTTokenGenerator(c.PrivateKey),
			RootCA:           c.RootCA,
			ServiceServingCA: c.ServiceServingCA,
		},
	).Run(int(ctx.KubeControllerContext.Options.ConcurrentSATokenSyncs), ctx.Stop)
	return true, nil
}

func RunServiceAccountPullSecretsController(ctx ControllerContext) (bool, error) {
	kc := ctx.ClientBuilder.KubeInternalClientOrDie(bootstrappolicy.InfraServiceAccountPullSecretsControllerServiceAccountName)

	go serviceaccountcontrollers.NewDockercfgDeletedController(
		ctx.DeprecatedOpenshiftInformers.InternalKubernetesInformers().Core().InternalVersion().Secrets(),
		kc,
		serviceaccountcontrollers.DockercfgDeletedControllerOptions{},
	).Run(ctx.Stop)

	go serviceaccountcontrollers.NewDockercfgTokenDeletedController(
		ctx.DeprecatedOpenshiftInformers.InternalKubernetesInformers().Core().InternalVersion().Secrets(),
		kc,
		serviceaccountcontrollers.DockercfgTokenDeletedControllerOptions{},
	).Run(ctx.Stop)

	dockerURLsInitialized := make(chan struct{})
	dockercfgController := serviceaccountcontrollers.NewDockercfgController(
		ctx.DeprecatedOpenshiftInformers.InternalKubernetesInformers().Core().InternalVersion().ServiceAccounts(),
		ctx.DeprecatedOpenshiftInformers.InternalKubernetesInformers().Core().InternalVersion().Secrets(),
		kc,
		serviceaccountcontrollers.DockercfgControllerOptions{DockerURLsInitialized: dockerURLsInitialized},
	)
	go dockercfgController.Run(5, ctx.Stop)

	dockerRegistryControllerOptions := serviceaccountcontrollers.DockerRegistryServiceControllerOptions{
		RegistryNamespace:     "default",
		RegistryServiceName:   "docker-registry",
		DockercfgController:   dockercfgController,
		DockerURLsInitialized: dockerURLsInitialized,
	}
	go serviceaccountcontrollers.NewDockerRegistryServiceController(
		ctx.DeprecatedOpenshiftInformers.InternalKubernetesInformers().Core().InternalVersion().Secrets(),
		kc,
		dockerRegistryControllerOptions,
	).Run(10, ctx.Stop)

	return true, nil
}
