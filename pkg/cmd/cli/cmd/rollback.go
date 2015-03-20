package cmd

import (
	"fmt"
	"io"

	kapi "github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	kubectl "github.com/GoogleCloudPlatform/kubernetes/pkg/kubectl"
	cmdutil "github.com/GoogleCloudPlatform/kubernetes/pkg/kubectl/cmd/util"
	"github.com/spf13/cobra"

	latest "github.com/openshift/origin/pkg/api/latest"
	describe "github.com/openshift/origin/pkg/cmd/cli/describe"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	deployapi "github.com/openshift/origin/pkg/deploy/api"
)

const rollbackLongDesc = `
Revert part of an application back to a previous deployment.

When you run this command your deployment configuration will be updated to match
the provided deployment. By default only the pod and container configuration
will be changed and scaling or trigger settings will be left as-is. Note that
environment variables and volumes are included in rollbacks, so if you've
recently updated security credentials in your environment your previous
deployment may not have the correct values.

If you would like to review the outcome of the rollback, pass '--dry-run' to print
a human-readable representation of the updated deployment configuration instead of
executing the rollback. This is useful if you're not quite sure what the outcome
will be.

Examples:

	# Perform a rollback
	$ %[1]s rollback deployment-1

	# See what the rollback will look like, but don't perform the rollback
	$ %[1]s rollback deployment-1 --dry-run

	# Perform the rollback manually by piping the JSON of the new config back to %[1]s
	$ %[1]s rollback deployment-1 --output=json | %[1]s update deploymentConfigs deployment -f -
`

func NewCmdRollback(fullName string, f *clientcmd.Factory, out io.Writer) *cobra.Command {
	rollback := &deployapi.DeploymentConfigRollback{
		Spec: deployapi.DeploymentConfigRollbackSpec{
			IncludeTemplate: true,
		},
	}

	cmd := &cobra.Command{
		Use:   "rollback <from-deployment>",
		Short: "Revert part of an application back to a previous deployment.",
		Long:  fmt.Sprintf(rollbackLongDesc, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 || len(args[0]) == 0 {
				usageError(cmd, "A deployment name is required.")
			}

			rollback.Spec.From.Name = args[0]

			outputFormat := cmdutil.GetFlagString(cmd, "output")
			outputTemplate := cmdutil.GetFlagString(cmd, "template")
			dryRun := cmdutil.GetFlagBool(cmd, "dry-run")

			osClient, _, err := f.Clients()
			checkErr(err)

			namespace, err := f.DefaultNamespace()
			checkErr(err)

			// Generate the rollback config
			newConfig, err := osClient.DeploymentConfigs(namespace).Rollback(rollback)
			checkErr(err)

			// If dry-run is specified, describe the rollback and exit
			if dryRun {
				describer := describe.NewDeploymentConfigDescriberForConfig(newConfig)
				description, descErr := describer.Describe(newConfig.Namespace, newConfig.Name)
				checkErr(descErr)
				out.Write([]byte(description))
				return
			}

			// If an output format is specified, display the rollback config JSON and exit
			// WITHOUT performing a rollback.
			if len(outputFormat) > 0 {
				printer, _, perr := kubectl.GetPrinter(outputFormat, outputTemplate)
				checkErr(perr)
				versionedPrinter := kubectl.NewVersionedPrinter(printer, kapi.Scheme, latest.Version)
				versionedPrinter.PrintObj(newConfig, out)
				return
			}

			// Apply the rollback config
			_, updateErr := osClient.DeploymentConfigs(namespace).Update(newConfig)
			checkErr(updateErr)
		},
	}

	cmd.Flags().BoolVar(&rollback.Spec.IncludeTriggers, "change-triggers", false, "Include the previous deployment's triggers in the rollback")
	cmd.Flags().BoolVar(&rollback.Spec.IncludeStrategy, "change-strategy", false, "Include the previous deployment's strategy in the rollback")
	cmd.Flags().BoolVar(&rollback.Spec.IncludeReplicationMeta, "change-scaling-settings", false, "Include the previous deployment's replicationController replica count and selector in the rollback")
	cmd.Flags().BoolP("dry-run", "d", false, "Instead of performing the rollback, describe what the rollback will look like in human-readable form")
	cmd.Flags().StringP("output", "o", "", "Instead of performing the rollback, print the updated deployment configuration in the specified format (json|yaml|template|templatefile)")
	cmd.Flags().StringP("template", "t", "", "Template string or path to template file to use when -o=template or -o=templatefile.")

	return cmd
}
