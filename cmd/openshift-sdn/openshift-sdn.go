package main

import (
	"math/rand"
	"os"
	"time"

	"k8s.io/apiserver/pkg/util/logs"

	"github.com/openshift/origin/pkg/cmd/server/start/network"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	rand.Seed(time.Now().UTC().UnixNano())

	cmd, _ := network.NewCommandStartNetwork("openshift-sdn", os.Stdout, os.Stderr)
	// cmd := &cobra.Command{
	// 	Use: "openshift-sdn",
	// 	Long: heredoc.Doc(`
	// 		Generate Kubelet configuration from node-config.yaml

	// 		This command converts an existing OpenShift node configuration into the appropriate
	// 		Kubelet command-line flags.
	// 	`),
	// 	RunE: func(cmd *cobra.Command, args []string) error {
	// 		configapi.InstallLegacy(configapi.Scheme)
	// 		configapiv1.InstallLegacy(configapi.Scheme)

	// 		if len(configFile) == 0 {
	// 			return fmt.Errorf("you must specify a --config file to read")
	// 		}
	// 		nodeConfig, err := configapilatest.ReadAndResolveNodeConfig(configFile)
	// 		if err != nil {
	// 			return fmt.Errorf("unable to read node config: %v", err)
	// 		}

	// 		if err := node.FinalizeNodeConfig(nodeConfig); err != nil {
	// 			return err
	// 		}

	// 		if glog.V(2) {
	// 			out, _ := yaml.Marshal(nodeConfig)
	// 			glog.V(2).Infof("Node config:\n%s", out)
	// 		}
	// 		return node.WriteKubeletFlags(*nodeConfig)
	// 	},
	// 	SilenceUsage: true,
	// }
	// cmd.Flags().StringVar(&configFile, "config", "", "The config file to convert to Kubelet arguments.")
	// flagtypes.GLog(cmd.PersistentFlags())

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
