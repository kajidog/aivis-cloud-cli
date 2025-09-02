package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/kajidog/aivis-cloud-cli/client"
	"github.com/kajidog/aivis-cloud-cli/client/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	apiKey      string
	verbose     bool
	logLevel    string
	logOutput   string
	logFormat   string
	aivisClient *client.Client
)

var rootCmd = &cobra.Command{
	Use:   "aivis-cloud-cli",
	Short: "Aivis Cloud CLI - Text-to-speech synthesis and model management",
	Long: `Aivis Cloud CLI provides command-line interface for Aivis Cloud API.
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
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "Aivis Cloud API key")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output (sets log level to DEBUG)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "INFO", "log level (DEBUG, INFO, WARN, ERROR)")
	rootCmd.PersistentFlags().StringVar(&logOutput, "log-output", "stdout", "log output destination (stdout, stderr, or file path)")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "text", "log output format (text, json)")

	rootCmd.AddCommand(ttsCmd)
	rootCmd.AddCommand(modelsCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(usersCmd)
	rootCmd.AddCommand(paymentCmd)
	rootCmd.AddCommand(McpCmd)
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

	// Configure logging
	if verbose {
		cfg.LogLevel = "DEBUG"
	} else if logLevel != "" {
		cfg.LogLevel = logLevel
	} else if configLogLevel := viper.GetString("log_level"); configLogLevel != "" {
		cfg.LogLevel = configLogLevel
	}

	if logOutput != "" {
		cfg.LogOutput = logOutput
	} else if configLogOutput := viper.GetString("log_output"); configLogOutput != "" {
		cfg.LogOutput = configLogOutput
	}

	if logFormat != "" {
		cfg.LogFormat = logFormat
	} else if configLogFormat := viper.GetString("log_format"); configLogFormat != "" {
		cfg.LogFormat = configLogFormat
	}

    // History settings
    if viper.IsSet("history_enabled") {
        cfg.HistoryEnabled = viper.GetBool("history_enabled")
    }
    if viper.IsSet("history_max_count") {
        if v := viper.GetInt("history_max_count"); v > 0 {
            cfg.HistoryMaxCount = v
        }
    }
    if v := viper.GetString("history_store_path"); v != "" {
        cfg.HistoryStorePath = v
    }

    // For MCP stdio mode, force log output to stderr to avoid protocol contamination
	if isMCPStdioMode() {
		cfg.LogOutput = "stderr"
	}

	var err error
	aivisClient, err = client.NewWithConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}

	// Share client with MCP package
	SetClient(aivisClient)

	return nil
}

// isMCPStdioMode checks if the current command is MCP with stdio transport
func isMCPStdioMode() bool {
	// Check if we're running the mcp command
	if len(os.Args) < 2 || os.Args[1] != "mcp" {
		return false
	}
	
	// Check if transport is stdio (default) or explicitly set to stdio
	for i, arg := range os.Args {
		if arg == "--transport" && i+1 < len(os.Args) {
			return os.Args[i+1] == "stdio"
		}
		if strings.HasPrefix(arg, "--transport=") {
			return strings.TrimPrefix(arg, "--transport=") == "stdio"
		}
	}
	
	// Default is stdio for mcp command
	return true
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
