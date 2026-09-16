// Package claude provides a search-grounded AIProvider (for running tracked
// prompts and collecting real citations) — distinct from providers/anthropic,
// which is used only for the citation-extraction step. Both wrap the same
// Anthropic SDK/account, but this one enables Claude's web_search tool.
package claude

import (
	"context"
	"errors"
	"fmt"

	sdk "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"github.com/ai-marketing/ai-marketing-server/logs"
	"github.com/ai-marketing/ai-marketing-server/providers/citation"
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

// Complete sends prompt to Claude with the web_search tool enabled and
// forced (tool_choice "any" — Claude's equivalent of OpenAI's
// tool_choice:"required") and returns the search-grounded response text, the
// model used, and the real source URLs it actually cited.
func (c *Client) Complete(prompt string) (response string, model string, citations []citation.Citation, tokenUsage usage.Usage, err error) {
	if c.apiKey == "" {
		return "", "", nil, usage.Usage{}, errors.New("claude api key not configured")
	}

	logs.Info(fmt.Sprintf("claude request: model=%s input=%q", c.model, prompt))

	msg, err := c.client.Messages.New(context.Background(), sdk.MessageNewParams{
		Model:     c.model,
		MaxTokens: 2048,
		Messages: []sdk.MessageParam{
			sdk.NewUserMessage(sdk.NewTextBlock(prompt)),
		},
		// No UserLocation/country bias here — unlike OpenAI/Gemini, Claude's
		// web_search tool rejects "TH" ("Country code TH is not supported"),
		// confirmed against the live API. Search runs ungrounded-by-region
		// until Anthropic adds support for it.
		Tools: []sdk.ToolUnionParam{
			{OfWebSearchTool20260318: &sdk.WebSearchTool20260318Param{}},
		},
		ToolChoice: sdk.ToolChoiceUnionParam{OfAny: &sdk.ToolChoiceAnyParam{}},
	})
	if err != nil {
		return "", "", nil, usage.Usage{}, err
	}

	seen := map[string]bool{}
	for _, block := range msg.Content {
		switch block.Type {
		case "text":
			response += block.Text
			// Citations attached directly to text blocks (the documented,
			// simple path) — in practice Opus's agentic tool loop below
			// (mixing web_search with code_execution) doesn't populate
			// these, but check anyway in case a simpler run does.
			for _, cit := range block.Citations {
				if cit.Type != "web_search_result_location" || cit.URL == "" || seen[cit.URL] {
					continue
				}
				seen[cit.URL] = true
				citations = append(citations, citation.Citation{URL: cit.URL, Title: cit.Title})
			}
		case "web_search_tool_result":
			// The actual real source list lives here — confirmed against
			// the live API that Claude's text blocks come back with empty
			// citations in this agentic flow, so this is the only reliable
			// source of real cited URLs.
			for _, result := range block.Content.OfWebSearchResultBlockArray {
				if result.URL == "" || seen[result.URL] {
					continue
				}
				seen[result.URL] = true
				citations = append(citations, citation.Citation{URL: result.URL, Title: result.Title})
			}
		}
	}
	if response == "" {
		return "", "", nil, usage.Usage{}, errors.New("claude returned no text content")
	}

	tokenUsage = usage.Usage{InputTokens: int(msg.Usage.InputTokens), OutputTokens: int(msg.Usage.OutputTokens)}
	logs.Info(fmt.Sprintf("claude response: model=%s citations=%d tokens=%d/%d text=%q", c.model, len(citations), tokenUsage.InputTokens, tokenUsage.OutputTokens, response))

	return response, c.model, citations, tokenUsage, nil
}
