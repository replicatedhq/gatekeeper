package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/replicatedhq/gatekeeper/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
func RootCmd(c *config.Config, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "gatekeeper",
		Short:         "manage admission controllers in kubernetes",
		Long:          `gatekeeper allows for managing admission controllers using open policy agent in kubernetes`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
			os.Exit(1)
		},
	}
	cobra.OnInitialize(initConfig)

	cmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is /etc/gatekeeper.yaml)")
	cmd.PersistentFlags().String("log-level", "off", "Log level")

	cmd.AddCommand(Status(c, out))
	cmd.AddCommand(Proxy(c, out))

	viper.BindPFlags(cmd.Flags())
	viper.BindPFlags(cmd.PersistentFlags())
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	return cmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	v := viper.New()
	v.AutomaticEnv()

	config.BindEnv(v, "mapstructure")
	c := config.New()
	v.Unmarshal(c)

	if err := RootCmd(c, os.Stdout).Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath("/etc/gatekeeper")
		viper.AddConfigPath("/etc/sysconfig/")
		viper.SetConfigName("gatekeeper")
	}

	viper.AutomaticEnv() // read in environment variables that match
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
