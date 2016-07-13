// +build !ignore_autogenerated_openshift

// This file was autogenerated by deepcopy-gen. Do not edit it manually!

package api

import (
	api "k8s.io/kubernetes/pkg/api"
	unversioned "k8s.io/kubernetes/pkg/api/unversioned"
	conversion "k8s.io/kubernetes/pkg/conversion"
	runtime "k8s.io/kubernetes/pkg/runtime"
	sets "k8s.io/kubernetes/pkg/util/sets"
)

func init() {
	if err := api.Scheme.AddGeneratedDeepCopyFuncs(
		DeepCopy_api_Action,
		DeepCopy_api_ClusterPolicy,
		DeepCopy_api_ClusterPolicyBinding,
		DeepCopy_api_ClusterPolicyBindingList,
		DeepCopy_api_ClusterPolicyList,
		DeepCopy_api_ClusterRole,
		DeepCopy_api_ClusterRoleBinding,
		DeepCopy_api_ClusterRoleBindingList,
		DeepCopy_api_ClusterRoleList,
		DeepCopy_api_IsPersonalSubjectAccessReview,
		DeepCopy_api_LocalResourceAccessReview,
		DeepCopy_api_LocalSubjectAccessReview,
		DeepCopy_api_Policy,
		DeepCopy_api_PolicyBinding,
		DeepCopy_api_PolicyBindingList,
		DeepCopy_api_PolicyList,
		DeepCopy_api_PolicyRule,
		DeepCopy_api_ResourceAccessReview,
		DeepCopy_api_ResourceAccessReviewResponse,
		DeepCopy_api_Role,
		DeepCopy_api_RoleBinding,
		DeepCopy_api_RoleBindingList,
		DeepCopy_api_RoleList,
		DeepCopy_api_SelfSubjectRulesReview,
		DeepCopy_api_SelfSubjectRulesReviewSpec,
		DeepCopy_api_SubjectAccessReview,
		DeepCopy_api_SubjectAccessReviewResponse,
		DeepCopy_api_SubjectRulesReviewStatus,
	); err != nil {
		// if one of the deep copy functions is malformed, detect it immediately.
		panic(err)
	}
}

func DeepCopy_api_Action(in Action, out *Action, c *conversion.Cloner) error {
	out.Namespace = in.Namespace
	out.Verb = in.Verb
	out.Group = in.Group
	out.Version = in.Version
	out.Resource = in.Resource
	out.ResourceName = in.ResourceName
	if in.Content == nil {
		out.Content = nil
	} else if newVal, err := c.DeepCopy(in.Content); err != nil {
		return err
	} else {
		out.Content = newVal.(runtime.Object)
	}
	return nil
}

func DeepCopy_api_ClusterPolicy(in ClusterPolicy, out *ClusterPolicy, c *conversion.Cloner) error {
	if err := unversioned.DeepCopy_unversioned_TypeMeta(in.TypeMeta, &out.TypeMeta, c); err != nil {
		return err
	}
	if err := api.DeepCopy_api_ObjectMeta(in.ObjectMeta, &out.ObjectMeta, c); err != nil {
		return err
	}
	if err := unversioned.DeepCopy_unversioned_Time(in.LastModified, &out.LastModified, c); err != nil {
		return err
	}
	if in.Roles != nil {
		in, out := in.Roles, &out.Roles
		*out = make(map[string]*ClusterRole)
		for key, val := range in {
			if newVal, err := c.DeepCopy(val); err != nil {
				return err
			} else {
				(*out)[key] = newVal.(*ClusterRole)
			}
		}
	} else {
		out.Roles = nil
	}
	return nil
}

func DeepCopy_api_ClusterPolicyBinding(in ClusterPolicyBinding, out *ClusterPolicyBinding, c *conversion.Cloner) error {
	if err := unversioned.DeepCopy_unversioned_TypeMeta(in.TypeMeta, &out.TypeMeta, c); err != nil {
		return err
	}
	if err := api.DeepCopy_api_ObjectMeta(in.ObjectMeta, &out.ObjectMeta, c); err != nil {
		return err
	}
	if err := unversioned.DeepCopy_unversioned_Time(in.LastModified, &out.LastModified, c); err != nil {
		return err
	}
	if err := api.DeepCopy_api_ObjectReference(in.PolicyRef, &out.PolicyRef, c); err != nil {
		return err
	}
	if in.RoleBindings != nil {
		in, out := in.RoleBindings, &out.RoleBindings
		*out = make(map[string]*ClusterRoleBinding)
		for key, val := range in {
			if newVal, err := c.DeepCopy(val); err != nil {
				return err
			} else {
				(*out)[key] = newVal.(*ClusterRoleBinding)
			}
		}
	} else {
		out.RoleBindings = nil
	}
	return nil
}

func DeepCopy_api_ClusterPolicyBindingList(in ClusterPolicyBindingList, out *ClusterPolicyBindingList, c *conversion.Cloner) error {
	if err := unversioned.DeepCopy_unversioned_TypeMeta(in.TypeMeta, &out.TypeMeta, c); err != nil {
		return err
	}
	if err := unversioned.DeepCopy_unversioned_ListMeta(in.ListMeta, &out.ListMeta, c); err != nil {
		return err
	}
	if in.Items != nil {
		in, out := in.Items, &out.Items
		*out = make([]ClusterPolicyBinding, len(in))
		for i := range in {
			if err := DeepCopy_api_ClusterPolicyBinding(in[i], &(*out)[i], c); err != nil {
				return err
			}
		}
	} else {
		out.Items = nil
	}
	return nil
}

func DeepCopy_api_ClusterPolicyList(in ClusterPolicyList, out *ClusterPolicyList, c *conversion.Cloner) error {
	if err := unversioned.DeepCopy_unversioned_TypeMeta(in.TypeMeta, &out.TypeMeta, c); err != nil {
		return err
	}
	if err := unversioned.DeepCopy_unversioned_ListMeta(in.ListMeta, &out.ListMeta, c); err != nil {
		return err
	}
	if in.Items != nil {
		in, out := in.Items, &out.Items
		*out = make([]ClusterPolicy, len(in))
		for i := range in {
			if err := DeepCopy_api_ClusterPolicy(in[i], &(*out)[i], c); err != nil {
				return err
			}
		}
	} else {
		out.Items = nil
	}
	return nil
}

func DeepCopy_api_ClusterRole(in ClusterRole, out *ClusterRole, c *conversion.Cloner) error {
	if err := unversioned.DeepCopy_unversioned_TypeMeta(in.TypeMeta, &out.TypeMeta, c); err != nil {
		return err
	}
	if err := api.DeepCopy_api_ObjectMeta(in.ObjectMeta, &out.ObjectMeta, c); err != nil {
		return err
	}
	if in.Rules != nil {
		in, out := in.Rules, &out.Rules
		*out = make([]PolicyRule, len(in))
		for i := range in {
			if err := DeepCopy_api_PolicyRule(in[i], &(*out)[i], c); err != nil {
				return err
			}
		}
	} else {
		out.Rules = nil
	}
	return nil
}

func DeepCopy_api_ClusterRoleBinding(in ClusterRoleBinding, out *ClusterRoleBinding, c *conversion.Cloner) error {
	if err := unversioned.DeepCopy_unversioned_TypeMeta(in.TypeMeta, &out.TypeMeta, c); err != nil {
		return err
	}
	if err := api.DeepCopy_api_ObjectMeta(in.ObjectMeta, &out.ObjectMeta, c); err != nil {
		return err
	}
	if in.Subjects != nil {
		in, out := in.Subjects, &out.Subjects
		*out = make([]api.ObjectReference, len(in))
		for i := range in {
			if err := api.DeepCopy_api_ObjectReference(in[i], &(*out)[i], c); err != nil {
				return err
			}
		}
	} else {
		out.Subjects = nil
	}
	if err := api.DeepCopy_api_ObjectReference(in.RoleRef, &out.RoleRef, c); err != nil {
		return err
	}
	return nil
}

func DeepCopy_api_ClusterRoleBindingList(in ClusterRoleBindingList, out *ClusterRoleBindingList, c *conversion.Cloner) error {
	if err := unversioned.DeepCopy_unversioned_TypeMeta(in.TypeMeta, &out.TypeMeta, c); err != nil {
		return err
	}
	if err := unversioned.DeepCopy_unversioned_ListMeta(in.ListMeta, &out.ListMeta, c); err != nil {
		return err
	}
	if in.Items != nil {
		in, out := in.Items, &out.Items
		*out = make([]ClusterRoleBinding, len(in))
		for i := range in {
			if err := DeepCopy_api_ClusterRoleBinding(in[i], &(*out)[i], c); err != nil {
				return err
			}
		}
	} else {
		out.Items = nil
	}
	return nil
}

func DeepCopy_api_ClusterRoleList(in ClusterRoleList, out *ClusterRoleList, c *conversion.Cloner) error {
	if err := unversioned.DeepCopy_unversioned_TypeMeta(in.TypeMeta, &out.TypeMeta, c); err != nil {
		return err
	}
	if err := unversioned.DeepCopy_unversioned_ListMeta(in.ListMeta, &out.ListMeta, c); err != nil {
		return err
	}
	if in.Items != nil {
		in, out := in.Items, &out.Items
		*out = make([]ClusterRole, len(in))
		for i := range in {
			if err := DeepCopy_api_ClusterRole(in[i], &(*out)[i], c); err != nil {
				return err
			}
		}
	} else {
		out.Items = nil
	}
	return nil
}

func DeepCopy_api_IsPersonalSubjectAccessReview(in IsPersonalSubjectAccessReview, out *IsPersonalSubjectAccessReview, c *conversion.Cloner) error {
	if err := unversioned.DeepCopy_unversioned_TypeMeta(in.TypeMeta, &out.TypeMeta, c); err != nil {
		return err
	}
	return nil
}

func DeepCopy_api_LocalResourceAccessReview(in LocalResourceAccessReview, out *LocalResourceAccessReview, c *conversion.Cloner) error {
	if err := unversioned.DeepCopy_unversioned_TypeMeta(in.TypeMeta, &out.TypeMeta, c); err != nil {
		return err
	}
	if err := DeepCopy_api_Action(in.Action, &out.Action, c); err != nil {
		return err
	}
	return nil
}

func DeepCopy_api_LocalSubjectAccessReview(in LocalSubjectAccessReview, out *LocalSubjectAccessReview, c *conversion.Cloner) error {
	if err := unversioned.DeepCopy_unversioned_TypeMeta(in.TypeMeta, &out.TypeMeta, c); err != nil {
		return err
	}
	if err := DeepCopy_api_Action(in.Action, &out.Action, c); err != nil {
		return err
	}
	out.User = in.User
	if in.Groups != nil {
		in, out := in.Groups, &out.Groups
		*out = make(sets.String)
		for key, val := range in {
			if newVal, err := c.DeepCopy(val); err != nil {
				return err
			} else {
				(*out)[key] = newVal.(sets.Empty)
			}
		}
	} else {
		out.Groups = nil
	}
	if in.Scopes != nil {
		in, out := in.Scopes, &out.Scopes
		*out = make([]string, len(in))
		copy(*out, in)
	} else {
		out.Scopes = nil
	}
	return nil
}

func DeepCopy_api_Policy(in Policy, out *Policy, c *conversion.Cloner) error {
	if err := unversioned.DeepCopy_unversioned_TypeMeta(in.TypeMeta, &out.TypeMeta, c); err != nil {
		return err
	}
	if err := api.DeepCopy_api_ObjectMeta(in.ObjectMeta, &out.ObjectMeta, c); err != nil {
		return err
	}
	if err := unversioned.DeepCopy_unversioned_Time(in.LastModified, &out.LastModified, c); err != nil {
		return err
	}
	if in.Roles != nil {
		in, out := in.Roles, &out.Roles
		*out = make(map[string]*Role)
		for key, val := range in {
			if newVal, err := c.DeepCopy(val); err != nil {
				return err
			} else {
				(*out)[key] = newVal.(*Role)
			}
		}
	} else {
		out.Roles = nil
	}
	return nil
}

func DeepCopy_api_PolicyBinding(in PolicyBinding, out *PolicyBinding, c *conversion.Cloner) error {
	if err := unversioned.DeepCopy_unversioned_TypeMeta(in.TypeMeta, &out.TypeMeta, c); err != nil {
		return err
	}
	if err := api.DeepCopy_api_ObjectMeta(in.ObjectMeta, &out.ObjectMeta, c); err != nil {
		return err
	}
	if err := unversioned.DeepCopy_unversioned_Time(in.LastModified, &out.LastModified, c); err != nil {
		return err
	}
	if err := api.DeepCopy_api_ObjectReference(in.PolicyRef, &out.PolicyRef, c); err != nil {
		return err
	}
	if in.RoleBindings != nil {
		in, out := in.RoleBindings, &out.RoleBindings
		*out = make(map[string]*RoleBinding)
		for key, val := range in {
			if newVal, err := c.DeepCopy(val); err != nil {
				return err
			} else {
				(*out)[key] = newVal.(*RoleBinding)
			}
		}
	} else {
		out.RoleBindings = nil
	}
	return nil
}

func DeepCopy_api_PolicyBindingList(in PolicyBindingList, out *PolicyBindingList, c *conversion.Cloner) error {
	if err := unversioned.DeepCopy_unversioned_TypeMeta(in.TypeMeta, &out.TypeMeta, c); err != nil {
		return err
	}
	if err := unversioned.DeepCopy_unversioned_ListMeta(in.ListMeta, &out.ListMeta, c); err != nil {
		return err
	}
	if in.Items != nil {
		in, out := in.Items, &out.Items
		*out = make([]PolicyBinding, len(in))
		for i := range in {
			if err := DeepCopy_api_PolicyBinding(in[i], &(*out)[i], c); err != nil {
				return err
			}
		}
	} else {
		out.Items = nil
	}
	return nil
}

func DeepCopy_api_PolicyList(in PolicyList, out *PolicyList, c *conversion.Cloner) error {
	if err := unversioned.DeepCopy_unversioned_TypeMeta(in.TypeMeta, &out.TypeMeta, c); err != nil {
		return err
	}
	if err := unversioned.DeepCopy_unversioned_ListMeta(in.ListMeta, &out.ListMeta, c); err != nil {
		return err
	}
	if in.Items != nil {
		in, out := in.Items, &out.Items
		*out = make([]Policy, len(in))
		for i := range in {
			if err := DeepCopy_api_Policy(in[i], &(*out)[i], c); err != nil {
				return err
			}
		}
	} else {
		out.Items = nil
	}
	return nil
}

func DeepCopy_api_PolicyRule(in PolicyRule, out *PolicyRule, c *conversion.Cloner) error {
	if in.Verbs != nil {
		in, out := in.Verbs, &out.Verbs
		*out = make(sets.String)
		for key, val := range in {
			if newVal, err := c.DeepCopy(val); err != nil {
				return err
			} else {
				(*out)[key] = newVal.(sets.Empty)
			}
		}
	} else {
		out.Verbs = nil
	}
	if in.AttributeRestrictions == nil {
		out.AttributeRestrictions = nil
	} else if newVal, err := c.DeepCopy(in.AttributeRestrictions); err != nil {
		return err
	} else {
		out.AttributeRestrictions = newVal.(runtime.Object)
	}
	if in.APIGroups != nil {
		in, out := in.APIGroups, &out.APIGroups
		*out = make([]string, len(in))
		copy(*out, in)
	} else {
		out.APIGroups = nil
	}
	if in.Resources != nil {
		in, out := in.Resources, &out.Resources
		*out = make(sets.String)
		for key, val := range in {
			if newVal, err := c.DeepCopy(val); err != nil {
				return err
			} else {
				(*out)[key] = newVal.(sets.Empty)
			}
		}
	} else {
		out.Resources = nil
	}
	if in.ResourceNames != nil {
		in, out := in.ResourceNames, &out.ResourceNames
		*out = make(sets.String)
		for key, val := range in {
			if newVal, err := c.DeepCopy(val); err != nil {
				return err
			} else {
				(*out)[key] = newVal.(sets.Empty)
			}
		}
	} else {
		out.ResourceNames = nil
	}
	if in.NonResourceURLs != nil {
		in, out := in.NonResourceURLs, &out.NonResourceURLs
		*out = make(sets.String)
		for key, val := range in {
			if newVal, err := c.DeepCopy(val); err != nil {
				return err
			} else {
				(*out)[key] = newVal.(sets.Empty)
			}
		}
	} else {
		out.NonResourceURLs = nil
	}
	return nil
}

func DeepCopy_api_ResourceAccessReview(in ResourceAccessReview, out *ResourceAccessReview, c *conversion.Cloner) error {
	if err := unversioned.DeepCopy_unversioned_TypeMeta(in.TypeMeta, &out.TypeMeta, c); err != nil {
		return err
	}
	if err := DeepCopy_api_Action(in.Action, &out.Action, c); err != nil {
		return err
	}
	return nil
}

func DeepCopy_api_ResourceAccessReviewResponse(in ResourceAccessReviewResponse, out *ResourceAccessReviewResponse, c *conversion.Cloner) error {
	if err := unversioned.DeepCopy_unversioned_TypeMeta(in.TypeMeta, &out.TypeMeta, c); err != nil {
		return err
	}
	out.Namespace = in.Namespace
	if in.Users != nil {
		in, out := in.Users, &out.Users
		*out = make(sets.String)
		for key, val := range in {
			if newVal, err := c.DeepCopy(val); err != nil {
				return err
			} else {
				(*out)[key] = newVal.(sets.Empty)
			}
		}
	} else {
		out.Users = nil
	}
	if in.Groups != nil {
		in, out := in.Groups, &out.Groups
		*out = make(sets.String)
		for key, val := range in {
			if newVal, err := c.DeepCopy(val); err != nil {
				return err
			} else {
				(*out)[key] = newVal.(sets.Empty)
			}
		}
	} else {
		out.Groups = nil
	}
	out.EvaluationError = in.EvaluationError
	return nil
}

func DeepCopy_api_Role(in Role, out *Role, c *conversion.Cloner) error {
	if err := unversioned.DeepCopy_unversioned_TypeMeta(in.TypeMeta, &out.TypeMeta, c); err != nil {
		return err
	}
	if err := api.DeepCopy_api_ObjectMeta(in.ObjectMeta, &out.ObjectMeta, c); err != nil {
		return err
	}
	if in.Rules != nil {
		in, out := in.Rules, &out.Rules
		*out = make([]PolicyRule, len(in))
		for i := range in {
			if err := DeepCopy_api_PolicyRule(in[i], &(*out)[i], c); err != nil {
				return err
			}
		}
	} else {
		out.Rules = nil
	}
	return nil
}

func DeepCopy_api_RoleBinding(in RoleBinding, out *RoleBinding, c *conversion.Cloner) error {
	if err := unversioned.DeepCopy_unversioned_TypeMeta(in.TypeMeta, &out.TypeMeta, c); err != nil {
		return err
	}
	if err := api.DeepCopy_api_ObjectMeta(in.ObjectMeta, &out.ObjectMeta, c); err != nil {
		return err
	}
	if in.Subjects != nil {
		in, out := in.Subjects, &out.Subjects
		*out = make([]api.ObjectReference, len(in))
		for i := range in {
			if err := api.DeepCopy_api_ObjectReference(in[i], &(*out)[i], c); err != nil {
				return err
			}
		}
	} else {
		out.Subjects = nil
	}
	if err := api.DeepCopy_api_ObjectReference(in.RoleRef, &out.RoleRef, c); err != nil {
		return err
	}
	return nil
}

func DeepCopy_api_RoleBindingList(in RoleBindingList, out *RoleBindingList, c *conversion.Cloner) error {
	if err := unversioned.DeepCopy_unversioned_TypeMeta(in.TypeMeta, &out.TypeMeta, c); err != nil {
		return err
	}
	if err := unversioned.DeepCopy_unversioned_ListMeta(in.ListMeta, &out.ListMeta, c); err != nil {
		return err
	}
	if in.Items != nil {
		in, out := in.Items, &out.Items
		*out = make([]RoleBinding, len(in))
		for i := range in {
			if err := DeepCopy_api_RoleBinding(in[i], &(*out)[i], c); err != nil {
				return err
			}
		}
	} else {
		out.Items = nil
	}
	return nil
}

func DeepCopy_api_RoleList(in RoleList, out *RoleList, c *conversion.Cloner) error {
	if err := unversioned.DeepCopy_unversioned_TypeMeta(in.TypeMeta, &out.TypeMeta, c); err != nil {
		return err
	}
	if err := unversioned.DeepCopy_unversioned_ListMeta(in.ListMeta, &out.ListMeta, c); err != nil {
		return err
	}
	if in.Items != nil {
		in, out := in.Items, &out.Items
		*out = make([]Role, len(in))
		for i := range in {
			if err := DeepCopy_api_Role(in[i], &(*out)[i], c); err != nil {
				return err
			}
		}
	} else {
		out.Items = nil
	}
	return nil
}

func DeepCopy_api_SelfSubjectRulesReview(in SelfSubjectRulesReview, out *SelfSubjectRulesReview, c *conversion.Cloner) error {
	if err := unversioned.DeepCopy_unversioned_TypeMeta(in.TypeMeta, &out.TypeMeta, c); err != nil {
		return err
	}
	if err := DeepCopy_api_SelfSubjectRulesReviewSpec(in.Spec, &out.Spec, c); err != nil {
		return err
	}
	if err := DeepCopy_api_SubjectRulesReviewStatus(in.Status, &out.Status, c); err != nil {
		return err
	}
	return nil
}

func DeepCopy_api_SelfSubjectRulesReviewSpec(in SelfSubjectRulesReviewSpec, out *SelfSubjectRulesReviewSpec, c *conversion.Cloner) error {
	if in.Scopes != nil {
		in, out := in.Scopes, &out.Scopes
		*out = make([]string, len(in))
		copy(*out, in)
	} else {
		out.Scopes = nil
	}
	return nil
}

func DeepCopy_api_SubjectAccessReview(in SubjectAccessReview, out *SubjectAccessReview, c *conversion.Cloner) error {
	if err := unversioned.DeepCopy_unversioned_TypeMeta(in.TypeMeta, &out.TypeMeta, c); err != nil {
		return err
	}
	if err := DeepCopy_api_Action(in.Action, &out.Action, c); err != nil {
		return err
	}
	out.User = in.User
	if in.Groups != nil {
		in, out := in.Groups, &out.Groups
		*out = make(sets.String)
		for key, val := range in {
			if newVal, err := c.DeepCopy(val); err != nil {
				return err
			} else {
				(*out)[key] = newVal.(sets.Empty)
			}
		}
	} else {
		out.Groups = nil
	}
	if in.Scopes != nil {
		in, out := in.Scopes, &out.Scopes
		*out = make([]string, len(in))
		copy(*out, in)
	} else {
		out.Scopes = nil
	}
	return nil
}

func DeepCopy_api_SubjectAccessReviewResponse(in SubjectAccessReviewResponse, out *SubjectAccessReviewResponse, c *conversion.Cloner) error {
	if err := unversioned.DeepCopy_unversioned_TypeMeta(in.TypeMeta, &out.TypeMeta, c); err != nil {
		return err
	}
	out.Namespace = in.Namespace
	out.Allowed = in.Allowed
	out.Reason = in.Reason
	return nil
}

func DeepCopy_api_SubjectRulesReviewStatus(in SubjectRulesReviewStatus, out *SubjectRulesReviewStatus, c *conversion.Cloner) error {
	if in.Rules != nil {
		in, out := in.Rules, &out.Rules
		*out = make([]PolicyRule, len(in))
		for i := range in {
			if err := DeepCopy_api_PolicyRule(in[i], &(*out)[i], c); err != nil {
				return err
			}
		}
	} else {
		out.Rules = nil
	}
	out.EvaluationError = in.EvaluationError
	return nil
}
