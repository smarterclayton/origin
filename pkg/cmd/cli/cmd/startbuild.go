package cmd

import (
	"fmt"
	"io"

	kapi "github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/fields"
	cmdutil "github.com/GoogleCloudPlatform/kubernetes/pkg/kubectl/cmd/util"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
	"github.com/spf13/cobra"

	buildapi "github.com/openshift/origin/pkg/build/api"
	buildutil "github.com/openshift/origin/pkg/build/util"
	osclient "github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
)

const startBuildLongDesc = `
Manually starts build from existing build or buildConfig

NOTE: This command is experimental and is subject to change in the future.

Examples:

	# Starts build from build configuration matching the name "3bd2ug53b"
	$ %[1]s start-build 3bd2ug53b

	# Starts build from build matching the name "3bd2ug53b"
	$ %[1]s start-build --from-build=3bd2ug53b

	# Starts build from build configuration matching the name "3bd2ug53b" and watches the logs until the build completes or fails
	$ %[1]s start-build 3bd2ug53b --follow
`

func NewCmdStartBuild(fullName string, f *clientcmd.Factory, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start-build (<buildConfig>|--from-build=<build>)",
		Short: "Starts a new build from existing build or buildConfig",
		Long:  fmt.Sprintf(startBuildLongDesc, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			buildName := cmdutil.GetFlagString(cmd, "from-build")
			follow := cmdutil.GetFlagBool(cmd, "follow")
			if len(args) != 1 && len(buildName) == 0 {
				usageError(cmd, "Must pass a name of buildConfig or specify build name with '--from-build' flag")
			}

			client, _, err := f.Clients(cmd)
			checkErr(err)

			namespace, err := f.DefaultNamespace(cmd)
			checkErr(err)

			var newBuild *buildapi.Build
			if len(buildName) == 0 {
				// from build config
				config, err := client.BuildConfigs(namespace).Get(args[0])
				checkErr(err)

				newBuild, err = buildutil.GenerateBuildWithImageTag(config, nil, client.ImageRepositories(kapi.NamespaceAll).(osclient.ImageRepositoryNamespaceGetter))
				checkErr(err)
			} else {
				build, err := client.Builds(namespace).Get(buildName)
				checkErr(err)

				newBuild = buildutil.GenerateBuildFromBuild(build)
			}

			// Start a build
			newBuild, err = client.Builds(namespace).Create(newBuild)
			checkErr(err)

			if follow {
				set := labels.Set(newBuild.Labels)
				selector := labels.SelectorFromSet(set)

				// Add a watcher for the build about to start
				watcher, err := client.Builds(namespace).Watch(selector, fields.Everything(), newBuild.ResourceVersion)
				checkErr(err)
				defer watcher.Stop()

				for event := range watcher.ResultChan() {
					build, ok := event.Object.(*buildapi.Build)
					if !ok {
						checkErr(fmt.Errorf("cannot convert input to Build"))
					}

					// Iterate over watcher's results and search for
					// the build we just started. Also make sure that
					// the build is running, complete, or has failed
					if build.Name == newBuild.Name {
						switch build.Status {
						case buildapi.BuildStatusRunning, buildapi.BuildStatusComplete, buildapi.BuildStatusFailed:
							rd, err := client.BuildLogs(namespace).Redirect(newBuild.Name).Stream()
							checkErr(err)
							defer rd.Close()

							_, err = io.Copy(out, rd)
							checkErr(err)
							break
						}

						if build.Status == buildapi.BuildStatusComplete || build.Status == buildapi.BuildStatusFailed {
							break
						}
					}
				}
			}

			fmt.Fprintf(out, "%s\n", newBuild.Name)
		},
	}
	cmd.Flags().String("from-build", "", "Specify the name of a build which should be re-run")
	cmd.Flags().Bool("follow", false, "Start a build and watch its logs until it completes or fails")
	return cmd
}
