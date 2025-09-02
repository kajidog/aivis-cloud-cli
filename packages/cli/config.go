package main

import (
    "errors"
    "fmt"
    "net/url"
    "os"
    "strconv"
    "strings"
    "time"

    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

// ConfigSpec defines an allowed configuration key, its type and validation.
type ConfigSpec struct {
    Key         string
    Type        string
    Description string
    Validate    func(string) (any, error)
}

func parseBool(s string) (any, error) {
    v, err := strconv.ParseBool(s)
    if err != nil {
        return nil, fmt.Errorf("expected boolean (true/false)")
    }
    return v, nil
}

func parseFloatInRange(min, max float64) func(string) (any, error) {
    return func(s string) (any, error) {
        f, err := strconv.ParseFloat(s, 64)
        if err != nil {
            return nil, fmt.Errorf("expected number")
        }
        if f < min || f > max {
            return nil, fmt.Errorf("value out of range (%.2f..%.2f)", min, max)
        }
        return f, nil
    }
}

func parseIntPositive(s string) (any, error) {
    i, err := strconv.Atoi(s)
    if err != nil || i <= 0 {
        return nil, fmt.Errorf("expected positive integer")
    }
    return i, nil
}

func parseDuration(s string) (any, error) {
    d, err := time.ParseDuration(s)
    if err != nil {
        return nil, fmt.Errorf("expected duration (e.g. 60s, 2m)")
    }
    if d <= 0 {
        return nil, errors.New("duration must be positive")
    }
    return s, nil // store as string to keep file human-friendly
}

func parseEnum(allowed ...string) func(string) (any, error) {
    return func(s string) (any, error) {
        for _, a := range allowed {
            if s == a {
                return s, nil
            }
        }
        return nil, fmt.Errorf("invalid value. allowed: %s", strings.Join(allowed, ", "))
    }
}

func getConfigSpecs() []ConfigSpec {
    return []ConfigSpec{
        {Key: "api_key", Type: "string", Description: "Aivis Cloud API key", Validate: func(s string) (any, error) { return s, nil }},
        {Key: "base_url", Type: "string", Description: "API base URL", Validate: func(s string) (any, error) {
            if s == "" { return s, nil }
            u, err := url.Parse(s); if err != nil || (u.Scheme != "http" && u.Scheme != "https") { return nil, fmt.Errorf("invalid URL") }
            return s, nil
        }},
        {Key: "timeout", Type: "duration", Description: "HTTP timeout (e.g. 60s)", Validate: parseDuration},
        {Key: "default_playback_mode", Type: "enum", Description: "Playback mode (immediate|queue|no_queue)", Validate: parseEnum("immediate", "queue", "no_queue")},
        {Key: "default_model_uuid", Type: "string", Description: "Default voice model UUID", Validate: func(s string) (any, error) { return s, nil }},
        {Key: "default_format", Type: "enum", Description: "Default audio format (wav|mp3|flac|aac|opus)", Validate: parseEnum("wav", "mp3", "flac", "aac", "opus")},
        {Key: "default_channels", Type: "enum", Description: "Audio channels (mono|stereo)", Validate: parseEnum("mono", "stereo")},
        {Key: "default_volume", Type: "number", Description: "Default TTS volume (0.0..2.0)", Validate: parseFloatInRange(0.0, 2.0)},
        {Key: "default_rate", Type: "number", Description: "Default speaking rate (0.5..2.0)", Validate: parseFloatInRange(0.5, 2.0)},
        {Key: "default_pitch", Type: "number", Description: "Default pitch (-1.0..1.0)", Validate: parseFloatInRange(-1.0, 1.0)},
        {Key: "default_ssml", Type: "bool", Description: "Enable SSML by default", Validate: parseBool},
        {Key: "default_emotional_intensity", Type: "number", Description: "Default emotional intensity (0.0..2.0)", Validate: parseFloatInRange(0.0, 2.0)},
        {Key: "default_tempo_dynamics", Type: "number", Description: "Default tempo dynamics (0.0..2.0)", Validate: parseFloatInRange(0.0, 2.0)},
        {Key: "default_leading_silence", Type: "number", Description: "Leading silence seconds (0.0..10.0)", Validate: parseFloatInRange(0.0, 10.0)},
        {Key: "default_trailing_silence", Type: "number", Description: "Trailing silence seconds (0.0..10.0)", Validate: parseFloatInRange(0.0, 10.0)},
        {Key: "default_wait_for_end", Type: "bool", Description: "Wait for playback completion by default", Validate: parseBool},
        {Key: "use_simplified_tts_tools", Type: "bool", Description: "Use simplified TTS tools for MCP", Validate: parseBool},
        {Key: "history_enabled", Type: "bool", Description: "Enable TTS history management", Validate: parseBool},
        {Key: "history_max_count", Type: "int", Description: "Max history records to keep (>0)", Validate: parseIntPositive},
        {Key: "history_store_path", Type: "string", Description: "History storage directory", Validate: func(s string) (any, error) { return s, nil }},
        {Key: "log_level", Type: "enum", Description: "Log level (DEBUG|INFO|WARN|ERROR)", Validate: parseEnum("DEBUG", "INFO", "WARN", "ERROR")},
        {Key: "log_output", Type: "enum|string", Description: "Log output (stdout|stderr|file path)", Validate: func(s string) (any, error) {
            if s == "stdout" || s == "stderr" || s == "" { return s, nil }
            return s, nil
        }},
        {Key: "log_format", Type: "enum", Description: "Log format (text|json)", Validate: parseEnum("text", "json")},
    }
}

func findSpec(key string) *ConfigSpec {
    for _, s := range getConfigSpecs() {
        if s.Key == key { return &s }
    }
    return nil
}

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
        raw := args[1]

        spec := findSpec(key)
        if spec == nil {
            return fmt.Errorf("unknown key: %s. See 'config keys' for available settings", key)
        }

        val, err := spec.Validate(raw)
        if err != nil {
            return fmt.Errorf("invalid value for %s (%s): %v", key, spec.Type, err)
        }

        viper.Set(key, val)
        if err := saveConfig(); err != nil {
            return fmt.Errorf("failed to save configuration: %v", err)
        }

        fmt.Printf("Set %s = %v\n", key, val)
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

        // Validate all known keys if set
        for _, spec := range getConfigSpecs() {
            if viper.IsSet(spec.Key) {
                // Convert the stored value to string form for validation where needed
                var raw string
                switch vv := viper.Get(spec.Key).(type) {
                case string:
                    raw = vv
                default:
                    raw = fmt.Sprintf("%v", vv)
                }
                if _, err := spec.Validate(raw); err != nil {
                    return fmt.Errorf("%s: %v", spec.Key, err)
                }
            }
        }

		fmt.Println("Configuration is valid ✓")
		return nil
	},
}

var configKeysCmd = &cobra.Command{
    Use:   "keys",
    Short: "List available configuration keys",
    Long:  "Show all supported configuration keys, expected types, and descriptions.",
    RunE: func(cmd *cobra.Command, args []string) error {
        fmt.Println("Available configuration keys:")
        fmt.Println("--------------------------------")
        for _, spec := range getConfigSpecs() {
            current := "(not set)"
            if viper.IsSet(spec.Key) {
                current = fmt.Sprintf("current=%v", viper.Get(spec.Key))
                if spec.Key == "api_key" && viper.GetString(spec.Key) != "" {
                    current = "current=[REDACTED]"
                }
            }
            fmt.Printf("- %s: %s [%s] %s\n", spec.Key, spec.Description, spec.Type, current)
        }
        fmt.Println("\nExamples:")
        fmt.Println("  aivis-cloud-cli config set default_playback_mode queue")
        fmt.Println("  aivis-cloud-cli config set timeout 90s")
        fmt.Println("  aivis-cloud-cli config set default_format mp3")
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
    configCmd.AddCommand(configKeysCmd)
}
