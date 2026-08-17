package config

import "github.com/spf13/viper"

type Config struct {
	Backend    BackendConfig
	Skills     SkillsConfig
	Prompt     PromptConfig
	Output     OutputConfig
	Middleware MiddlewareConfig
}

type OutputConfig struct {
	Reports string
}

type BackendConfig struct {
	Provider      string
	BaseURL       string
	APIKey        string
	Model         string
	Temperature   float64
	MaxTokens     int64
	UsageTracking bool
	MaxTurns      int
}

type SkillsConfig struct {
	SkillsPath string
}

type FilesystemConfig struct {
	RootDir   string
	AllowEdit bool
}

type PromptConfig struct {
	PromptPath string
	PromptFile string
}

type MiddlewareConfig struct {
	Retry      RetryConfig
	Skills     SkillsConfig
	Filesystem FilesystemConfig
}

type RetryConfig struct {
	MaxRetries   int
	Backoff      float64
	InitialDelay int
	MaxDelay     int
}

func Load(configPath string) (*Config, error) {
	v := viper.New()

	setDefaults(v)

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("auror")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME/.auror")
		v.AddConfigPath("/etc/auror")
	}

	v.SetEnvPrefix("AUROR")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
