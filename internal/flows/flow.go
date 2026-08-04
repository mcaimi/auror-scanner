package flows

import (
	"context"
	"fmt"

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
	promptContent, err := utils.LoadPromptFile(&cfg.Prompt)
	if err != nil {
		return nil, err
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
			userPrompt := fmt.Sprintf("%s\n---\n%s\n---", promptContent, pkgbuildContent)
			result, err := c.GenerateTextStreaming(ctx, userPrompt)
			if err != nil {
				return AurorAnalysisResult{}, fmt.Errorf("generating analysis: %w", err)
			}

			return AurorAnalysisResult{Result: result}, nil
		},
	)

	return flow, nil
}
