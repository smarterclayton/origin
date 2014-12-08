package cmd

import (
	"io"

	kubecmd "github.com/GoogleCloudPlatform/kubernetes/pkg/kubectl/cmd"
	"github.com/openshift/origin/pkg/api/latest"
	appgen "github.com/openshift/origin/pkg/generate/app"
	"github.com/spf13/cobra"
)

func (f *OriginFactory) NewCmdGenerate(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate -s <source>",
		Short: "Generate a configuration based on source or image repository",
		Long: `Generate configuration based on a specified source repository or image repository

Examples:
  $ kubectl generate --source="https://github.com/openshift/ruby-hello-world.git"
  <creates a docker build for the given repository, and a deployment configuration for the resulting image>

  $ kubectl generate --source="https://github.com/openshift/simple-ruby.git" --buildImage="openshift/ruby-20-centos"
  <creates an STI build for the given repository and builder image, as well as a deployment configuration>
  
  $ kubectl generate dockerfile/mongodb
  <creates a deployment config for the given image>
  `,
		Run: func(cmd *cobra.Command, args []string) {
			generator := &appgen.Generator{
				Source:       kubecmd.GetFlagString(cmd, "source"),
				BuilderImage: kubecmd.GetFlagString(cmd, "buildImage"),
				Name:         kubecmd.GetFlagString(cmd, "name"),
				Images:       args,
			}
			cfg, err := generator.Generate()
			checkErr(err)
			data, err := latest.Codec.Encode(cfg)
			checkErr(err)
			_, err = out.Write(data)
			checkErr(err)
		},
	}
	cmd.Flags().StringP("source", "r", "", "Git source repository URL for build config generation")
	cmd.Flags().StringP("buildImage", "i", "", "Builder image for STI builds")
	cmd.Flags().StringP("name", "o", "", "Name to give build-related artifacts. If not specified, it will be inferred from repsotory URL")
	return cmd
}
