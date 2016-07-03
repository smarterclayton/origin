package admin

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	kubectl "k8s.io/kubernetes/pkg/kubectl/cmd"

	"github.com/openshift/openshift-sdn/pkg/cmd/admin/network"
	"github.com/openshift/origin/pkg/cmd/admin/cert"
	diagnostics "github.com/openshift/origin/pkg/cmd/admin/diagnostics"
	"github.com/openshift/origin/pkg/cmd/admin/groups"
	"github.com/openshift/origin/pkg/cmd/admin/node"
	"github.com/openshift/origin/pkg/cmd/admin/policy"
	"github.com/openshift/origin/pkg/cmd/admin/project"
	"github.com/openshift/origin/pkg/cmd/admin/prune"
	"github.com/openshift/origin/pkg/cmd/admin/registry"
	"github.com/openshift/origin/pkg/cmd/admin/router"
	"github.com/openshift/origin/pkg/cmd/cli/cmd"
	"github.com/openshift/origin/pkg/cmd/experimental/buildchain"
	exipfailover "github.com/openshift/origin/pkg/cmd/experimental/ipfailover"
	"github.com/openshift/origin/pkg/cmd/server/admin"
	"github.com/openshift/origin/pkg/cmd/templates"
	cmdutil "github.com/openshift/origin/pkg/cmd/util"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/openshift/origin/pkg/version"
)

const adminLong = `
Administrative Commands

Commands for managing a cluster are exposed here. Many administrative
actions involve interaction with the command-line client as well.`

func NewCommandAdmin(name, fullName string, out io.Writer, errout io.Writer) *cobra.Command {
	// Main command
	cmds := &cobra.Command{
		Use:   name,
		Short: "Tools for managing a cluster",
		Long:  fmt.Sprintf(adminLong),
		Run:   cmdutil.DefaultSubCommandRun(out),
	}

	f := clientcmd.New(cmds.PersistentFlags())

	groups := templates.CommandGroups{
		{
			Message: "Component Installation:",
			Commands: []*cobra.Command{
				router.NewCmdRouter(f, fullName, "router", out),
				exipfailover.NewCmdIPFailoverConfig(f, fullName, "ipfailover", out, errout),
				registry.NewCmdRegistry(f, fullName, "registry", out),
			},
		},
		{
			Message: "Security and Policy:",
			Commands: []*cobra.Command{
				project.NewCmdNewProject(project.NewProjectRecommendedName, fullName+" "+project.NewProjectRecommendedName, f, out),
				policy.NewCmdPolicy(policy.PolicyRecommendedName, fullName+" "+policy.PolicyRecommendedName, f, out, errout),
				groups.NewCmdGroups(groups.GroupsRecommendedName, fullName+" "+groups.GroupsRecommendedName, f, out),
				cert.NewCmdCert(cert.CertRecommendedName, fullName+" "+cert.CertRecommendedName, out, errout),
				admin.NewCommandOverwriteBootstrapPolicy(admin.OverwriteBootstrapPolicyCommandName, fullName+" "+admin.OverwriteBootstrapPolicyCommandName, fullName+" "+admin.CreateBootstrapPolicyFileCommand, out),
			},
		},
		{
			Message: "Node Management:",
			Commands: []*cobra.Command{
				admin.NewCommandNodeConfig(admin.NodeConfigCommandName, fullName+" "+admin.NodeConfigCommandName, out),
				node.NewCommandManageNode(f, node.ManageNodeCommandName, fullName+" "+node.ManageNodeCommandName, out, errout),
				cmdutil.ReplaceCommandName("kubectl", fullName, kubectl.NewCmdCordon(f.Factory, out)),
				cmdutil.ReplaceCommandName("kubectl", fullName, kubectl.NewCmdUncordon(f.Factory, out)),
				cmdutil.ReplaceCommandName("kubectl", fullName, kubectl.NewCmdDrain(f.Factory, out)),
				cmdutil.ReplaceCommandName("kubectl", fullName, kubectl.NewCmdTaint(f.Factory, out)),
				network.NewCmdPodNetwork(network.PodNetworkCommandName, fullName+" "+network.PodNetworkCommandName, f, out),
			},
		},
		{
			Message: "Maintenance:",
			Commands: []*cobra.Command{
				diagnostics.NewCmdDiagnostics(diagnostics.DiagnosticsRecommendedName, fullName+" "+diagnostics.DiagnosticsRecommendedName, out),
				prune.NewCommandPrune(prune.PruneRecommendedName, fullName+" "+prune.PruneRecommendedName, f, out),
				buildchain.NewCmdBuildChain(name, fullName+" "+buildchain.BuildChainRecommendedCommandName, f, out),
			},
		},
		{
			Message: "Configuration:",
			Commands: []*cobra.Command{
				admin.NewCommandCreateKubeConfig(admin.CreateKubeConfigCommandName, fullName+" "+admin.CreateKubeConfigCommandName, out),
				admin.NewCommandCreateClient(admin.CreateClientCommandName, fullName+" "+admin.CreateClientCommandName, out),

				admin.NewCommandCreateBootstrapProjectTemplate(f, admin.CreateBootstrapProjectTemplateCommand, fullName+" "+admin.CreateBootstrapProjectTemplateCommand, out),
				admin.NewCommandCreateBootstrapPolicyFile(admin.CreateBootstrapPolicyFileCommand, fullName+" "+admin.CreateBootstrapPolicyFileCommand, out),

				admin.NewCommandCreateLoginTemplate(f, admin.CreateLoginTemplateCommand, fullName+" "+admin.CreateLoginTemplateCommand, out),
				admin.NewCommandCreateProviderSelectionTemplate(f, admin.CreateProviderSelectionTemplateCommand, fullName+" "+admin.CreateProviderSelectionTemplateCommand, out),
				admin.NewCommandCreateErrorTemplate(f, admin.CreateErrorTemplateCommand, fullName+" "+admin.CreateErrorTemplateCommand, out),
			},
		},
	}

	groups.Add(cmds)
	templates.ActsAsRootCommand(cmds, []string{"options"}, groups...)

	// Deprecated commands that are bundled with the binary but not displayed to end users directly
	deprecatedCommands := []*cobra.Command{
		admin.NewCommandCreateMasterCerts(admin.CreateMasterCertsCommandName, fullName+" "+admin.CreateMasterCertsCommandName, out),
		admin.NewCommandCreateKeyPair(admin.CreateKeyPairCommandName, fullName+" "+admin.CreateKeyPairCommandName, out),
		admin.NewCommandCreateServerCert(admin.CreateServerCertCommandName, fullName+" "+admin.CreateServerCertCommandName, out),
		admin.NewCommandCreateSignerCert(admin.CreateSignerCertCommandName, fullName+" "+admin.CreateSignerCertCommandName, out),
	}
	for _, cmd := range deprecatedCommands {
		// Unsetting Short description will not show this command in help
		cmd.Short = ""
		cmd.Deprecated = fmt.Sprintf("Use '%s ca' instead.", fullName)
		cmds.AddCommand(cmd)
	}

	cmds.AddCommand(
		// part of every root command
		cmd.NewCmdConfig(fullName, "config"),
		cmd.NewCmdCompletion(fullName, f, out),

		// hidden
		cmd.NewCmdOptions(out),
	)

	if name == fullName {
		cmds.AddCommand(version.NewVersionCommand(fullName, false))
	}

	return cmds
}
