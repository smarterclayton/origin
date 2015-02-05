package cmd

import (
	"fmt"
	"io"
	"os"

	kcmd "github.com/GoogleCloudPlatform/kubernetes/pkg/kubectl/cmd"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/kubectl/resource"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/util/errors"
	"github.com/golang/glog"
	"github.com/spf13/cobra"

	dockerutil "github.com/openshift/origin/pkg/cmd/util/docker"
	newcmd "github.com/openshift/origin/pkg/generate/app/cmd"
)

type usage interface {
	UsageError(commandName string) string
}

const longNewAppDescription = `
Create a new application in OpenShift by specifying source code, templates, and/or images.

Examples:

    $ osc new-app .
    <try to create an application based on the source code in the current directory>

    $ osc new-app mysql
    <use the public DockerHub MySQL image to create an app>

    $ osc new-app myregistry.com/mycompany/mysql
    <use a MySQL image in a private registry to create an app>

    $ osc new-app openshift/ruby-20-centos~git@github.com/mfojtik/sinatra-app-example
    <build an application using the OpenShift Ruby DockerHub image and an example repo>`

func NewCmdNewApplication(f *Factory, out io.Writer) *cobra.Command {
	config := newcmd.NewAppConfig()

	helper := dockerutil.NewHelper()
	cmd := &cobra.Command{
		Use:   "new-app <components> [--code=<path|url>]",
		Short: "Create a new application",
		Long:  longNewAppDescription,

		Run: func(c *cobra.Command, args []string) {
			namespace, err := f.DefaultNamespace(c)
			checkErr(err)

			if dockerClient, _, err := helper.GetClient(); err == nil {
				if err := dockerClient.Ping(); err == nil {
					config.SetDockerClient(dockerClient)
				} else {
					glog.V(2).Infof("No local Docker daemon detected: %v", err)
				}
			}

			osclient, _, err := f.Clients(c)
			if err != nil {
				glog.Fatalf("Error getting client: %v", err)
			}
			config.SetOpenShiftClient(osclient, namespace)

			unknown := config.AddArguments(args)
			if len(unknown) != 0 {
				glog.Fatalf("Did not recognize the following arguments: %v", unknown)
			}

			list, err := config.Run(out)
			if err != nil {
				if errs, ok := err.(errors.Aggregate); ok {
					if len(errs.Errors()) == 1 {
						err = errs.Errors()[0]
					}
				}
				if err == newcmd.ErrNoInputs {
					// TODO: suggest things to the user
					glog.Fatal("You must specify one or more images, image repositories, or source code locations to create an application.")
				}
				if u, ok := err.(usage); ok {
					glog.Fatal(u.UsageError(c.CommandPath()))
				}
				glog.Fatalf("Error: %v", err)
			}

			if len(kcmd.GetFlagString(c, "output")) != 0 {
				if err := kcmd.PrintObject(c, list, f.Factory, out); err != nil {
					glog.Fatalf("Error: %v", err)
				}
				return
			}

			mapper, typer := f.Object(c)
			resourceMapper := &resource.Mapper{typer, mapper, kcmd.ClientMapperForCommand(c, f.Factory)}
			errs := []error{}
			for _, item := range list.Items {
				info, err := resourceMapper.InfoForObject(item)
				if err != nil {
					errs = append(errs, err)
					continue
				}
				data, err := info.Mapping.Codec.Encode(item)
				if err != nil {
					errs = append(errs, err)
					glog.Error(err)
					continue
				}
				if err := resource.NewHelper(info.Client, info.Mapping).Create(namespace, false, data); err != nil {
					errs = append(errs, err)
					glog.Error(err)
					continue
				}
				fmt.Fprintf(out, "%s\n", info.Name)
			}
			if len(errs) != 0 {
				os.Exit(1)
			}
		},
	}

	cmd.Flags().Var(&config.SourceRepositories, "code", "Source code to use to build this application.")
	cmd.Flags().VarP(&config.ImageStreams, "image", "i", "Name of an OpenShift image repository to use in the app.")
	cmd.Flags().Var(&config.DockerImages, "docker-image", "Name of a Docker image to include in the app.")
	cmd.Flags().Var(&config.Groups, "group", "Indicate components that should be grouped together as <comp1>+<comp2>.")
	cmd.Flags().VarP(&config.Environment, "env", "e", "Specify key value pairs of environment variables to set into each container.")
	cmd.Flags().StringVar(&config.TypeOfBuild, "build", "", "Specify the type of build to use if you don't want to detect (docker|source)")

	kcmd.AddPrinterFlags(cmd)

	return cmd
}
