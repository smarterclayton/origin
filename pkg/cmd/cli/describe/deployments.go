package describe

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/tabwriter"

	kapi "github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	kerrors "github.com/GoogleCloudPlatform/kubernetes/pkg/api/errors"
	kclient "github.com/GoogleCloudPlatform/kubernetes/pkg/client"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"

	"github.com/openshift/origin/pkg/client"
	deployapi "github.com/openshift/origin/pkg/deploy/api"
	deployutil "github.com/openshift/origin/pkg/deploy/util"
)

// DeploymentConfigDescriber generates information about a DeploymentConfig
type DeploymentConfigDescriber struct {
	client deploymentDescriberClient
}

type deploymentDescriberClient interface {
	getDeploymentConfig(namespace, name string) (*deployapi.DeploymentConfig, error)
	getDeployment(namespace, name string) (*kapi.ReplicationController, error)
	listPods(namespace string, selector labels.Selector) (*kapi.PodList, error)
}

type genericDeploymentDescriberClient struct {
	getDeploymentConfigFunc func(namespace, name string) (*deployapi.DeploymentConfig, error)
	getDeploymentFunc       func(namespace, name string) (*kapi.ReplicationController, error)
	listPodsFunc            func(namespace string, selector labels.Selector) (*kapi.PodList, error)
}

func (c *genericDeploymentDescriberClient) getDeploymentConfig(namespace, name string) (*deployapi.DeploymentConfig, error) {
	return c.getDeploymentConfigFunc(namespace, name)
}

func (c *genericDeploymentDescriberClient) getDeployment(namespace, name string) (*kapi.ReplicationController, error) {
	return c.getDeploymentFunc(namespace, name)
}

func (c *genericDeploymentDescriberClient) listPods(namespace string, selector labels.Selector) (*kapi.PodList, error) {
	return c.listPodsFunc(namespace, selector)
}

func NewDeploymentConfigDescriberForConfig(config *deployapi.DeploymentConfig) *DeploymentConfigDescriber {
	return &DeploymentConfigDescriber{
		client: &genericDeploymentDescriberClient{
			getDeploymentConfigFunc: func(namespace, name string) (*deployapi.DeploymentConfig, error) {
				return config, nil
			},
			getDeploymentFunc: func(namespace, name string) (*kapi.ReplicationController, error) {
				return nil, kerrors.NewNotFound("ReplicatonController", name)
			},
			listPodsFunc: func(namespace string, selector labels.Selector) (*kapi.PodList, error) {
				return nil, kerrors.NewNotFound("PodList", fmt.Sprintf("%v", selector))
			},
		},
	}
}

func NewDeploymentConfigDescriber(client client.Interface, kclient kclient.Interface) *DeploymentConfigDescriber {
	return &DeploymentConfigDescriber{
		client: &genericDeploymentDescriberClient{
			getDeploymentConfigFunc: func(namespace, name string) (*deployapi.DeploymentConfig, error) {
				return client.DeploymentConfigs(namespace).Get(name)
			},
			getDeploymentFunc: func(namespace, name string) (*kapi.ReplicationController, error) {
				return kclient.ReplicationControllers(namespace).Get(name)
			},
			listPodsFunc: func(namespace string, selector labels.Selector) (*kapi.PodList, error) {
				return kclient.Pods(namespace).List(selector)
			},
		},
	}
}

func (d *DeploymentConfigDescriber) Describe(namespace, name string) (string, error) {
	deploymentConfig, err := d.client.getDeploymentConfig(namespace, name)
	if err != nil {
		return "", err
	}

	return tabbedString(func(out *tabwriter.Writer) error {
		formatMeta(out, deploymentConfig.ObjectMeta)

		if deploymentConfig.LatestVersion == 0 {
			formatString(out, "Latest Version", "Not deployed")
		} else {
			formatString(out, "Latest Version", strconv.Itoa(deploymentConfig.LatestVersion))
		}

		printStrategy(deploymentConfig.Template.Strategy, out)
		printTriggers(deploymentConfig.Triggers, out)
		printReplicationControllerSpec(deploymentConfig.Template.ControllerTemplate, out)

		deploymentName := deployutil.LatestDeploymentNameForConfig(deploymentConfig)
		deployment, err := d.client.getDeployment(namespace, deploymentName)
		if err != nil {
			if kerrors.IsNotFound(err) {
				formatString(out, "Latest Deployment", "<none>")
			} else {
				formatString(out, "Latest Deployment", fmt.Sprintf("error: %v", err))
			}
		} else {
			printDeploymentRc(deployment, d.client, out)
		}

		return nil
	})
}

func printStrategy(strategy deployapi.DeploymentStrategy, w io.Writer) {
	fmt.Fprintf(w, "Strategy:\t%s\n", strategy.Type)
	switch strategy.Type {
	case deployapi.DeploymentStrategyTypeRecreate:
	case deployapi.DeploymentStrategyTypeCustom:
		fmt.Fprintf(w, "\t- Image:\t%s\n", strategy.CustomParams.Image)

		if len(strategy.CustomParams.Environment) > 0 {
			fmt.Fprintf(w, "\t- Environment:\t%s\n", formatLabels(convertEnv(strategy.CustomParams.Environment)))
		}

		if len(strategy.CustomParams.Command) > 0 {
			fmt.Fprintf(w, "\t- Command:\t%v\n", strings.Join(strategy.CustomParams.Command, " "))
		}
	}
}

func printTriggers(triggers []deployapi.DeploymentTriggerPolicy, w io.Writer) {
	if len(triggers) == 0 {
		fmt.Fprint(w, "Triggers:\t<none>\n")
		return
	}

	fmt.Fprint(w, "Triggers:\n")
	for _, t := range triggers {
		fmt.Fprintf(w, "\t- %s\n", t.Type)
		switch t.Type {
		case deployapi.DeploymentTriggerOnConfigChange:
			fmt.Fprintf(w, "\t\t<no options>\n")
		case deployapi.DeploymentTriggerOnImageChange:
			if len(t.ImageChangeParams.RepositoryName) > 0 {
				fmt.Fprintf(w, "\t\tAutomatic:\t%v\n\t\tRepository:\t%s\n\t\tTag:\t%s\n",
					t.ImageChangeParams.Automatic,
					t.ImageChangeParams.RepositoryName,
					t.ImageChangeParams.Tag,
				)
			} else if len(t.ImageChangeParams.From.Name) > 0 {
				fmt.Fprintf(w, "\t\tAutomatic:\t%v\n\t\tImage Repository:\t%s\n\t\tTag:\t%s\n",
					t.ImageChangeParams.Automatic,
					t.ImageChangeParams.From.Name,
					t.ImageChangeParams.Tag,
				)
			}
		default:
			fmt.Fprint(w, "unknown\n")
		}
	}
}

func printReplicationControllerSpec(spec kapi.ReplicationControllerSpec, w io.Writer) error {
	fmt.Fprint(w, "Template:\n")

	fmt.Fprintf(w, "\tSelector:\t%s\n\tReplicas:\t%d\n",
		formatLabels(spec.Selector),
		spec.Replicas)

	fmt.Fprintf(w, "\tContainers:\n\t\tNAME\tIMAGE\tENV\n")
	for _, container := range spec.Template.Spec.Containers {
		fmt.Fprintf(w, "\t\t%s\t%s\t%s\n",
			container.Name,
			container.Image,
			formatLabels(convertEnv(container.Env)))
	}
	return nil
}

func printDeploymentRc(deployment *kapi.ReplicationController, client deploymentDescriberClient, w io.Writer) error {
	running, waiting, succeeded, failed, err := getPodStatusForDeployment(deployment, client)
	if err != nil {
		return err
	}

	fmt.Fprint(w, "Latest Deployment:\n")
	fmt.Fprintf(w, "\tName:\t%s\n", deployment.Name)
	fmt.Fprintf(w, "\tStatus:\t%s\n", deployment.Annotations[deployapi.DeploymentStatusAnnotation])
	fmt.Fprintf(w, "\tSelector:\t%s\n", formatLabels(deployment.Spec.Selector))
	fmt.Fprintf(w, "\tLabels:\t%s\n", formatLabels(deployment.Labels))
	fmt.Fprintf(w, "\tReplicas:\t%d current / %d desired\n", deployment.Status.Replicas, deployment.Spec.Replicas)
	fmt.Fprintf(w, "\tPods Status:\t%d Running / %d Waiting / %d Succeeded / %d Failed\n", running, waiting, succeeded, failed)

	return nil
}

func getPodStatusForDeployment(deployment *kapi.ReplicationController, client deploymentDescriberClient) (running, waiting, succeeded, failed int, err error) {
	rcPods, err := client.listPods(deployment.Namespace, labels.SelectorFromSet(deployment.Spec.Selector))
	if err != nil {
		return
	}
	for _, pod := range rcPods.Items {
		switch pod.Status.Phase {
		case kapi.PodRunning:
			running++
		case kapi.PodPending:
			waiting++
		case kapi.PodSucceeded:
			succeeded++
		case kapi.PodFailed:
			failed++
		}
	}
	return
}

// DeploymentDescriber generates information about a deployment
// DEPRECATED.
type DeploymentDescriber struct {
	client.Interface
}

func (d *DeploymentDescriber) Describe(namespace, name string) (string, error) {
	c := d.Deployments(namespace)
	deployment, err := c.Get(name)
	if err != nil {
		return "", err
	}

	return tabbedString(func(out *tabwriter.Writer) error {
		formatMeta(out, deployment.ObjectMeta)
		formatString(out, "Status", bold(deployment.Status))
		formatString(out, "Strategy", deployment.Strategy.Type)
		causes := []string{}
		if deployment.Details != nil {
			for _, c := range deployment.Details.Causes {
				causes = append(causes, string(c.Type))
			}
		}
		formatString(out, "Causes", strings.Join(causes, ","))
		return nil
	})
}
