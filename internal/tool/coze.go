package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type cozeConfig struct {
	WorkflowID  string `json:"workflow_id"`
	APIKey      string `json:"api_key"`
	Region      string `json:"region"`
	InputField  string `json:"input_field"`
	OutputField string `json:"output_field"`
}

type cozeExecutor struct {
	baseURL string
}

func (e *cozeExecutor) execute(ctx context.Context, rawConfig json.RawMessage, arguments string) (string, error) {
	var cfg cozeConfig
	if err := json.Unmarshal(rawConfig, &cfg); err != nil {
		return "", fmt.Errorf("parse coze config: %w", err)
	}

	inputField := cfg.InputField
	if inputField == "" {
		inputField = "input"
	}
	outputField := cfg.OutputField
	if outputField == "" {
		outputField = "output"
	}

	var args map[string]interface{}
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		args = map[string]interface{}{inputField: arguments}
	}

	if _, ok := args[inputField]; !ok {
		args[inputField] = arguments
	}

	reqBody := map[string]interface{}{
		"workflow_id": cfg.WorkflowID,
		"parameters":  args,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal coze request: %w", err)
	}

	baseURL := e.baseURL
	if cfg.Region != "" && cfg.Region != "cn" {
		baseURL = "https://api.coze.com"
	}

	url := baseURL + "/v1/workflow/run"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create coze request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("coze http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read coze response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("coze API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return string(respBody), nil
	}

	if data, ok := result["data"].(string); ok {
		var dataObj map[string]interface{}
		if err := json.Unmarshal([]byte(data), &dataObj); err == nil {
			if output, ok := dataObj[outputField]; ok {
				outputBytes, _ := json.Marshal(output)
				return string(outputBytes), nil
			}
		}
		return data, nil
	}

	return string(respBody), nil
}
