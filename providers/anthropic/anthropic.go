package anthropic

import (
	"context"
	"errors"

	sdk "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"github.com/ai-marketing/ai-marketing-server/providers/usage"
)

const defaultModel = "claude-opus-4-8"

type Client struct {
	apiKey string
	client sdk.Client
	model  string
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		client: sdk.NewClient(option.WithAPIKey(apiKey)),
		model:  defaultModel,
	}
}

// Complete sends a system + user prompt to Claude and returns the response
// text, the model that produced it, and real token usage.
func (c *Client) Complete(systemPrompt, userPrompt string) (response string, model string, tokenUsage usage.Usage, err error) {
	if c.apiKey == "" {
		return "", "", usage.Usage{}, errors.New("anthropic api key not configured")
	}

	msg, err := c.client.Messages.New(context.Background(), sdk.MessageNewParams{
		Model:     c.model,
		MaxTokens: 2048,
		System:    []sdk.TextBlockParam{{Text: systemPrompt}},
		Messages: []sdk.MessageParam{
			sdk.NewUserMessage(sdk.NewTextBlock(userPrompt)),
		},
	})
	if err != nil {
		return "", "", usage.Usage{}, err
	}

	var text string
	for _, block := range msg.Content {
		if block.Type == "text" {
			text += block.Text
		}
	}
	if text == "" {
		return "", "", usage.Usage{}, errors.New("anthropic returned no text content")
	}

	tokenUsage = usage.Usage{InputTokens: int(msg.Usage.InputTokens), OutputTokens: int(msg.Usage.OutputTokens)}
	return text, c.model, tokenUsage, nil
}
