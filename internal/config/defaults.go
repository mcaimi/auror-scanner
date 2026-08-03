package config

import "github.com/spf13/viper"

func setDefaults(v *viper.Viper) {
	v.SetDefault("backend.provider", "llama.cpp")
	v.SetDefault("backend.baseurl", "http://lodalhost:11434/v1")
	v.SetDefault("backend.model", "ollama/llama3:1b")
	v.SetDefault("backend.apikey", "ollama/llama3:1b")
	v.SetDefault("backend.temperature", 0.7)
	v.SetDefault("backend.maxtokens", 2048)
	v.SetDefault("backend.usagetracking", true)

	v.SetDefault("skills.skillspath", "skill")
	v.SetDefault("skills.skillfile", "pkgbuild-security-assessment.md")
	v.SetDefault("skills.systemprompt", "")

	v.SetDefault("output.reports", "reports")
}
