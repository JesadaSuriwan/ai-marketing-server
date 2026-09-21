package openai

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
	"github.com/ai-marketing/ai-marketing-server/providers/usage"
)

const apiURL = "https://api.openai.com/v1/responses"

const defaultModel = "gpt-4.1"

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

// Citation is a real, search-grounded source URL the model actually cited.
// Aliased to the shared type so both openai.Client and gemini.Client satisfy
// the same pkg/promptrun/service.AIProvider interface.
type Citation = citation.Citation

type responsesRequest struct {
	Model      string          `json:"model"`
	Tools      []responsesTool `json:"tools"`
	ToolChoice string          `json:"tool_choice"`
	Input      string          `json:"input"`
}

type responsesTool struct {
	Type         string        `json:"type"`
	UserLocation *userLocation `json:"user_location,omitempty"`
}

// userLocation biases web search results to a region. Without it, search has
// no geographic refinement and can return market-irrelevant results (e.g. US
// retailers for a Thai-language query) — unlike chatgpt.com, which infers
// this from the browser session, the bare API has no such signal.
type userLocation struct {
	Type    string `json:"type"`
	Country string `json:"country"`
}

type responseAnnotation struct {
	Type  string `json:"type"`
	URL   string `json:"url"`
	Title string `json:"title"`
}

type responseContent struct {
	Type        string               `json:"type"`
	Text        string               `json:"text"`
	Annotations []responseAnnotation `json:"annotations"`
}

type responseOutputItem struct {
	Type    string            `json:"type"`
	Content []responseContent `json:"content"`
}

type responsesUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type responsesResponse struct {
	Output []responseOutputItem `json:"output"`
	Usage  responsesUsage       `json:"usage"`
	Error  *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Complete sends prompt to OpenAI's Responses API with the web_search tool
// enabled and returns the search-grounded response text, the model used,
// the real source URLs the model actually cited, and real token usage.
func (c *Client) Complete(prompt, country string) (response string, model string, citations []Citation, tokenUsage usage.Usage, err error) {
	if c.apiKey == "" {
		return "", "", nil, usage.Usage{}, errors.New("openai api key not configured")
	}

	logs.Info(fmt.Sprintf("openai request: model=%s country=%s tool_choice=required input=%q", c.model, country, prompt))

	webSearch := responsesTool{Type: "web_search"}
	if country != "" {
		webSearch.UserLocation = &userLocation{Type: "approximate", Country: country}
	}

	reqBody, err := json.Marshal(responsesRequest{
		Model: c.model,
		Tools: []responsesTool{webSearch},
		// Force the search to actually run — left as "auto" the model will
		// sometimes ask a clarifying question instead of searching, which
		// defeats the point of tracking what it actually finds.
		ToolChoice: "required",
		Input:      prompt,
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

	var parsed responsesResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", "", nil, usage.Usage{}, err
	}

	if resp.StatusCode != http.StatusOK {
		if parsed.Error != nil {
			return "", "", nil, usage.Usage{}, fmt.Errorf("openai error: %s", parsed.Error.Message)
		}
		return "", "", nil, usage.Usage{}, fmt.Errorf("openai request failed with status %d", resp.StatusCode)
	}

	for _, item := range parsed.Output {
		if item.Type != "message" {
			continue
		}
		for _, content := range item.Content {
			if content.Type != "output_text" {
				continue
			}
			response = content.Text
			for _, a := range content.Annotations {
				if a.Type == "url_citation" && a.URL != "" {
					citations = append(citations, Citation{URL: a.URL, Title: a.Title})
				}
			}
		}
	}

	if response == "" {
		return "", "", nil, usage.Usage{}, errors.New("openai returned no message content")
	}

	tokenUsage = usage.Usage{InputTokens: parsed.Usage.InputTokens, OutputTokens: parsed.Usage.OutputTokens}
	logs.Info(fmt.Sprintf("openai response: model=%s citations=%d tokens=%d/%d text=%q", c.model, len(citations), tokenUsage.InputTokens, tokenUsage.OutputTokens, response))

	return response, c.model, citations, tokenUsage, nil
}
