// Package citation defines the shared, provider-agnostic shape of a
// search-grounded source URL, so multiple AI providers (openai, gemini, ...)
// can satisfy the same pkg/promptrun/service.AIProvider interface without
// importing each other.
package citation

// Citation is a real, search-grounded source URL an AI provider actually cited.
type Citation struct {
	URL   string
	Title string
}
