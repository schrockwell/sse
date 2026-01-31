package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/schrockwell/sse/internal/keyfile"
	"github.com/schrockwell/sse/internal/secrets"
	"github.com/spf13/cobra"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Compare keys and values across environments",
	Long: `Analyze the env.toml file to find:
- Keys that are missing from some environments
- Values that are identical across multiple environments

This helps identify configuration inconsistencies and potential copy-paste errors.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		identity, err := keyfile.LoadIdentity()
		if err != nil {
			return err
		}

		f, err := secrets.Load(secrets.DefaultFile)
		if err != nil {
			return err
		}

		// Collect all environment names
		envNames := make([]string, 0, len(f.Environments))
		for name := range f.Environments {
			envNames = append(envNames, name)
		}
		sort.Strings(envNames)

		if len(envNames) < 2 {
			fmt.Println("Need at least 2 environments to analyze.")
			return nil
		}

		// Decrypt all environments
		decrypted := make(map[string]map[string]string)
		for _, envName := range envNames {
			env := f.Environments[envName]
			dec, err := secrets.DecryptEnvironment(env, identity)
			if err != nil {
				return fmt.Errorf("failed to decrypt %s: %w", envName, err)
			}
			decrypted[envName] = dec
		}

		// Collect all keys across all environments
		allKeys := make(map[string]bool)
		for _, env := range decrypted {
			for key := range env {
				allKeys[key] = true
			}
		}

		sortedKeys := make([]string, 0, len(allKeys))
		for key := range allKeys {
			sortedKeys = append(sortedKeys, key)
		}
		sort.Strings(sortedKeys)

		// Check for missing keys
		var missingIssues []string
		for _, key := range sortedKeys {
			var missingFrom []string
			for _, envName := range envNames {
				if _, exists := decrypted[envName][key]; !exists {
					missingFrom = append(missingFrom, envName)
				}
			}
			if len(missingFrom) > 0 && len(missingFrom) < len(envNames) {
				missingIssues = append(missingIssues, fmt.Sprintf("%s is not set in: %s", key, strings.Join(missingFrom, ", ")))
			}
		}

		// Check for equal values and uniquely defined keys
		var equalIssues []string
		var uniqueKeys []string
		for _, key := range sortedKeys {
			// Group environments by value
			valueToEnvs := make(map[string][]string)
			presentInAll := true
			for _, envName := range envNames {
				if value, exists := decrypted[envName][key]; exists {
					valueToEnvs[value] = append(valueToEnvs[value], envName)
				} else {
					presentInAll = false
				}
			}

			// Check if any value appears in multiple environments
			hasEqual := false
			for _, envs := range valueToEnvs {
				if len(envs) > 1 {
					equalIssues = append(equalIssues, fmt.Sprintf("%s is equal in: %s", key, strings.Join(envs, ", ")))
					hasEqual = true
				}
			}

			// Key is uniquely defined if present in all environments with all different values
			if presentInAll && !hasEqual {
				uniqueKeys = append(uniqueKeys, key)
			}
		}

		// Print results
		sections := 0

		if len(missingIssues) > 0 {
			fmt.Println("Missing keys:")
			for _, issue := range missingIssues {
				fmt.Printf("  %s\n", issue)
			}
			sections++
		}

		if len(equalIssues) > 0 {
			if sections > 0 {
				fmt.Println()
			}
			fmt.Println("Equal values:")
			for _, issue := range equalIssues {
				fmt.Printf("  %s\n", issue)
			}
			sections++
		}

		if len(uniqueKeys) > 0 {
			if sections > 0 {
				fmt.Println()
			}
			fmt.Println("Unique values:")
			for _, key := range uniqueKeys {
				fmt.Printf("  %s\n", key)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
}
