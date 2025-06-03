/*
Copyright © 2025 Hi120ki <12624257+hi120ki@users.noreply.github.com>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// These will be set by goreleaser
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gh-custom-property-manager",
	Short: "Manage GitHub repository custom properties efficiently",
	Long: `GitHub Custom Property Manager is a CLI tool for managing GitHub repository
custom properties at scale. It allows you to define custom properties in YAML
configuration files and apply them to multiple repositories.

Features:
  - Bulk configuration of GitHub repository custom properties
  - Preview changes before applying (plan command)
  - Apply changes to repositories (apply command)
  - YAML-based configuration management

Examples:
  gh-custom-property-manager plan --config property.yaml
  gh-custom-property-manager apply --config property.yaml`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gh-custom-property-manager.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// Add version flag
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built at: %s)", version, commit, date)
}
