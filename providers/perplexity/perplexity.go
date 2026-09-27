package perplexity

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ai-marketing/ai-marketing-server/logs"
	"github.com/ai-marketing/ai-marketing-server/providers/citation"
	"github.com/ai-marketing/ai-marketing-server/providers/location"
	"github.com/ai-marketing/ai-marketing-server/providers/usage"
)

// Perplexity retired the Sonar chat-completions endpoint in favor of the
// Agent API — same product, different request/response shape. Confirmed
// against three independent doc pages (migration guide, API reference, and
// the quickstart's own example curl) before changing this, since guessing
// at an external API's schema wrong would just trade one production outage
// for another.
// https://docs.perplexity.ai/docs/agent-api/migrate-from-sonar/overview
const apiURL = "https://api.perplexity.ai/v1/agent"

// defaultPreset replaces the old "sonar" model name — the Agent API takes a
// configuration-intensity preset instead of a model. "fast" is documented as
// the direct Sonar replacement (mirrors the old default of plain "sonar").
const defaultPreset = "fast"

type Client struct {
	apiKey     string
	preset     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		preset:     defaultPreset,
		httpClient: &http.Client{Timeout: 45 * time.Second},
	}
}

type agentRequest struct {
	Preset string `json:"preset"`
	Input  string `json:"input"`
	// Instructions is the Agent API's dedicated system-prompt field — a
	// cleaner equivalent to the old code's hack of prepending a
	// system-role chat message, now that there's no messages array at all.
	Instructions string `json:"instructions,omitempty"`
}

type urlCitation struct {
	Type  string `json:"type"`
	Url   string `json:"url"`
	Title string `json:"title"`
}

type messageContent struct {
	Type        string        `json:"type"`
	Text        string        `json:"text"`
	Annotations []urlCitation `json:"annotations"`
}

type searchResult struct {
	Url     string `json:"url"`
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
}

// outputItem covers only the two item types this app cares about (message
// text and search results) — the Agent API's output array can also carry
// function-call/tool-use step items this app never requests, which
// json.Unmarshal just leaves as zero values here.
type outputItem struct {
	Type    string           `json:"type"`
	Content []messageContent `json:"content"`
	Results []searchResult   `json:"results"`
}

type agentUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type agentResponse struct {
	Model  string       `json:"model"`
	Status string       `json:"status"`
	Output []outputItem `json:"output"`
	Usage  agentUsage   `json:"usage"`
	Error  *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Complete sends prompt to Perplexity's Agent API — search grounding is on
// by default for the "fast" preset, same as Sonar before it — and returns
// the response text, model used, the real source URLs it actually cited,
// and real token usage.
func (c *Client) Complete(prompt, country string) (response string, model string, citations []citation.Citation, tokenUsage usage.Usage, err error) {
	if c.apiKey == "" {
		return "", "", nil, usage.Usage{}, errors.New("perplexity api key not configured")
	}

	logs.Info(fmt.Sprintf("perplexity request: preset=%s country=%s input=%q", c.preset, country, prompt))

	reqBody, err := json.Marshal(agentRequest{
		Preset:       c.preset,
		Input:        prompt,
		Instructions: location.Hint(country),
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

	var parsed agentResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", "", nil, usage.Usage{}, err
	}

	if resp.StatusCode != http.StatusOK || parsed.Status == "failed" {
		if parsed.Error != nil {
			return "", "", nil, usage.Usage{}, fmt.Errorf("perplexity error: %s", parsed.Error.Message)
		}
		return "", "", nil, usage.Usage{}, fmt.Errorf("perplexity request failed with status %d", resp.StatusCode)
	}

	var textParts []string
	var inlineCitations []citation.Citation
	var searchResultCitations []citation.Citation
	seen := map[string]bool{}

	for _, item := range parsed.Output {
		switch item.Type {
		case "message":
			for _, mc := range item.Content {
				if mc.Text != "" {
					textParts = append(textParts, mc.Text)
				}
				for _, a := range mc.Annotations {
					if a.Type != "url_citation" || a.Url == "" || seen[a.Url] {
						continue
					}
					seen[a.Url] = true
					inlineCitations = append(inlineCitations, citation.Citation{URL: a.Url, Title: a.Title})
				}
			}
		case "search_results":
			for _, r := range item.Results {
				if r.Url == "" || seen[r.Url] {
					continue
				}
				seen[r.Url] = true
				searchResultCitations = append(searchResultCitations, citation.Citation{URL: r.Url, Title: r.Title})
			}
		}
	}

	response = strings.Join(textParts, "\n\n")
	if response == "" {
		return "", "", nil, usage.Usage{}, errors.New("perplexity returned no message content")
	}

	// Prefer the inline url_citation annotations — the URLs actually cited
	// in the answer text, equivalent to what the old SearchResults field
	// meant — and only fall back to the broader search_results list (pages
	// the model looked at but didn't necessarily cite) if it cited nothing
	// inline. Same fallback shape as the old Sonar response handling.
	if len(inlineCitations) > 0 {
		citations = inlineCitations
	} else {
		citations = searchResultCitations
	}

	model = parsed.Model
	if model == "" {
		model = c.preset
	}

	tokenUsage = usage.Usage{InputTokens: parsed.Usage.InputTokens, OutputTokens: parsed.Usage.OutputTokens}
	logs.Info(fmt.Sprintf("perplexity response: model=%s citations=%d tokens=%d/%d text=%q", model, len(citations), tokenUsage.InputTokens, tokenUsage.OutputTokens, response))

	return response, model, citations, tokenUsage, nil
}
