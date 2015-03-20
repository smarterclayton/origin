package app

import (
	"log"
	"net/url"
	"testing"

	kapi "github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/runtime"

	"github.com/openshift/origin/pkg/api/latest"
	build "github.com/openshift/origin/pkg/build/api"
	imageapi "github.com/openshift/origin/pkg/image/api"
)

func testImageInfo() *imageapi.DockerImage {
	return &imageapi.DockerImage{
		Config: imageapi.DockerConfig{},
	}
}

func TestWithType(t *testing.T) {
	out := &Generated{
		Items: []runtime.Object{
			&build.BuildConfig{
				ObjectMeta: kapi.ObjectMeta{
					Name: "foo",
				},
			},
			&kapi.Service{
				ObjectMeta: kapi.ObjectMeta{
					Name: "foo",
				},
			},
		},
	}

	builds := []build.BuildConfig{}
	if !out.WithType(&builds) {
		t.Errorf("expected true")
	}
	if len(builds) != 1 {
		t.Errorf("unexpected slice: %#v", builds)
	}

	buildPtrs := []*build.BuildConfig{}
	if out.WithType(&buildPtrs) {
		t.Errorf("expected false")
	}
	if len(buildPtrs) != 0 {
		t.Errorf("unexpected slice: %#v", buildPtrs)
	}
}

func TestSimpleBuildConfig(t *testing.T) {
	url, err := url.Parse("https://github.com/openshift/origin.git")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	source := &SourceRef{URL: url}
	build := &BuildRef{Source: source}
	config, err := build.BuildConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.Name != "origin" {
		t.Errorf("unexpected name: %#v", config)
	}

	output := &ImageRef{
		DockerImageReference: imageapi.DockerImageReference{
			Registry:  "myregistry",
			Namespace: "openshift",
			Name:      "origin",
		},
	}
	build = &BuildRef{Source: source, Output: output}
	config, err = build.BuildConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.Name != "origin" || config.Parameters.Output.To.Name != "origin" {
		t.Errorf("unexpected name: %#v", config)
	}
}

func TestSimpleDeploymentConfig(t *testing.T) {
	image := &ImageRef{
		DockerImageReference: imageapi.DockerImageReference{
			Registry:  "myregistry",
			Namespace: "openshift",
			Name:      "origin",
		},
		Info:              testImageInfo(),
		AsImageRepository: true,
	}
	deploy := &DeploymentConfigRef{Images: []*ImageRef{image}}
	config, err := deploy.DeploymentConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.Name != "origin" || len(config.Triggers) != 2 || config.Template.ControllerTemplate.Template.Spec.Containers[0].Image != image.String() {
		t.Errorf("unexpected value: %#v", config)
	}
}

func TestImageRefDeployableContainerPorts(t *testing.T) {
	tests := []struct {
		name          string
		inputPorts    map[string]struct{}
		expectedPorts map[int]string
		expectError   bool
	}{
		{
			name: "tcp implied, individual ports",
			inputPorts: map[string]struct{}{
				"123": {},
				"456": {},
			},
			expectedPorts: map[int]string{
				123: "TCP",
				456: "TCP",
			},
			expectError: false,
		},
		{
			name: "tcp implied, multiple ports",
			inputPorts: map[string]struct{}{
				"123 456":  {},
				"678 1123": {},
			},
			expectedPorts: map[int]string{
				123:  "TCP",
				678:  "TCP",
				456:  "TCP",
				1123: "TCP",
			},
			expectError: false,
		},
		{
			name: "tcp and udp, individual ports",
			inputPorts: map[string]struct{}{
				"123/tcp": {},
				"456/udp": {},
			},
			expectedPorts: map[int]string{
				123: "TCP",
				456: "UDP",
			},
			expectError: false,
		},
		{
			name: "tcp implied, multiple ports",
			inputPorts: map[string]struct{}{
				"123/tcp 456/udp":  {},
				"678/udp 1123/tcp": {},
			},
			expectedPorts: map[int]string{
				123:  "TCP",
				456:  "UDP",
				678:  "UDP",
				1123: "TCP",
			},
			expectError: false,
		},
		{
			name: "invalid format",
			inputPorts: map[string]struct{}{
				"123/tcp abc": {},
			},
			expectedPorts: map[int]string{},
			expectError:   true,
		},
	}
	for _, test := range tests {
		imageRef := &ImageRef{
			DockerImageReference: imageapi.DockerImageReference{
				Namespace: "test",
				Name:      "image",
				Tag:       "latest",
			},
			Info: &imageapi.DockerImage{
				Config: imageapi.DockerConfig{
					ExposedPorts: test.inputPorts,
				},
			},
		}
		container, _, err := imageRef.DeployableContainer()
		if err != nil && !test.expectError {
			t.Errorf("%s: unexpected error: %v", test.name, err)
			continue
		}
		if err == nil && test.expectError {
			t.Errorf("%s: got no error and expected an error", test.name)
			continue
		}
		if test.expectError {
			continue
		}
		remaining := test.expectedPorts
		for _, port := range container.Ports {
			proto, ok := remaining[port.ContainerPort]
			if !ok {
				t.Errorf("%s: got unexpected port: %v", test.name, port)
				continue
			}
			if kapi.Protocol(proto) != port.Protocol {
				t.Errorf("%s: got unexpected protocol %s for port %v", test.name, port.Protocol, port)
			}
			delete(remaining, port.ContainerPort)
		}
		if len(remaining) > 0 {
			t.Errorf("%s: did not find expected ports: %#v", test.name, remaining)
		}
	}
}

func TestSourceRefBuildSourceURI(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "URL without hash",
			input:    "https://github.com/openshift/ruby-hello-world.git",
			expected: "https://github.com/openshift/ruby-hello-world.git",
		},
		{
			name:     "URL with hash",
			input:    "https://github.com/openshift/ruby-hello-world.git#testref",
			expected: "https://github.com/openshift/ruby-hello-world.git",
		},
	}
	for _, tst := range tests {
		u, _ := url.Parse(tst.input)
		s := SourceRef{
			URL: u,
		}
		buildSource, _ := s.BuildSource()
		if buildSource.Git.URI != tst.expected {
			t.Errorf("%s: unexpected build source URI: %s. Expected: %s", tst.name, buildSource.Git.URI, tst.expected)
		}
	}
}

func ExampleGenerateSimpleDockerApp() {
	// TODO: determine if the repo is secured prior to fetching
	// TODO: determine whether we want to clone this repo, or use it directly. Using it directly would require setting hooks
	// if we have source, assume we are going to go into a build flow.
	// TODO: get info about git url: does this need STI?
	url, _ := url.Parse("https://github.com/openshift/origin.git")
	source := &SourceRef{URL: url}
	// generate a local name for the repo
	name, _ := source.SuggestName()
	// BUG: an image repo (if we want to create one) needs to tell other objects its pullspec, but we don't know what that will be
	// until the object is placed into a namespace and we lookup what registry (registries?) serve the object.
	// QUESTION: Is it ok for generation to require a namespace?  Do we want to be able to create apps with builds, image repos, and
	// deployment configs in templates (hint: yes).
	// SOLUTION? Make deployment config accept unqualified image repo names (foo) and then prior to creating the RC resolve those.
	output := &ImageRef{
		DockerImageReference: imageapi.DockerImageReference{
			Name: name,
		},
		AsImageRepository: true,
	}
	// create our build based on source and input
	// TODO: we might need to pick a base image if this is STI
	build := &BuildRef{Source: source, Output: output}
	// take the output image and wire it into a deployment config
	deploy := &DeploymentConfigRef{Images: []*ImageRef{output}}

	outputRepo, _ := output.ImageRepository()
	buildConfig, _ := build.BuildConfig()
	deployConfig, _ := deploy.DeploymentConfig()
	items := []runtime.Object{
		outputRepo,
		buildConfig,
		deployConfig,
	}
	out := &kapi.List{
		Items: items,
	}

	data, err := latest.Codec.Encode(out)
	if err != nil {
		log.Fatalf("Unable to generate output: %v", err)
	}
	log.Print(string(data))
	// output:
}
