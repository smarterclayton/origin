/*
Copyright 2015 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// deepcopy-gen is a tool for auto-generating DeepCopy functions.
//
// Structs in the input directories with the below line in their comments
// will be ignored during generation.
// // +gencopy=false
package main

import (
	"strings"

	"github.com/golang/glog"

	"k8s.io/kubernetes/cmd/libs/go2idl/args"
	"k8s.io/kubernetes/cmd/libs/go2idl/deepcopy-gen/generators"
	"k8s.io/kubernetes/cmd/libs/go2idl/generator"
)

func main() {
	arguments := args.Default()

	// Override defaults. These are Kubernetes specific input locations.
	arguments.InputDirs = []string{
		"k8s.io/kubernetes/pkg/api",
		"k8s.io/kubernetes/pkg/api/unversioned",
		"k8s.io/kubernetes/pkg/api/v1",
		"k8s.io/kubernetes/pkg/apis/authorization",
		"k8s.io/kubernetes/pkg/apis/authorization/v1beta1",
		"k8s.io/kubernetes/pkg/apis/autoscaling",
		"k8s.io/kubernetes/pkg/apis/autoscaling/v1",
		"k8s.io/kubernetes/pkg/apis/batch",
		"k8s.io/kubernetes/pkg/apis/batch/v1",
		"k8s.io/kubernetes/pkg/apis/componentconfig",
		"k8s.io/kubernetes/pkg/apis/componentconfig/v1alpha1",
		"k8s.io/kubernetes/pkg/apis/extensions",
		"k8s.io/kubernetes/pkg/apis/extensions/v1beta1",
		"github.com/openshift/origin/pkg/authorization/api/v1",
		"github.com/openshift/origin/pkg/authorization/api",
		"github.com/openshift/origin/pkg/build/api/v1",
		"github.com/openshift/origin/pkg/build/api",
		"github.com/openshift/origin/pkg/deploy/api/v1",
		"github.com/openshift/origin/pkg/deploy/api",
		"github.com/openshift/origin/pkg/image/api/v1",
		"github.com/openshift/origin/pkg/image/api",
		"github.com/openshift/origin/pkg/oauth/api/v1",
		"github.com/openshift/origin/pkg/oauth/api",
		"github.com/openshift/origin/pkg/project/api/v1",
		"github.com/openshift/origin/pkg/project/api",
		"github.com/openshift/origin/pkg/quota/api/v1",
		"github.com/openshift/origin/pkg/quota/api",
		"github.com/openshift/origin/pkg/route/api/v1",
		"github.com/openshift/origin/pkg/route/api",
		"github.com/openshift/origin/pkg/sdn/api/v1",
		"github.com/openshift/origin/pkg/sdn/api",
		"github.com/openshift/origin/pkg/template/api/v1",
		"github.com/openshift/origin/pkg/template/api",
		"github.com/openshift/origin/pkg/user/api/v1",
		"github.com/openshift/origin/pkg/user/api",
		"github.com/openshift/origin/pkg/security/api/v1",
		"github.com/openshift/origin/pkg/security/api",
	}

	arguments.GoHeaderFilePath = "hack/boilerplate.txt"

	if err := arguments.Execute(
		generators.NameSystems(),
		generators.DefaultNameSystem(),
		func(context *generator.Context, arguments *args.GeneratorArgs) generator.Packages {
			pkgs := generators.Packages(context, arguments)
			var include generator.Packages
			for _, pkg := range pkgs {
				if strings.HasPrefix(pkg.Path(), "k8s.io/") {
					continue
				}
				include = append(include, pkg)
			}
			return include
		},
	); err != nil {
		glog.Fatalf("Error: %v", err)
	}
	glog.Info("Completed successfully.")
}
