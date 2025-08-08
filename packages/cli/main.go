package main

import (
	"fmt"
	"os"

	"github.com/kajidog/aiviscloud-mcp/client"
	"github.com/kajidog/aiviscloud-mcp/client/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile    string
	apiKey     string
	verbose    bool
	aivisClient *client.Client
)

var rootCmd = &cobra.Command{
	Use:   "aivis-cli",
	Short: "AivisCloud CLI - Text-to-speech synthesis and model management",
	Long: `AivisCloud CLI provides command-line interface for AivisCloud API.
Features include text-to-speech synthesis, audio playback, and model management.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip client initialization for config commands
		if cmd.Parent() != nil && cmd.Parent().Name() == "config" {
			return nil
		}
		if cmd.Name() == "config" {
			return nil
		}
		return initializeClient()
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.aivis-cli.yaml)")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "AivisCloud API key")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	rootCmd.AddCommand(ttsCmd)
	rootCmd.AddCommand(modelsCmd)
	rootCmd.AddCommand(configCmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".aivis-cli")
	}

	viper.SetEnvPrefix("AIVIS")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil && verbose {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func initializeClient() error {
	if apiKey == "" {
		apiKey = viper.GetString("api_key")
	}
	if apiKey == "" {
		return fmt.Errorf("API key is required. Set it via --api-key flag, AIVIS_API_KEY environment variable, or config file")
	}

	cfg := config.NewConfig(apiKey)
	if baseURL := viper.GetString("base_url"); baseURL != "" {
		cfg.BaseURL = baseURL
	}
	if timeout := viper.GetDuration("timeout"); timeout > 0 {
		cfg.HTTPTimeout = timeout
	}

	var err error
	aivisClient, err = client.NewWithConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}

	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}