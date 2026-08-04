package config

import "github.com/spf13/viper"

func setDefaults(v *viper.Viper) {
	// backend section
	v.SetDefault("backend.provider", "llama.cpp")
	v.SetDefault("backend.baseurl", "http://lodalhost:11434/v1")
	v.SetDefault("backend.model", "ollama/llama3:1b")
	v.SetDefault("backend.apikey", "ollama/llama3:1b")
	v.SetDefault("backend.temperature", 0.7)
	v.SetDefault("backend.maxtokens", 2048)
	v.SetDefault("backend.usagetracking", true)

	// middleware section
	v.SetDefault("middleware.retry.maxretries", 3)
	v.SetDefault("middleware.retry.backoff", 2.0)
	v.SetDefault("middleware.retry.initialdelay", 10)
	v.SetDefault("middleware.retry.maxdelay", 60)

	// skill section
	v.SetDefault("skills.skillspath", "skill")
	v.SetDefault("skills.skillfile", "pkgbuild-security-assessment.md")

	// output section
	v.SetDefault("output.reports", "reports")
}
