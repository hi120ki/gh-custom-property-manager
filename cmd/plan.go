/*
Copyright © 2025 Hi120ki <12624257+hi120ki@users.noreply.github.com>
*/
package cmd

import (
	"context"
	"os"

	"github.com/hi120ki/gh-custom-property-manager/client"
	"github.com/hi120ki/gh-custom-property-manager/config"
	"github.com/spf13/cobra"
)

var planConfigurationFilePaths []string

// planCmd represents the plan command
var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Show planned changes for custom properties",
	Long: `Plan command shows what changes would be made to GitHub repository custom properties
based on the configuration files. It compares the current state with the desired state
defined in the configuration files and displays the differences.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		githubToken := os.Getenv("GITHUB_TOKEN")
		if githubToken == "" {
			cmd.Println("GITHUB_TOKEN environment variable is not set.")
			return
		}

		if len(planConfigurationFilePaths) == 0 {
			cmd.Println("No configuration files specified. Use --config flag to specify one or more configuration files.")
			return
		}

		githubClient := client.NewClient(ctx, githubToken)
		configManager := config.NewConfig(githubClient)

		// Load all configuration files
		for _, configFilePath := range planConfigurationFilePaths {
			configFile, err := os.Open(configFilePath)
			if err != nil {
				cmd.Printf("Error opening config file %s: %v\n", configFilePath, err)
				return
			}
			defer configFile.Close()

			if err := configManager.LoadConfig(configFile); err != nil {
				cmd.Printf("Error loading config from %s: %v\n", configFilePath, err)
				return
			}
			cmd.Printf("Loaded config file: %s\n", configFilePath)
		}

		// Generate repositories
		if err := configManager.GenerateRepositories(ctx); err != nil {
			cmd.Printf("Error generating repositories: %v\n", err)
			return
		}

		// Generate diffs
		propertyDiffs, err := configManager.GenerateDiffs(ctx)
		if err != nil {
			cmd.Printf("Error generating diffs: %v\n", err)
			return
		}

		if len(propertyDiffs) == 0 {
			cmd.Println("No changes needed.")
			return
		}

		cmd.Println("Planned changes:")
		for _, diff := range propertyDiffs {
			if diff.OldValue == "" {
				cmd.Printf("  %s/%s: Set %s = %s\n", diff.Organization, diff.Repository, diff.PropertyName, diff.NewValue)
			} else {
				cmd.Printf("  %s/%s: Change %s from %s to %s\n", diff.Organization, diff.Repository, diff.PropertyName, diff.OldValue, diff.NewValue)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(planCmd)

	// Add config flag that can be specified multiple times
	planCmd.Flags().StringArrayVar(&planConfigurationFilePaths, "config", []string{}, "Configuration file paths (can be specified multiple times)")
}
