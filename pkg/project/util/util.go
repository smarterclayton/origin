package util

import (
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/v1"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/clientset/typed/core/v1"

	oapi "github.com/openshift/origin/pkg/api"
	api "github.com/openshift/origin/pkg/project/api"
	apiv1 "github.com/openshift/origin/pkg/project/api/v1"
)

// Associated returns true if the spec.finalizers contains the origin finalizer
func Associated(namespace *v1.Namespace) bool {
	for i := range namespace.Spec.Finalizers {
		if apiv1.FinalizerOrigin == namespace.Spec.Finalizers[i] {
			return true
		}
	}
	return false
}

// Associate adds the origin finalizer to spec.finalizers if its not there already
func Associate(kubeClient clientset.CoreV1Interface, namespace *v1.Namespace) (*v1.Namespace, error) {
	if Associated(namespace) {
		return namespace, nil
	}
	return finalizeInternal(kubeClient, namespace, true)
}

// Finalized returns true if the spec.finalizers does not contain the origin finalizer
func Finalized(namespace *v1.Namespace) bool {
	for i := range namespace.Spec.Finalizers {
		if apiv1.FinalizerOrigin == namespace.Spec.Finalizers[i] {
			return false
		}
	}
	return true
}

// Finalize will remove the origin finalizer from the namespace
func Finalize(kubeClient clientset.CoreV1Interface, namespace *v1.Namespace) (result *v1.Namespace, err error) {
	if Finalized(namespace) {
		return namespace, nil
	}

	// there is a potential for a resource conflict with base kubernetes finalizer
	// as a result, we handle resource conflicts in case multiple finalizers try
	// to finalize at same time
	for {
		result, err = finalizeInternal(kubeClient, namespace, false)
		if err == nil {
			return result, nil
		}

		if !kerrors.IsConflict(err) {
			return nil, err
		}

		namespace, err = kubeClient.Namespaces().Get(namespace.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
	}
}

// finalizeInternal will update the namespace finalizer list to either have or not have origin finalizer
func finalizeInternal(kubeClient clientset.CoreV1Interface, namespace *v1.Namespace, withOrigin bool) (*v1.Namespace, error) {
	namespaceFinalize := v1.Namespace{}
	namespaceFinalize.ObjectMeta = namespace.ObjectMeta
	namespaceFinalize.Spec = namespace.Spec

	finalizerSet := sets.NewString()
	for i := range namespace.Spec.Finalizers {
		finalizerSet.Insert(string(namespace.Spec.Finalizers[i]))
	}

	if withOrigin {
		finalizerSet.Insert(string(apiv1.FinalizerOrigin))
	} else {
		finalizerSet.Delete(string(apiv1.FinalizerOrigin))
	}

	namespaceFinalize.Spec.Finalizers = make([]v1.FinalizerName, 0, len(finalizerSet))
	for _, value := range finalizerSet.List() {
		namespaceFinalize.Spec.Finalizers = append(namespaceFinalize.Spec.Finalizers, v1.FinalizerName(value))
	}
	return kubeClient.Namespaces().Finalize(&namespaceFinalize)
}

// ConvertNamespace transforms a Namespace into a Project
func ConvertNamespaceV1(namespace *v1.Namespace) *api.Project {
	var finalizers []kapi.FinalizerName
	for _, s := range namespace.Spec.Finalizers {
		finalizers = append(finalizers, kapi.FinalizerName(s))
	}
	return &api.Project{
		ObjectMeta: namespace.ObjectMeta,
		Spec: api.ProjectSpec{
			Finalizers: finalizers,
		},
		Status: api.ProjectStatus{
			Phase: kapi.NamespacePhase(namespace.Status.Phase),
		},
	}
}

// ConvertNamespace transforms a Namespace into a Project
func ConvertNamespace(namespace *kapi.Namespace) *api.Project {
	return &api.Project{
		ObjectMeta: namespace.ObjectMeta,
		Spec: api.ProjectSpec{
			Finalizers: namespace.Spec.Finalizers,
		},
		Status: api.ProjectStatus{
			Phase: namespace.Status.Phase,
		},
	}
}

// ConvertProject transforms a Project into a Namespace
func ConvertProject(project *api.Project) *kapi.Namespace {
	namespace := &kapi.Namespace{
		ObjectMeta: project.ObjectMeta,
		Spec: kapi.NamespaceSpec{
			Finalizers: project.Spec.Finalizers,
		},
		Status: kapi.NamespaceStatus{
			Phase: project.Status.Phase,
		},
	}
	if namespace.Annotations == nil {
		namespace.Annotations = map[string]string{}
	}
	namespace.Annotations[oapi.OpenShiftDisplayName] = project.Annotations[oapi.OpenShiftDisplayName]
	return namespace
}

// ConvertProject transforms a Project into a Namespace
func ConvertProjectV1(project *api.Project) *v1.Namespace {
	var finalizers []v1.FinalizerName
	for _, s := range project.Spec.Finalizers {
		finalizers = append(finalizers, v1.FinalizerName(s))
	}
	namespace := &v1.Namespace{
		ObjectMeta: project.ObjectMeta,
		Spec: v1.NamespaceSpec{
			Finalizers: finalizers,
		},
		Status: v1.NamespaceStatus{
			Phase: v1.NamespacePhase(project.Status.Phase),
		},
	}
	if namespace.Annotations == nil {
		namespace.Annotations = map[string]string{}
	}
	namespace.Annotations[oapi.OpenShiftDisplayName] = project.Annotations[oapi.OpenShiftDisplayName]
	return namespace
}

// ConvertNamespaceListV1 transforms a NamespaceList into a ProjectList
func ConvertNamespaceListV1(namespaceList *v1.NamespaceList) *api.ProjectList {
	projects := &api.ProjectList{}
	for _, n := range namespaceList.Items {
		projects.Items = append(projects.Items, *ConvertNamespaceV1(&n))
	}
	return projects
}

// ConvertNamespaceList transforms a NamespaceList into a ProjectList
func ConvertNamespaceList(namespaceList *kapi.NamespaceList) *api.ProjectList {
	projects := &api.ProjectList{}
	for _, n := range namespaceList.Items {
		projects.Items = append(projects.Items, *ConvertNamespace(&n))
	}
	return projects
}
