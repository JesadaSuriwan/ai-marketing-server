package perplexity

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ai-marketing/ai-marketing-server/logs"
	"github.com/ai-marketing/ai-marketing-server/providers/citation"
	"github.com/ai-marketing/ai-marketing-server/providers/location"
	"github.com/ai-marketing/ai-marketing-server/providers/usage"
)

const apiURL = "https://api.perplexity.ai/chat/completions"

// defaultModel is Perplexity's standard search-grounded model — every
// response is web-search-grounded by default, unlike OpenAI/Gemini there's
// no separate "tool" flag to enable it. "sonar-pro" is available for deeper
// multi-step search reasoning at higher cost, if ever needed.
const defaultModel = "sonar"

type Client struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		model:      defaultModel,
		httpClient: &http.Client{Timeout: 45 * time.Second},
	}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type searchResult struct {
	Title string `json:"title"`
	Url   string `json:"url"`
}

type chatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

type chatResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	// SearchResults is the current, richer citation field (title + url).
	SearchResults []searchResult `json:"search_results"`
	// Citations is an older/simpler fallback: plain URL strings, no titles.
	Citations []string  `json:"citations"`
	Usage     chatUsage `json:"usage"`
	Error     *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Complete sends prompt to Perplexity's chat completions endpoint — search
// grounding is on by default for the "sonar" model family, no separate tool
// flag needed — and returns the response text, model used, the real source
// URLs it actually cited, and real token usage.
func (c *Client) Complete(prompt, country string) (response string, model string, citations []citation.Citation, tokenUsage usage.Usage, err error) {
	if c.apiKey == "" {
		return "", "", nil, usage.Usage{}, errors.New("perplexity api key not configured")
	}

	logs.Info(fmt.Sprintf("perplexity request: model=%s country=%s input=%q", c.model, country, prompt))

	messages := []chatMessage{}
	if hint := location.Hint(country); hint != "" {
		messages = append(messages, chatMessage{Role: "system", Content: hint})
	}
	messages = append(messages, chatMessage{Role: "user", Content: prompt})

	reqBody, err := json.Marshal(chatRequest{
		Model:    c.model,
		Messages: messages,
	})
	if err != nil {
		return "", "", nil, usage.Usage{}, err
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(reqBody))
	if err != nil {
		return "", "", nil, usage.Usage{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", nil, usage.Usage{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", nil, usage.Usage{}, err
	}

	var parsed chatResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", "", nil, usage.Usage{}, err
	}

	if resp.StatusCode != http.StatusOK {
		if parsed.Error != nil {
			return "", "", nil, usage.Usage{}, fmt.Errorf("perplexity error: %s", parsed.Error.Message)
		}
		return "", "", nil, usage.Usage{}, fmt.Errorf("perplexity request failed with status %d", resp.StatusCode)
	}

	if len(parsed.Choices) == 0 {
		return "", "", nil, usage.Usage{}, errors.New("perplexity returned no choices")
	}
	response = parsed.Choices[0].Message.Content
	if response == "" {
		return "", "", nil, usage.Usage{}, errors.New("perplexity returned no message content")
	}

	if len(parsed.SearchResults) > 0 {
		for _, sr := range parsed.SearchResults {
			if sr.Url == "" {
				continue
			}
			citations = append(citations, citation.Citation{URL: sr.Url, Title: sr.Title})
		}
	} else {
		for _, u := range parsed.Citations {
			if u == "" {
				continue
			}
			citations = append(citations, citation.Citation{URL: u})
		}
	}

	model = parsed.Model
	if model == "" {
		model = c.model
	}

	tokenUsage = usage.Usage{InputTokens: parsed.Usage.PromptTokens, OutputTokens: parsed.Usage.CompletionTokens}
	logs.Info(fmt.Sprintf("perplexity response: model=%s citations=%d tokens=%d/%d text=%q", model, len(citations), tokenUsage.InputTokens, tokenUsage.OutputTokens, response))

	return response, model, citations, tokenUsage, nil
}
