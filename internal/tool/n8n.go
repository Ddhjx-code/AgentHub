package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type n8nConfig struct {
	WebhookURL  string `json:"webhook_url"`
	AuthType    string `json:"auth_type"`
	AuthToken   string `json:"auth_token"`
	Timeout     int    `json:"timeout"`
	PayloadTmpl string `json:"payload_tmpl"`
}

type n8nExecutor struct {
	defaultTimeout int
}

func (e *n8nExecutor) execute(ctx context.Context, rawConfig json.RawMessage, arguments string) (string, error) {
	var cfg n8nConfig
	if err := json.Unmarshal(rawConfig, &cfg); err != nil {
		return "", fmt.Errorf("parse n8n config: %w", err)
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = e.defaultTimeout
	}

	var payload string
	if cfg.PayloadTmpl != "" {
		payload = strings.ReplaceAll(cfg.PayloadTmpl, "{{.input}}", arguments)
	} else {
		payload = arguments
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.WebhookURL,
		bytes.NewReader([]byte(payload)))
	if err != nil {
		return "", fmt.Errorf("create n8n request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	switch cfg.AuthType {
	case "bearer":
		httpReq.Header.Set("Authorization", "Bearer "+cfg.AuthToken)
	case "header":
		httpReq.Header.Set("Authorization", cfg.AuthToken)
	}

	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("n8n http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read n8n response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("n8n webhook error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return string(respBody), nil
}
