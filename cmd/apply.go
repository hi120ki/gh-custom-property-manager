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

var applyConfigurationFilePaths []string

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply custom property changes to repositories",
	Long: `Apply command executes the changes to GitHub repository custom properties
based on the configuration files. It compares the current state with the desired state
and applies the necessary changes.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		githubToken := os.Getenv("GITHUB_TOKEN")
		if githubToken == "" {
			cmd.Println("GITHUB_TOKEN environment variable is not set.")
			return
		}

		if len(applyConfigurationFilePaths) == 0 {
			cmd.Println("No configuration files specified. Use --config flag to specify one or more configuration files.")
			return
		}

		githubClient := client.NewClient(ctx, githubToken)
		configManager := config.NewConfig(githubClient)

		// Load all configuration files
		for _, configFilePath := range applyConfigurationFilePaths {
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

		cmd.Println("Applying changes:")

		// Apply changes using the ApplyChanges function
		if err := configManager.ApplyChanges(ctx, propertyDiffs); err != nil {
			cmd.Printf("Error applying changes: %v\n", err)
			return
		}

		// Show applied changes
		for _, diff := range propertyDiffs {
			cmd.Printf("  %s/%s: Set %s = %s\n", diff.Organization, diff.Repository, diff.PropertyName, diff.NewValue)
		}

		cmd.Println("All changes applied successfully.")
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)

	// Add config flag that can be specified multiple times
	applyCmd.Flags().StringArrayVar(&applyConfigurationFilePaths, "config", []string{}, "Configuration file paths (can be specified multiple times)")
}
