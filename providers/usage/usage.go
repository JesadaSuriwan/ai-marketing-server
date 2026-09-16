// Package usage carries real token counts back from each AI provider call,
// and estimates a USD cost from them via a hand-maintained price table.
// Token counts always come straight from the provider's own API response —
// never estimated. Cost is a best-effort multiplication against list prices
// that WILL drift out of date; treat it as a directional estimate, not an
// invoice, and re-check pricing.md-style provider pages periodically.
package usage

import "strings"

type Usage struct {
	InputTokens  int
	OutputTokens int
}

// ModelPrice is USD per 1,000,000 tokens.
type ModelPrice struct {
	InputPerMillion  float64
	OutputPerMillion float64
}

// prices are list prices as of when this was written. They are NOT fetched
// live from anywhere — providers don't expose pricing via API — so this
// table needs manual upkeep whenever a provider changes rates. Matched by
// exact model string first, then by substring for model families whose
// resolved name varies at runtime (e.g. Gemini's "-latest" alias).
var prices = map[string]ModelPrice{
	"gpt-4.1":         {InputPerMillion: 2.00, OutputPerMillion: 8.00},
	"gpt-4.1-mini":    {InputPerMillion: 0.40, OutputPerMillion: 1.60},
	"gpt-4o":          {InputPerMillion: 2.50, OutputPerMillion: 10.00},
	"claude-opus-4-8": {InputPerMillion: 15.00, OutputPerMillion: 75.00},
	"claude-sonnet-4": {InputPerMillion: 3.00, OutputPerMillion: 15.00},
	"sonar":           {InputPerMillion: 1.00, OutputPerMillion: 1.00},
	"sonar-pro":       {InputPerMillion: 3.00, OutputPerMillion: 15.00},
}

// familyPrices is checked when no exact key matches — substring, checked in
// order, first match wins. Needed because Gemini's "-latest" alias resolves
// to a dated model name at runtime that we can't hardcode exactly.
var familyPrices = []struct {
	contains string
	price    ModelPrice
}{
	{contains: "gemini-flash", price: ModelPrice{InputPerMillion: 0.075, OutputPerMillion: 0.30}},
	{contains: "gemini-pro", price: ModelPrice{InputPerMillion: 1.25, OutputPerMillion: 5.00}},
	{contains: "gemini", price: ModelPrice{InputPerMillion: 0.075, OutputPerMillion: 0.30}},
	{contains: "opus", price: ModelPrice{InputPerMillion: 15.00, OutputPerMillion: 75.00}},
	{contains: "sonnet", price: ModelPrice{InputPerMillion: 3.00, OutputPerMillion: 15.00}},
	{contains: "haiku", price: ModelPrice{InputPerMillion: 0.80, OutputPerMillion: 4.00}},
	{contains: "sonar-pro", price: ModelPrice{InputPerMillion: 3.00, OutputPerMillion: 15.00}},
	{contains: "sonar", price: ModelPrice{InputPerMillion: 1.00, OutputPerMillion: 1.00}},
	{contains: "gpt-4.1-mini", price: ModelPrice{InputPerMillion: 0.40, OutputPerMillion: 1.60}},
	{contains: "gpt-4.1", price: ModelPrice{InputPerMillion: 2.00, OutputPerMillion: 8.00}},
	{contains: "gpt-4o", price: ModelPrice{InputPerMillion: 2.50, OutputPerMillion: 10.00}},
}

// EstimateCost returns 0 for a model with no known price rather than
// guessing — a $0 row with real token counts still visible is an honest
// signal that this table needs a new entry, not a fabricated number.
func EstimateCost(model string, u Usage) float64 {
	p, ok := prices[model]
	if !ok {
		lower := strings.ToLower(model)
		for _, f := range familyPrices {
			if strings.Contains(lower, f.contains) {
				p = f.price
				ok = true
				break
			}
		}
	}
	if !ok {
		return 0
	}
	return float64(u.InputTokens)/1_000_000*p.InputPerMillion + float64(u.OutputTokens)/1_000_000*p.OutputPerMillion
}
