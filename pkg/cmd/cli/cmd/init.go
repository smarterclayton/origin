package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/golang/glog"
	"github.com/spf13/cobra"

	//	kapi "k8s.io/kubernetes/pkg/api"
	//kclient "k8s.io/kubernetes/pkg/client/unversioned"
	kclientcmd "k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	clientapi "k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	//"k8s.io/kubernetes/pkg/kubectl/resource"
	"k8s.io/kubernetes/pkg/util/sets"

	cmdutil "github.com/openshift/origin/pkg/cmd/util"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	app "github.com/openshift/origin/pkg/generate/app"
	appcmd "github.com/openshift/origin/pkg/generate/app/cmd"
	"github.com/openshift/origin/pkg/generate/git"
)

const (
	initLong = `
Initialize an application from a Git repository

`
)

type InitOptions struct {
	Out       io.Writer
	ErrOut    io.Writer
	Git       git.Repository
	Config    kclientcmd.ClientConfig
	Directory string
	Remote    string

	InitTypeDirect    bool
	InitTypeWebhook   bool
	InitTypeGit       bool
	InitTypeHotDeploy bool

	LoginCmd      string
	NewProjectCmd string

	NewAppDefaultsFn func() (*appcmd.AppConfig, error)
}

// NewCmdInit initializes a directory with Git.
func NewCmdInit(fullName string, f *clientcmd.Factory, out io.Writer) *cobra.Command {
	o := &InitOptions{
		Out:    out,
		Git:    git.NewRepository(),
		Config: f.OpenShiftClientConfig,

		LoginCmd:      fmt.Sprintf("%s login", fullName),
		NewProjectCmd: fmt.Sprintf("%s new-project", fullName),
	}
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a Git repository for building",
		Long:  fmt.Sprintf(initLong, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			kcmdutil.CheckErr(o.Complete(f, cmd, args))
			err := o.Run()
			if err == cmdutil.ErrExit {
				os.Exit(1)
			}
			kcmdutil.CheckErr(err)
		},
	}
	cmd.Flags().StringVar(&o.Remote, "remote", o.Remote, "Specify the Git remote to use as the upstream for this repository")
	cmd.Flags().BoolVar(&o.InitTypeDirect, "direct", o.InitTypeDirect, "Initialize this repository to build directly from local code")
	cmd.Flags().BoolVar(&o.InitTypeWebhook, "webhook", o.InitTypeWebhook, "Initialize the remote for this repository as a build")
	cmd.Flags().BoolVar(&o.InitTypeGit, "git", o.InitTypeGit, "Initialize a Git server on OpenShift to push to")
	cmd.Flags().BoolVar(&o.InitTypeHotDeploy, "hot-deploy", o.InitTypeHotDeploy, "Initialize this repository for hot-deployment")
	return cmd
}

func (o *InitOptions) Complete(f *clientcmd.Factory, cmd *cobra.Command, args []string) error {
	if o.ErrOut == nil {
		o.ErrOut = cmd.Out()
	}
	o.NewAppDefaultsFn = func() (*appcmd.AppConfig, error) {
		config := appcmd.NewAppConfig()
		config.Out, config.ErrOut = o.Out, o.ErrOut
		if err := CompleteAppConfig(config, f, cmd, nil); err != nil {
			return nil, err
		}
		return config, nil
	}
	switch len(args) {
	case 0:
		o.Directory = "."
	case 1:
		o.Directory = args[0]
	default:
		return kcmdutil.UsageError(cmd, "Accepts a single optional Git directory to initialize")
	}
	return nil
}

type initData struct {
	Version   string `git:"io.openshift.version"`
	Server    string `git:"io.openshift.server"`
	Namespace string `git:"io.openshift.namespace"`

	BuildRemote string `git:"io.openshift.build-remote"`
	BuildBranch string `git:"io.openshift.build-branch"`

	Builder string `git:"io.openshift.builder"`
}

func (d *initData) Empty() bool {
	return len(d.Version) == 0
}

func (o *InitOptions) Run() error {
	info, errs := o.Git.GetInfo(o.Directory)
	if len(errs) > 0 {
		glog.V(2).Infof("Unable to check repository %q: %v", o.Directory, errs)
		return fmt.Errorf("%q is not a Git repository or repository info could not be loaded", o.Directory)
	}
	short := info.CommitID
	if len(short) > 20 {
		short = short[:7]
	}

	data := &initData{}
	if err := git.UnmarshalLocalConfig(o.Git, o.Directory, data); err != nil {
		return fmt.Errorf("unable to load init information from Git repository %q: %v", o.Directory, err)
	}

	if !data.Empty() {
		c, err := o.Config.RawConfig()
		if err != nil {
			glog.V(2).Infof("Unable to get config: %v", err)
			return fmt.Errorf("You must log in to a cluster with '%s' first or set up a configuration file", o.LoginCmd)
		}
		if len(data.Server) > 0 && len(contextsForServer(c, data.Server)) == 0 {
			return fmt.Errorf("You are not logged in - run '%s %s' to reconnect to this server", o.LoginCmd, data.Server)
		}
		switch {
		case len(data.Builder) > 0:
			fmt.Fprintf(o.Out, "This repository is linked to build config %q in project %q on %q", data.Builder, data.Namespace, data.Server)
		case len(data.BuildRemote) > 0:
			_, ok, err := o.Git.GetOriginURL(o.Directory, data.BuildRemote)
			if err != nil {
				return fmt.Errorf("unable to load repository remotes from %q: %v", o.Directory, err)
			}
			if !ok {
				return fmt.Errorf("The configured OpenShift remote was not found in %q - run 'git remote add %q URL' to restore it", data.BuildRemote, o.Directory, data.BuildRemote)
			}
			remoteBranch := data.BuildBranch
			if len(remoteBranch) == 0 {
				remoteBranch = "master"
			}
			// TODO: try to login via Git and prompt to reauthenticate if necessary
			fmt.Fprintf(o.Out, "This repository is configured to deploy when you 'git push %q %q'", data.BuildRemote, remoteBranch)
		default:
			fmt.Fprintf(o.Out, "This repository is linked to build config %q in project %q on %q", data.Builder, data.Namespace, data.Server)
		}
		return nil
	}

	cfg, err := o.Config.ClientConfig()
	if err != nil {
		glog.V(2).Infof("Unable to get client config: %v", err)
		return fmt.Errorf("You must log in to a cluster with '%s' first or set up a configuration file", o.LoginCmd)
	}
	ns, defaulted, err := o.Config.Namespace()
	if err != nil || defaulted {
		glog.V(2).Infof("Unable to get client namespace: %v", err)
		return fmt.Errorf("You must create a project with '%s' to continue", o.NewProjectCmd)
	}
	data.Server = cfg.Host
	data.Namespace = ns

	hasBranch := len(info.Branch) > 0

	var remotes []string
	if len(o.Remote) > 0 {
		remotes = append(remotes, o.Remote)
	}
	location, hasRemote, err := o.Git.GetOriginURL(o.Directory, remotes...)
	if err != nil {
		return fmt.Errorf("unable to locate Git remote for %q: %v", o.Directory, err)
	}

	switch {
	case hasBranch && hasRemote:
		fmt.Fprintf(o.Out, "On branch %s @ %s with remote %s\n", info.Branch, short, location)
	case hasBranch && !hasRemote:
		fmt.Fprintf(o.Out, "On branch %s @ %s, using binary builds\n", info.Branch, short)
	case hasRemote:
		fmt.Fprintf(o.Out, "No branch @ %s with remote %s\n", short, location)
	default:
		fmt.Fprintf(o.Out, "No branch @ %s, using binary builds\n", short)
	}

	repoRoot, err := o.Git.GetRootDir(o.Directory)
	if err != nil {
		return fmt.Errorf("unable to find root of Git repository: %v", err)
	}

	var input string

	switch {
	case o.InitTypeDirect, o.InitTypeWebhook, o.InitTypeHotDeploy:
		// perform detection
		config := appcmd.NewAppConfig()
		info, err := config.Detector.Detect(repoRoot, false)
		if err != nil || len(info.Types) == 0 {
			if err != nil && err != app.ErrNoLanguageDetected {
				return fmt.Errorf("unable to detect an appropriate base image for this repository: %v", err)
			}
			fmt.Fprintf(o.ErrOut, heredoc.Doc(`

        We were unable to detect what language or framework to use for this repository.
        Pass --with=IMAGE to provide a Docker image to build this code on top of, or search
        for frameworks on the server.
      `))
			return cmdutil.ErrExit
		}
		set := sets.NewString()
		for _, t := range info.Types {
			set.Insert(t.Platform)
		}
		if set.Len() > 1 {
			fmt.Fprintf(o.ErrOut, heredoc.Doc(`

        The repository contains multiple languages or frameworks. Run the command again with
        the --with=TYPE flag to specify the right flag (or provide your own image):

          %s
      `), strings.Join(set.List(), ", "))
			return cmdutil.ErrExit
		}
		input = set.List()[0]
		fmt.Fprintf(o.Out, "Repository appears to %q\n", input)

	case o.InitTypeGit:
		// detection will be done on push
	default:
		fmt.Fprintf(o.ErrOut, heredoc.Doc(`

      Choose how you want to deploy the code in this Git repository:

      * Build from your local code using a binary build [--direct]

        This is the simplest setup. Your local code will be uploaded on each build, but you won't
        be able to rebuild if the base image changes. Ensures that the code you build is exactly
        the same as the code you deploy.

      * Build from the remote Git repository and trigger builds with webhooks [--webhook]

        Create a build that uses your upstream Git remote as the input. The code will be retrieved
        from that upstream and built - works best if you are already collaborating with others on
        GitHub or a hosted Git repository. You'll have to configure your remote repository to send
        a webhook notification to your OpenShift server and decide which branch should be tracked.

      * Set up a Git server on OpenShift and push code to that repository to build [--git]

        This allows you to collaborate with others when you don't have or want a hosted Git repository.
        Each time you git push, your code will be built and deployed. A new 'openshift' remote will be
        created in the local Git repository for you to push to.

      * Use hot-deployment to rapidly iterate on your code [--hot-deploy]

        For dynamic languages like JavaScript or Ruby, or compiled languages like Java where your IDE
        compiles .class files, you can mirror your local directory to the running instance in OpenShift
        and have changes be reflected instantly. Only works if the base image is properly configured,
        and you may need to specify additional environment variables to enable code reload.

      Re-run the init command with one of the flags defined above to continue.
    `))
		return cmdutil.ErrExit
	}

	switch {
	case o.InitTypeDirect:
		config, err := o.NewAppDefaultsFn()
		if err != nil {
			return err
		}
		config.SourceRepositories = []string{repoRoot}
		config.DryRun = true
		result, err := config.Run()
		if err != nil {
			return err
		}
		glog.V(4).Infof("result: %v", result)

		// If we can't find an exact match, prompt user to tell us which image to use as a build environment
		// Create a new binary build from that type
		// Start the first build and stream the output
		fmt.Fprintf(o.Out, "Build succeeded - to build again run 'oc start-build'\n")

	case o.InitTypeWebhook:
		if !hasRemote {
			return fmt.Errorf("A Git remote must be specified - use --remote=NAME to tell us which one to use")
		}
		if !hasBranch {
			return fmt.Errorf("You are not currently on a branch in this Git repository - use 'git checkout BRANCH' to switch the one you want to build")
		}
		webhookGitHub := "https://server/path"
		webhookGeneric := "https://server/path"
		// Check for secrets?
		// Do type detection for this repository
		// If we can't find an exact match, prompt user to tell us which image to use as a build environment
		// Run new-build equivalent
		fmt.Fprintf(o.Out, heredoc.Docf(`
      A build of %[1]q has been created.

      To build whenever commits are pushed, you'll need to add a webhook trigger to the repository.

      If this is a GitHub repository:

      1. Browse to Settings -> Webhooks & services and click 'Add a webhook'.
      2. Set the Payload URL field to:

        %[2]s

      3. Click 'Add webhook'.

      If you are using another type of Git server, perform an HTTP GET request using curl or another
      tool like:

        curl %[3]s

      You can also trigger a build of the configured %[4]q branch at any time with 'oc start-build'.
    `, location, webhookGitHub, webhookGeneric, info.Branch))

	case o.InitTypeHotDeploy:
		// Identify the base image for this app.
		// Check whether the base image has hot deploy markers
		// Warn if hot deploy doesn't look enabled
		// Create a build config with remote (if set) and perform a binary push to create the first build
		// Stream build output
		fmt.Fprintf(o.Out, "Initial build completed - make a change and then run 'oc rsync' to see it live.\n")
	case o.InitTypeGit:
		remote := o.Remote
		if len(remote) == 0 {
			remote = "openshift"
		}
		// Check whether a local Git server is deployed
		// Create and push
		// Print output
		fmt.Fprintf(o.Out, "Run 'git push %s %s' to build and deploy this repository.\n", remote, "master")

	default:
	}

	return nil
}

func contextsForServer(cfg clientapi.Config, server string) map[string]*clientapi.Context {
	out := make(map[string]*clientapi.Context)
	for _, c := range cfg.Clusters {
		if c.Server != server {
			continue
		}
		for name, ctx := range cfg.Contexts {
			if ctx.Cluster != name {
				continue
			}
			out[name] = ctx
		}
	}
	return out
}
