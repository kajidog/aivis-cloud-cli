package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
	Long:  "Manage CLI configuration settings",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  "Display the current configuration settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Current Configuration:")
		fmt.Println("======================")
		
		// Show config file being used
		if viper.ConfigFileUsed() != "" {
			fmt.Printf("Config file: %s\n", viper.ConfigFileUsed())
		} else {
			fmt.Println("Config file: None (using defaults)")
		}
		fmt.Println()

		// Show all settings
		settings := viper.AllSettings()
		if len(settings) == 0 {
			fmt.Println("No configuration settings found")
			return nil
		}

		for key, value := range settings {
			// Don't show sensitive values like API keys
			if key == "api_key" {
				if value != "" {
					fmt.Printf("%s: [REDACTED]\n", key)
				} else {
					fmt.Printf("%s: [NOT SET]\n", key)
				}
			} else {
				fmt.Printf("%s: %v\n", key, value)
			}
		}

		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set configuration value",
	Long:  "Set a configuration key-value pair",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		// Set the value
		viper.Set(key, value)

		// Save to config file
		if err := saveConfig(); err != nil {
			return fmt.Errorf("failed to save configuration: %v", err)
		}

		fmt.Printf("Set %s = %s\n", key, value)
		return nil
	},
}

var configUnsetCmd = &cobra.Command{
	Use:   "unset [key]",
	Short: "Unset configuration value",
	Long:  "Remove a configuration setting",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]

		// Check if key exists
		if !viper.IsSet(key) {
			fmt.Printf("Key '%s' is not set\n", key)
			return nil
		}

		// Remove the key by setting it to nil
		viper.Set(key, nil)

		// Save to config file
		if err := saveConfig(); err != nil {
			return fmt.Errorf("failed to save configuration: %v", err)
		}

		fmt.Printf("Unset %s\n", key)
		return nil
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration",
	Long:  "Create a new configuration file with default settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %v", err)
		}

		configPath := fmt.Sprintf("%s/.aivis-cli.yaml", home)

		// Check if config file already exists
		if _, err := os.Stat(configPath); err == nil {
			overwrite, _ := cmd.Flags().GetBool("force")
			if !overwrite {
				fmt.Printf("Configuration file already exists at %s\n", configPath)
				fmt.Println("Use --force to overwrite")
				return nil
			}
		}

		// Set default configuration
		viper.SetConfigFile(configPath)
		
		// Set some default values
		if !viper.IsSet("api_key") {
			viper.Set("api_key", "")
		}
		if !viper.IsSet("base_url") {
			viper.Set("base_url", "https://api.aivis-project.com")
		}
		if !viper.IsSet("timeout") {
			viper.Set("timeout", "60s")
		}
		if !viper.IsSet("default_playback_mode") {
			viper.Set("default_playback_mode", "immediate")
		}
		if !viper.IsSet("default_model_uuid") {
			viper.Set("default_model_uuid", "")
		}
		if !viper.IsSet("default_format") {
			viper.Set("default_format", "wav")
		}
		if !viper.IsSet("default_volume") {
			viper.Set("default_volume", 1.0)
		}
		if !viper.IsSet("default_rate") {
			viper.Set("default_rate", 1.0)
		}
		if !viper.IsSet("default_pitch") {
			viper.Set("default_pitch", 0.0)
		}
		if !viper.IsSet("use_simplified_tts_tools") {
			viper.Set("use_simplified_tts_tools", false)
		}
		if !viper.IsSet("default_wait_for_end") {
			viper.Set("default_wait_for_end", false)
		}

		// Write config file
		if err := viper.WriteConfig(); err != nil {
			return fmt.Errorf("failed to write configuration file: %v", err)
		}

		fmt.Printf("Configuration file created at: %s\n", configPath)
		fmt.Println("\nPlease set your API key:")
		fmt.Printf("  %s config set api_key YOUR_API_KEY\n", os.Args[0])
		
		return nil
	},
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit configuration file",
	Long:  "Open configuration file in the default editor",
	RunE: func(cmd *cobra.Command, args []string) error {
		configFile := viper.ConfigFileUsed()
		if configFile == "" {
			return fmt.Errorf("no configuration file found. Run 'config init' first")
		}

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "nano" // fallback editor
		}

		fmt.Printf("Opening %s with %s...\n", configFile, editor)
		
		// Note: In a real implementation, you would execute the editor
		// For this example, we'll just show the command
		fmt.Printf("Run: %s %s\n", editor, configFile)
		
		return nil
	},
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration",
	Long:  "Check if the current configuration is valid",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check required settings
		apiKey := viper.GetString("api_key")
		if apiKey == "" {
			return fmt.Errorf("API key is not set. Use: config set api_key YOUR_API_KEY")
		}

		baseURL := viper.GetString("base_url")
		if baseURL == "" {
			fmt.Println("Warning: base_url is not set, using default")
		}

		// Validate playback mode if set
		playbackMode := viper.GetString("default_playback_mode")
		if playbackMode != "" {
			validModes := []string{"immediate", "queue", "no_queue"}
			valid := false
			for _, mode := range validModes {
				if playbackMode == mode {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("invalid default_playback_mode: %s. Valid values: %s", playbackMode, strings.Join(validModes, ", "))
			}
		}

		fmt.Println("Configuration is valid ✓")
		return nil
	},
}

func saveConfig() error {
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		// No config file set, create default one
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		configFile = fmt.Sprintf("%s/.aivis-cli.yaml", home)
		viper.SetConfigFile(configFile)
	}

	return viper.WriteConfig()
}

func init() {
	// Config init command flags
	configInitCmd.Flags().Bool("force", false, "Overwrite existing configuration file")

	// Add subcommands to config command
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configUnsetCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configValidateCmd)
}