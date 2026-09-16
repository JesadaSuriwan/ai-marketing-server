package resend

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const apiURL = "https://api.resend.com/emails"

type Client struct {
	apiKey string
	from   string
}

func NewClient(apiKey, from string) *Client {
	return &Client{apiKey: apiKey, from: from}
}

type sendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Html    string   `json:"html"`
}

// Send delivers a single HTML email via the Resend REST API.
func (c *Client) Send(to, subject, html string) error {
	if c.apiKey == "" {
		return errors.New("resend api key not configured")
	}

	body, err := json.Marshal(sendRequest{From: c.from, To: []string{to}, Subject: subject, Html: html})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend: status %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}
