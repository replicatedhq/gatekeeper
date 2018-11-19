package cli

import (
	"context"
	"io"
	"strings"

	"github.com/replicatedhq/gatekeeper/pkg/config"
	"github.com/replicatedhq/gatekeeper/pkg/gatekeeper"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Status(c *config.Config, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "view the status of current admission policies",
		Long: `
View the status of all currently installed and deployed Admission Controllers.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			g, err := gatekeeper.Get(c, out)
			if err != nil {
				return err
			}

			return g.RunStatus(context.Background())
		},
	}

	viper.BindPFlags(cmd.Flags())
	viper.BindPFlags(cmd.PersistentFlags())
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	return cmd
}
