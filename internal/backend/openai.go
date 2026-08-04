package backend

import (
	"context"
	"fmt"
	"strings"

	"github.com/mcaimi/auror/internal/config"
	"github.com/mcaimi/auror/internal/tools"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/compat_oai"
	"github.com/firebase/genkit/go/plugins/middleware"
	"github.com/openai/openai-go"
	"github.com/sirupsen/logrus"
)

type OpenAIContext struct {
	log    *logrus.Logger
	gkit   *genkit.Genkit
	gcfg   *config.Config
	gparms *openai.ChatCompletionNewParams
}

func (c *OpenAIContext) SetCompletionParams() {
	c.gparms = &openai.ChatCompletionNewParams{
		Temperature:         openai.Float(c.gcfg.Backend.Temperature),
		MaxCompletionTokens: openai.Int(c.gcfg.Backend.MaxTokens),
		Model:               c.gcfg.Backend.Model,
		StreamOptions:       openai.ChatCompletionStreamOptionsParam{IncludeUsage: openai.Bool(c.gcfg.Backend.UsageTracking)},
	}
	c.log.Info(
		fmt.Sprintf(
			"MaxTokens %d, Temperature: %f, Model: %s",
			c.gcfg.Backend.MaxTokens,
			c.gcfg.Backend.Temperature,
			c.gcfg.Backend.Model,
		),
	)
}

func (c *OpenAIContext) GetOpenAIAdapter(
	ctx context.Context,
	cfg *config.Config,
	log *logrus.Logger,
) error {
	c.log = log
	c.gcfg = cfg

	c.log.Info("Initializing OpenAI-Compatible Backend")
	c.gkit = genkit.Init(
		ctx,
		genkit.WithPlugins(
			&compat_oai.OpenAICompatible{
				Provider: c.gcfg.Backend.Provider,
				APIKey:   c.gcfg.Backend.APIKey,
				BaseURL:  c.gcfg.Backend.BaseURL,
			},
			&middleware.Middleware{},
		),
	)
	if c.gkit == nil {
		return fmt.Errorf("genkit init returned nil")
	}

	c.log.Info("BaseURL: ", c.gcfg.Backend.BaseURL)
	return nil
}

func (c *OpenAIContext) Gkit() *genkit.Genkit {
	return c.gkit
}

// GenerateTextStreaming streams the response and prints a running token count
// estimate to stdout, returning the full generated text when done.
func (c *OpenAIContext) GenerateTextStreaming(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	if c.gparms == nil {
		return "", fmt.Errorf("completion params not set: call SetCompletionParams before GenerateTextStreaming")
	}
	if c.gkit == nil {
		return "", fmt.Errorf("backend not initialized: call GetOpenAIAdapter before GenerateTextStreaming")
	}

	var buf strings.Builder
	var responseChars, reasoningChars int

	// use callback-based streaming capabilities
	// i need to process tokens as they are generated
	resp, err := genkit.Generate(
		ctx,
		c.gkit,
		ai.WithModelName(c.gparms.Model),
		ai.WithConfig(c.gparms),
		ai.WithSystem(systemPrompt),
		ai.WithPrompt(userPrompt),
		ai.WithTools(tools.RegisterShellTool(c.gkit)),
		ai.WithUse(&middleware.Retry{
			MaxRetries:     c.gcfg.Middleware.Retry.MaxRetries,
			InitialDelayMs: c.gcfg.Middleware.Retry.InitialDelay,
			MaxDelayMs:     c.gcfg.Middleware.Retry.MaxDelay,
			BackoffFactor:  c.gcfg.Middleware.Retry.Backoff,
		},
		),
		ai.WithStreaming(func(ctx context.Context, chunk *ai.ModelResponseChunk) error {
			for _, part := range chunk.Content {
				if part.IsReasoning() {
					reasoningChars += len(part.Text)
				} else {
					buf.WriteString(part.Text)
					responseChars += len(part.Text)
				}
			}
			// Approximate token count: ~4 chars per token
			fmt.Printf("\r(Running Estimate) Tokens generated: ~%d response | ~%d reasoning", responseChars/4, reasoningChars/4)

			return nil
		}),
	)
	c.log.Info("Genration Done")

	if err != nil {
		return "", err
	}

	if resp != nil && resp.Usage != nil {
		c.log.Info(
			fmt.Sprintf("Total tokens — input: %d, output: %d (reasoning: %d)\n",
				resp.Usage.InputTokens, resp.Usage.OutputTokens, resp.Usage.ThoughtsTokens),
		)
	}

	return buf.String(), nil
}

// One-shot full text generation with no streaming.
func (c *OpenAIContext) GenerateText(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	if c.gparms == nil {
		return "", fmt.Errorf("completion params not set: call SetCompletionParams before GenerateText")
	}
	if c.gkit == nil {
		return "", fmt.Errorf("backend not initialized: call GetOpenAIAdapter before GenerateText")
	}

	return genkit.GenerateText(
		ctx,
		c.gkit,
		ai.WithModelName(c.gparms.Model),
		ai.WithConfig(c.gparms),
		ai.WithSystem(systemPrompt),
		ai.WithPrompt(userPrompt),
	)
}
