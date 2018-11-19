package cli

import (
	"context"
	"io"
	"os"

	"github.com/replicatedhq/gatekeeper/pkg/config"
	"github.com/replicatedhq/gatekeeper/pkg/proxy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Proxy(c *config.Config, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proxy",
		Short: "run the proxy gatekeeper server",
		Long: `
Start and run the proxy server
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := proxy.Get(c, os.Stdout)
			if err != nil {
				return err
			}

			return p.Serve(context.Background())
		},
	}

	cmd.Flags().BoolP("enable-tls", "", true, "enable tls listen only")
	cmd.Flags().StringP("tls-cert-file", "", "", "path to tls cert file")
	cmd.Flags().StringP("tls-key-file", "", "", "path to tls key file")

	viper.BindPFlags(cmd.Flags())
	viper.BindPFlags(cmd.PersistentFlags())
	return cmd
}
