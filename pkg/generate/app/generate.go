package app

import (
	"fmt"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/runtime"

	config "github.com/openshift/origin/pkg/config/api"
)

type Generator struct {
	Name         string
	Source       string
	BuilderImage string
	Images       []string
}

func (c *Generator) Generate() (*config.Config, error) {
	artifacts := []runtime.Object{}
	if c.Source != "" {
		result, err := c.generateFromSource()
		if err != nil {
			return nil, err
		}
		artifacts = append(artifacts, result...)
	}
	if len(c.Images) > 0 {
		result, err := c.generateImageConfigs()
		if err != nil {
			return nil, err
		}
		artifacts = append(artifacts, result...)
	}

	return &config.Config{
		Items: artifacts,
	}, nil
}

func (c *Generator) generateFromSource() ([]runtime.Object, error) {
	source, err := SourceRefForGitURL(c.Source)
	if err != nil {
		return nil, err
	}
	name := c.Name
	if len(name) == 0 {
		var ok bool
		if name, ok = source.SuggestName(); !ok {
			err := fmt.Errorf("Unable to determine name from source URL %v", source.URL)
			return nil, err
		}
	}
	output := &ImageRef{Name: name, AsImageRepository: true}
	build := &BuildRef{Source: source, Output: output}
	if len(c.BuilderImage) > 0 {
		build.Base = &ImageRef{Name: c.BuilderImage}
	}
	deploy := &DeploymentConfigRef{[]*ImageRef{output}}

	outputRepo, err := output.ImageRepository()
	if err != nil {
		return nil, err
	}
	buildConfig, err := build.BuildConfig()
	if err != nil {
		return nil, err
	}
	deployConfig, err := deploy.DeploymentConfig()
	if err != nil {
		return nil, err
	}
	return []runtime.Object{
		outputRepo,
		buildConfig,
		deployConfig,
	}, nil
}

func (c *Generator) generateImageConfigs() ([]runtime.Object, error) {
	result := []runtime.Object{}
	for _, image := range c.Images {
		imageRef := &ImageRef{Name: image, AsImageRepository: true}
		repo, err := imageRef.ImageRepository()
		if err != nil {
			return nil, err
		}
		result = append(result, repo)
	}
	return result, nil
}
