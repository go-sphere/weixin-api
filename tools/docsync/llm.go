package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// The LLM fallback covers the pages the deterministic parser cannot handle.
// It may only fill in endpoint facts (method, path, parameter tables); the
// operation ID and every Go identifier are still derived from the endpoint by
// the fixed rules in assignOperationIDs, so LLM randomness can never change
// the generated API surface.

// llmConfig holds the OpenAI-compatible endpoint settings from LLM.ENV.
type llmConfig struct {
	BaseURL string
	APIKey  string
	Model   string
}

// loadLLMConfig reads LLM.ENV from the repository root.
func loadLLMConfig(root string) (*llmConfig, error) {
	data, err := os.ReadFile(filepath.Join(root, "LLM.ENV"))
	if err != nil {
		return nil, fmt.Errorf("read LLM.ENV: %w", err)
	}
	cfg := &llmConfig{Model: "deepseek-chat"}
	for line := range strings.SplitSeq(string(data), "\n") {
		k, v, ok := strings.Cut(strings.TrimSpace(line), "=")
		if !ok {
			continue
		}
		switch k {
		case "OPENAI_API_BASE":
			cfg.BaseURL = strings.TrimSpace(v)
		case "OPENAI_API_KEY":
			cfg.APIKey = strings.TrimSpace(v)
		case "OPENAI_MODEL":
			cfg.Model = strings.TrimSpace(v)
		}
	}
	if cfg.APIKey == "" || cfg.BaseURL == "" {
		return nil, errors.New("LLM.ENV must define OPENAI_API_BASE and OPENAI_API_KEY")
	}
	return cfg, nil
}

// llmTarget is a page that needs LLM-assisted extraction.
type llmTarget struct {
	Page  *page
	URL   string
	Title string
	Body  string
}

// refineWithLLM re-extracts the pages whose deterministic parse produced no
// operation even though they look like API pages.
func refineWithLLM(ctx context.Context, pages map[string]*page, ops []*Operation) ([]*Operation, error) {
	root, err := repoRoot()
	if err != nil {
		return nil, err
	}
	cfg, err := loadLLMConfig(root)
	if err != nil {
		return nil, err
	}
	client := openai.NewClient(
		option.WithAPIKey(cfg.APIKey),
		option.WithBaseURL(cfg.BaseURL),
	)

	var targets []*llmTarget
	for _, url := range slices.Sorted(maps.Keys(pages)) {
		p := pages[url]
		if len(p.Ops) > 0 {
			continue
		}
		text := string(p.Body)
		if !strings.Contains(text, "调用方式") && !strings.Contains(text, "请求参数") {
			continue // catalogue pages are not API pages
		}
		targets = append(targets, &llmTarget{Page: p, URL: url, Title: p.Title, Body: text})
	}
	if len(targets) == 0 {
		fmt.Println("llm: no pages need fallback extraction")
		return ops, nil
	}
	fmt.Printf("llm: extracting %d page(s) the deterministic parser missed\n", len(targets))
	for _, t := range targets {
		op, err := extractWithLLM(ctx, client, cfg.Model, t)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", t.URL, err)
		}
		if op != nil {
			t.Page.Ops = append(t.Page.Ops, op)
			ops = append(ops, op)
		}
	}
	return ops, nil
}

// extractAPITool is the strict JSON schema the model must fill via tool call.
var extractAPITool = openai.ChatCompletionToolParam{
	Type: "function",
	Function: openai.FunctionDefinitionParam{
		Name:        "extract_api",
		Description: openai.String("Extract one WeChat server API definition from its documentation page text."),
		Parameters: openai.FunctionParameters{
			"type": "object",
			"properties": map[string]any{
				"method": map[string]any{"type": "string", "enum": []string{"GET", "POST", "PUT", "DELETE"}},
				"path":   map[string]any{"type": "string", "description": "Request path on api.weixin.qq.com, e.g. /cgi-bin/token"},
				"query":  fieldsSchema(),
				"body":   fieldsSchema(),
				"response": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type":                 "object",
						"properties":           fieldProperties(),
						"required":             []string{"name", "type", "required", "description"},
						"additionalProperties": false,
					},
				},
			},
			"required": []string{"method", "path", "query", "body", "response"},
		},
	},
}

func fieldsSchema() map[string]any {
	return map[string]any{
		"type": "array",
		"items": map[string]any{
			"type":                 "object",
			"properties":           fieldProperties(),
			"required":             []string{"name", "type", "required", "description"},
			"additionalProperties": false,
		},
	}
}

func fieldProperties() map[string]any {
	return map[string]any{
		"name":        map[string]any{"type": "string"},
		"type":        map[string]any{"type": "string", "enum": []string{"string", "number", "boolean", "array", "object"}},
		"required":    map[string]any{"type": "boolean"},
		"description": map[string]any{"type": "string"},
	}
}

// extractWithLLM asks the model to fill the extraction schema for one page.
func extractWithLLM(ctx context.Context, client openai.Client, model string, t *llmTarget) (*Operation, error) {
	prompt := fmt.Sprintf(
		"Below is the normalized text of one WeChat server API documentation page. "+
			"Extract the single API it documents. Keep parameter names exactly as written. "+
			"Translate nothing. Page title: %s\n\n%s", t.Title, truncate(t.Body, 24000))

	completion, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		Tools: []openai.ChatCompletionToolParam{extractAPITool},
		ToolChoice: openai.ChatCompletionToolChoiceOptionParamOfChatCompletionNamedToolChoice(
			openai.ChatCompletionNamedToolChoiceFunctionParam{Name: "extract_api"},
		),
		Temperature: openai.Float(0),
	})
	if err != nil {
		return nil, err
	}
	if len(completion.Choices) == 0 {
		return nil, errors.New("empty completion")
	}
	for _, tc := range completion.Choices[0].Message.ToolCalls {
		if tc.Function.Name != "extract_api" {
			continue
		}
		return operationFromToolArgs(t.URL, t.Title, []byte(tc.Function.Arguments))
	}
	return nil, errors.New("model did not call extract_api")
}

// operationFromToolArgs validates the tool call payload and builds an
// Operation. Identifiers still come from assignOperationIDs later.
func operationFromToolArgs(pageURL, title string, args []byte) (*Operation, error) {
	var payload struct {
		Method   string  `json:"method"`
		Path     string  `json:"path"`
		Query    []Field `json:"query"`
		Body     []Field `json:"body"`
		Response []Field `json:"response"`
	}
	dec := json.NewDecoder(bytes.NewReader(args))
	if err := dec.Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode tool args: %w", err)
	}
	payload.Method = strings.ToUpper(payload.Method)
	payload.Path = strings.TrimSuffix(payload.Path, "/")
	if payload.Method == "" || !strings.HasPrefix(payload.Path, "/") {
		return nil, errors.New("tool args missing method/path")
	}
	for i := range payload.Query {
		payload.Query[i].Type = normalizeFieldType(payload.Query[i].Type, payload.Query[i].Desc)
	}
	for i := range payload.Body {
		payload.Body[i].Type = normalizeType(payload.Body[i].Type)
	}
	for i := range payload.Response {
		payload.Response[i].Type = normalizeFieldType(payload.Response[i].Type, payload.Response[i].Desc)
	}
	return &Operation{
		Tree:     treeOf(pageURL),
		Method:   payload.Method,
		Path:     payload.Path,
		Summary:  title,
		Page:     pageURL,
		Query:    payload.Query,
		Body:     payload.Body,
		Response: payload.Response,
	}, nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "\n...[truncated]"
}
