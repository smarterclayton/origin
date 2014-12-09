package generate

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/openshift/origin/pkg/generate/server"
)

const longCommandDesc = `
Start a configuration generator

The configuration generator responds to requests of the form:
/generate?source=<sourceURL>
/generate?source=<sourceURL>&baseImage=<baseImage>
/generate?image[s]=image1,image2
`

func NewCommandGenerator(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("%s", name),
		Short: "Run an OpenShift configuration generator",
		Long:  longCommandDesc,
		Run: func(c *cobra.Command, args []string) {
			cfg := server.NewConfig(c.Flag("bindAddr").Value.String())
			server.Serve(cfg)
		},
	}
	cmd.Flags().StringP("bindAddr", "a", ":8080", "Bind address for generator server")
	return cmd
}
