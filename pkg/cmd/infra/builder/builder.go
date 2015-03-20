package builder

import (
	"github.com/spf13/cobra"

	"github.com/openshift/origin/pkg/build/builder/cmd"
	"github.com/openshift/origin/pkg/cmd/templates"
	"github.com/openshift/origin/pkg/version"
)

const longCommandSTIDesc = `
Perform a Source-to-Image Build

This command executes a Source-to-Image build using arguments passed via the environment.
It expects to be run inside of a container.
`

// NewCommandSTIBuilder provides a CLI handler for STI build type
func NewCommandSTIBuilder(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: "Run an OpenShift Source-to-Images build",
		Long:  longCommandSTIDesc,
		Run: func(c *cobra.Command, args []string) {
			cmd.RunSTIBuild()
		},
	}

	templates.UseMainTemplates(cmd)

	cmd.AddCommand(version.NewVersionCommand(name))
	return cmd
}

const longCommandDockerDesc = `
Perform a Docker Build

This command executes a Docker build using arguments passed via the environment.
It expects to be run inside of a container.
`

// NewCommandDockerBuilder provides a CLI handler for Docker build type
func NewCommandDockerBuilder(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: "Run an OpenShift Docker build",
		Long:  longCommandDockerDesc,
		Run: func(c *cobra.Command, args []string) {
			cmd.RunDockerBuild()
		},
	}
	cmd.SetUsageTemplate(templates.MainUsageTemplate())
	cmd.SetHelpTemplate(templates.MainHelpTemplate())
	cmd.AddCommand(version.NewVersionCommand(name))
	return cmd
}
