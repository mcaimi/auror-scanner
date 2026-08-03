package flows

import (
	"context"
	"fmt"
	"strings"

	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	"github.com/mcaimi/auror/internal/backend"
	"github.com/mcaimi/auror/internal/config"
	"github.com/mcaimi/auror/internal/utils"
)

type AurorPackageInput struct {
	// PackageName is the PKGBUILD source: a local filesystem path or an HTTP/HTTPS URL.
	PackageName string `json:"pkgname"`
}

type AurorAnalysisResult struct {
	Result string `json:"result"`
}

// AurorAnalyticsFlow defines and registers a genkit flow that accepts a PKGBUILD
// source (URL or filesystem path) and returns a structured security analysis.
func AurorAnalyticsFlow(c *backend.OpenAIContext, cfg *config.Config) (*core.Flow[AurorPackageInput, AurorAnalysisResult, struct{}], error) {
	skillContent, err := utils.LoadSkillFile(&cfg.Skills)
	if err != nil {
		return nil, err
	}

	// Use everything up to the "Prompt Template" section as the system prompt so
	// the model receives the full assessment procedure without the meta-instructions
	// meant for the caller.
	systemPrompt := skillContent
	userPrompt := "Analyze this PKGBUILD:"
	if idx := strings.Index(skillContent, "## Prompt Template"); idx != -1 {
		systemPrompt = strings.TrimSpace(skillContent[:idx])
		userPrompt = strings.TrimSpace(skillContent[idx:])
	}

	// define flow in genkit. No streaming, the agent produces a report
	flow := genkit.DefineFlow(
		c.Gkit(),
		"auror-pkgbuild-analysis",
		func(ctx context.Context, input AurorPackageInput) (AurorAnalysisResult, error) {
			pkgbuildContent, err := utils.FetchPKGBUILD(input.PackageName)
			if err != nil {
				return AurorAnalysisResult{}, fmt.Errorf("loading PKGBUILD: %w", err)
			}

			// inject prompt template
			userPrompt = fmt.Sprintf("%s\n---\n%s\n---", userPrompt, pkgbuildContent)
			result, err := c.GenerateTextStreaming(ctx, systemPrompt, userPrompt)
			if err != nil {
				return AurorAnalysisResult{}, fmt.Errorf("generating analysis: %w", err)
			}

			return AurorAnalysisResult{Result: result}, nil
		},
	)

	return flow, nil
}
