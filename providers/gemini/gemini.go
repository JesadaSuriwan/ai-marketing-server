package gemini

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

const apiURLTemplate = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s"

// "-latest" alias always resolves to Google's current flash-tier model,
// rather than a pinned dated version that can get cut off from new API keys
// (confirmed: "gemini-2.5-flash" 404s for this key with "no longer available
// to new users" even though it still appears in the models list).
const defaultModel = "gemini-flash-latest"

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

type generateRequest struct {
	SystemInstruction *content  `json:"system_instruction,omitempty"`
	Contents          []content `json:"contents"`
	Tools             []tool    `json:"tools"`
}

type content struct {
	Parts []part `json:"parts"`
}

type part struct {
	Text string `json:"text"`
}

// tool enables Grounding with Google Search — Gemini's equivalent of
// OpenAI's web_search tool. Without it the model has no live web access and
// can't return real cited URLs.
type tool struct {
	GoogleSearch struct{} `json:"google_search"`
}

type groundingChunk struct {
	Web struct {
		Uri   string `json:"uri"`
		Title string `json:"title"`
	} `json:"web"`
}

type usageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
}

type generateResponse struct {
	Candidates []struct {
		Content struct {
			Parts []part `json:"parts"`
		} `json:"content"`
		GroundingMetadata *struct {
			GroundingChunks []groundingChunk `json:"groundingChunks"`
		} `json:"groundingMetadata"`
	} `json:"candidates"`
	UsageMetadata usageMetadata `json:"usageMetadata"`
	Error         *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Complete sends prompt to Gemini with Google Search grounding enabled and
// returns the search-grounded response text, the model used, the real source
// URLs the model actually cited, and real token usage.
func (c *Client) Complete(prompt, country string) (response string, model string, citations []citation.Citation, tokenUsage usage.Usage, err error) {
	if c.apiKey == "" {
		return "", "", nil, usage.Usage{}, errors.New("gemini api key not configured")
	}

	logs.Info(fmt.Sprintf("gemini request: model=%s country=%s input=%q", c.model, country, prompt))

	var system *content
	if hint := location.Hint(country); hint != "" {
		system = &content{Parts: []part{{Text: hint}}}
	}

	reqBody, err := json.Marshal(generateRequest{
		SystemInstruction: system,
		Contents:          []content{{Parts: []part{{Text: prompt}}}},
		Tools:             []tool{{}},
	})
	if err != nil {
		return "", "", nil, usage.Usage{}, err
	}

	url := fmt.Sprintf(apiURLTemplate, c.model, c.apiKey)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return "", "", nil, usage.Usage{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", nil, usage.Usage{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", nil, usage.Usage{}, err
	}

	var parsed generateResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", "", nil, usage.Usage{}, err
	}

	if resp.StatusCode != http.StatusOK {
		if parsed.Error != nil {
			return "", "", nil, usage.Usage{}, fmt.Errorf("gemini error: %s", parsed.Error.Message)
		}
		return "", "", nil, usage.Usage{}, fmt.Errorf("gemini request failed with status %d", resp.StatusCode)
	}

	if len(parsed.Candidates) == 0 {
		return "", "", nil, usage.Usage{}, errors.New("gemini returned no candidates")
	}

	cand := parsed.Candidates[0]
	var textParts []string
	for _, p := range cand.Content.Parts {
		if p.Text != "" {
			textParts = append(textParts, p.Text)
		}
	}
	response = strings.Join(textParts, "")
	if response == "" {
		return "", "", nil, usage.Usage{}, errors.New("gemini returned no message content")
	}

	if cand.GroundingMetadata != nil {
		for _, gc := range cand.GroundingMetadata.GroundingChunks {
			if gc.Web.Uri == "" {
				continue
			}
			// gc.Web.Uri is a single-use vertexaisearch.cloud.google.com
			// redirect token, not a stable real URL — resolving it here is
			// required for citation_frequency tracking to work at all
			// (re-citing the same real source across runs must produce the
			// same URL to upsert onto, not a fresh row every time).
			citations = append(citations, citation.Citation{URL: c.resolveRedirect(gc.Web.Uri), Title: gc.Web.Title})
		}
	}

	tokenUsage = usage.Usage{InputTokens: parsed.UsageMetadata.PromptTokenCount, OutputTokens: parsed.UsageMetadata.CandidatesTokenCount}
	logs.Info(fmt.Sprintf("gemini response: model=%s citations=%d tokens=%d/%d text=%q", c.model, len(citations), tokenUsage.InputTokens, tokenUsage.OutputTokens, response))

	return response, c.model, citations, tokenUsage, nil
}

// resolveRedirect follows a vertexaisearch.cloud.google.com redirect token to
// the real destination URL. Best-effort: if resolution fails for any reason
// (network error, timeout), falls back to the original redirect URL rather
// than dropping the citation — a working-but-opaque link beats none.
func (c *Client) resolveRedirect(redirectURL string) string {
	req, err := http.NewRequest(http.MethodHead, redirectURL, nil)
	if err != nil {
		return redirectURL
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return redirectURL
	}
	defer resp.Body.Close()
	if resp.Request != nil && resp.Request.URL != nil {
		return resp.Request.URL.String()
	}
	return redirectURL
}
