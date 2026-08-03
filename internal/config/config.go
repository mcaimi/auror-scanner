package config

import "github.com/spf13/viper"

type Config struct {
	Backend BackendConfig
	Skills  SkillsConfig
	Output  OutputConfig
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
}

type SkillsConfig struct {
	SkillsPath   string
	SkillFile    string
	SystemPrompt string
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

