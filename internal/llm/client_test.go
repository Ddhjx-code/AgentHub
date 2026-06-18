package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChatCompletion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}

		var req ChatRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Model != "test-model" {
			t.Fatalf("expected model test-model, got %s", req.Model)
		}

		resp := ChatResponse{
			Choices: []Choice{
				{
					Message:      Message{Role: "assistant", Content: "Hello!"},
					FinishReason: "stop",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient()
	resp, err := c.ChatCompletion(context.Background(), server.URL, "test-key", ChatRequest{
		Model:    "test-model",
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	if err != nil {
		t.Fatalf("chat completion: %v", err)
	}
	if len(resp.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(resp.Choices))
	}
	if resp.Choices[0].Message.Content != "Hello!" {
		t.Fatalf("expected Hello!, got %s", resp.Choices[0].Message.Content)
	}
}

func TestChatCompletionWithToolCalls(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ChatRequest
		json.NewDecoder(r.Body).Decode(&req)

		hasToolResult := false
		for _, m := range req.Messages {
			if m.Role == "tool" {
				hasToolResult = true
				break
			}
		}

		var resp ChatResponse
		if !hasToolResult {
			resp = ChatResponse{
				Choices: []Choice{
					{
						Message: Message{
							Role: "assistant",
							ToolCalls: []ToolCall{
								{
									ID:   "call_1",
									Type: "function",
									Function: ToolCallFunction{
										Name:      "search",
										Arguments: `{"query":"test"}`,
									},
								},
							},
						},
						FinishReason: "tool_calls",
					},
				},
			}
		} else {
			resp = ChatResponse{
				Choices: []Choice{
					{
						Message:      Message{Role: "assistant", Content: "Found results."},
						FinishReason: "stop",
					},
				},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient()

	resp, err := c.ChatCompletion(context.Background(), server.URL, "test-key", ChatRequest{
		Model:    "test-model",
		Messages: []Message{{Role: "user", Content: "search for test"}},
		Tools: []Tool{
			{
				Type: "function",
				Function: ToolFunction{
					Name:        "search",
					Description: "Search the web",
					Parameters:  json.RawMessage(`{"type":"object","properties":{"query":{"type":"string"}}}`),
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	if resp.Choices[0].FinishReason != "tool_calls" {
		t.Fatalf("expected tool_calls, got %s", resp.Choices[0].FinishReason)
	}
	if len(resp.Choices[0].Message.ToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(resp.Choices[0].Message.ToolCalls))
	}

	resp2, err := c.ChatCompletion(context.Background(), server.URL, "test-key", ChatRequest{
		Model: "test-model",
		Messages: []Message{
			{Role: "user", Content: "search for test"},
			{Role: "assistant", ToolCalls: resp.Choices[0].Message.ToolCalls},
			{Role: "tool", ToolCallID: "call_1", Content: "result data"},
		},
	})
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if resp2.Choices[0].Message.Content != "Found results." {
		t.Fatalf("expected Found results., got %s", resp2.Choices[0].Message.Content)
	}
}

func TestChatCompletionAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal server error"}`))
	}))
	defer server.Close()

	c := NewClient()
	_, err := c.ChatCompletion(context.Background(), server.URL, "test-key", ChatRequest{
		Model:    "test-model",
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}
