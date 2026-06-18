package tool

import (
	"context"
	"encoding/json"
	"fmt"
)

type Executor interface {
	Execute(ctx context.Context, toolType string, config json.RawMessage, arguments string) (string, error)
}

type executor struct {
	coze *cozeExecutor
	n8n  *n8nExecutor
}

func NewExecutor(cozeBaseURL string, n8nDefaultTimeout int) Executor {
	return &executor{
		coze: &cozeExecutor{baseURL: cozeBaseURL},
		n8n:  &n8nExecutor{defaultTimeout: n8nDefaultTimeout},
	}
}

func (e *executor) Execute(ctx context.Context, toolType string, config json.RawMessage, arguments string) (string, error) {
	switch toolType {
	case "coze":
		return e.coze.execute(ctx, config, arguments)
	case "n8n":
		return e.n8n.execute(ctx, config, arguments)
	default:
		return "", fmt.Errorf("unsupported tool type: %s", toolType)
	}
}
