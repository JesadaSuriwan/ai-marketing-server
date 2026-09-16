// Package serpapi backs the "google-ai" prompt-run engine. Unlike the other
// engines, it isn't an LLM call — it queries Google Search through SerpApi
// and reads back whatever AI Overview Google actually showed for that
// query, plus the real source links it cited.
package serpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ai-marketing/ai-marketing-server/logs"
	"github.com/ai-marketing/ai-marketing-server/providers/citation"
	"github.com/ai-marketing/ai-marketing-server/providers/usage"
)

const searchURL = "https://serpapi.com/search.json"

// modelName tags prompt_runs/usage rows — there's no real "model" since this
// isn't an LLM call, this is just a stable label for the UI and the usage
// price table (which correctly has no entry for it, so cost shows as an
// honest $0 rather than a fabricated per-token estimate).
const modelName = "google-ai-overview"

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{apiKey: apiKey, httpClient: &http.Client{Timeout: 45 * time.Second}}
}

// textBlock mirrors SerpApi's ai_overview.text_blocks shape — paragraphs and
// lists, optionally nested.
type textBlock struct {
	Type    string      `json:"type"`
	Snippet string      `json:"snippet"`
	List    []textBlock `json:"list"`
}

type reference struct {
	Title string `json:"title"`
	Link  string `json:"link"`
}

// aiOverview covers both the initial /search response (which may only carry
// a page_token for larger overviews) and the follow-up google_ai_overview
// response (which carries the actual content).
type aiOverview struct {
	PageToken  string      `json:"page_token"`
	TextBlocks []textBlock `json:"text_blocks"`
	References []reference `json:"references"`
}

type searchResponse struct {
	AiOverview *aiOverview `json:"ai_overview"`
	Error      string      `json:"error"`
}

// Complete queries Google Search via SerpApi for prompt and returns the AI
// Overview Google showed for it. Most queries don't trigger one at all —
// that's reported back as an error, the same as any other provider
// returning no usable response, so this engine is simply skipped for that
// run rather than saving something empty.
func (c *Client) Complete(prompt string) (response string, model string, citations []citation.Citation, tokenUsage usage.Usage, err error) {
	if c.apiKey == "" {
		return "", "", nil, usage.Usage{}, errors.New("serpapi api key not configured")
	}

	logs.Info(fmt.Sprintf("serpapi request: q=%q", prompt))

	overview, err := c.fetch(url.Values{
		"engine":  {"google"},
		"q":       {prompt},
		"api_key": {c.apiKey},
	})
	if err != nil {
		return "", "", nil, usage.Usage{}, err
	}

	// Larger overviews come back as just a page_token on the initial search
	// — a second request against the dedicated engine resolves the content.
	if overview != nil && len(overview.TextBlocks) == 0 && overview.PageToken != "" {
		overview, err = c.fetch(url.Values{
			"engine":     {"google_ai_overview"},
			"page_token": {overview.PageToken},
			"api_key":    {c.apiKey},
		})
		if err != nil {
			return "", "", nil, usage.Usage{}, err
		}
	}

	if overview == nil || len(overview.TextBlocks) == 0 {
		return "", "", nil, usage.Usage{}, errors.New("no AI Overview shown for this query")
	}

	response = strings.TrimSpace(flattenTextBlocks(overview.TextBlocks))
	if response == "" {
		return "", "", nil, usage.Usage{}, errors.New("AI Overview returned no readable content")
	}

	for _, ref := range overview.References {
		if ref.Link == "" {
			continue
		}
		citations = append(citations, citation.Citation{URL: ref.Link, Title: ref.Title})
	}

	logs.Info(fmt.Sprintf("serpapi response: citations=%d text=%q", len(citations), response))

	return response, modelName, citations, usage.Usage{}, nil
}

func (c *Client) fetch(params url.Values) (*aiOverview, error) {
	req, err := http.NewRequest(http.MethodGet, searchURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("serpapi request failed with status %d", resp.StatusCode)
	}

	var parsed searchResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	if parsed.Error != "" {
		return nil, fmt.Errorf("serpapi error: %s", parsed.Error)
	}

	return parsed.AiOverview, nil
}

func flattenTextBlocks(blocks []textBlock) string {
	var sb strings.Builder
	for _, b := range blocks {
		if b.Snippet != "" {
			sb.WriteString(b.Snippet)
			sb.WriteString("\n")
		}
		if len(b.List) > 0 {
			sb.WriteString(flattenTextBlocks(b.List))
		}
	}
	return sb.String()
}
