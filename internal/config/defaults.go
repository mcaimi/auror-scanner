package config

import "github.com/spf13/viper"

func setDefaults(v *viper.Viper) {
	// backend section
	v.SetDefault("backend.provider", "llama.cpp")
	v.SetDefault("backend.baseurl", "http://localhost:8080/v1")
	v.SetDefault("backend.model", "unsloth/qwen3.5-9B-gguf:Q4_K_M")
	v.SetDefault("backend.apikey", "")
	v.SetDefault("backend.temperature", 0.7)
	v.SetDefault("backend.maxtokens", 2048)
	v.SetDefault("backend.usagetracking", true)
	v.SetDefault("backend.maxturns", 10)

	// middleware section
	v.SetDefault("middleware.retry.maxretries", 3)
	v.SetDefault("middleware.retry.backoff", 2.0)
	v.SetDefault("middleware.retry.initialdelay", 10)
	v.SetDefault("middleware.retry.maxdelay", 60)

	// skill section
	v.SetDefault("middleware.skills.skillspath", "skill")

	// filesystem section
	v.SetDefault("middleware.filesystem.rootdir", "/tmp")
	v.SetDefault("middleware.filesystem.allowedit", false)

	// prompts section
	v.SetDefault("prompt.promptpath", "prompts")
	v.SetDefault("prompt.promptfile", "auror.prompt")

	// output section
	v.SetDefault("output.reports", "reports")
}
