package api_call

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type LLM struct {
	OpenAIAPIKey string
}

type GenerateResponse struct {
	Text         string
	InputTokens  int
	OutputTokens int
}

func NewLLM(openAIAPIKey string) *LLM {
	return &LLM{
		OpenAIAPIKey: openAIAPIKey,
	}
}

func (l *LLM) Generate(provider, model, userPrompt, systemPrompt string) (*GenerateResponse, error) {
	switch strings.ToLower(provider) {
	case "ollama":
		return l.generateOllama(model, userPrompt, systemPrompt)
	case "openai":
		return l.generateOpenAI(model, userPrompt, systemPrompt)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

func (l *LLM) generateOllama(model, userPrompt, systemPrompt string) (*GenerateResponse, error) {
	url := fmt.Sprintf("http://ollama:%s/api/generate", os.Getenv("OLLAMA_PORT"))
	data := map[string]interface{}{
		"model":  model,
		"prompt": userPrompt,
		"system": systemPrompt,
		"stream": false,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		var err net.Error
		if errors.As(err, &err) && err.Timeout() {
			return nil, fmt.Errorf("request to Ollama API timed out. The server might be overloaded or not responding")
		}
		return nil, fmt.Errorf("failed to connect to Ollama API. Is the server running? Error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama API returned non-OK status: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON response: %w", err)
	}

	return &GenerateResponse{
		Text:         result["response"].(string),
		InputTokens:  int(result["prompt_eval_count"].(float64)),
		OutputTokens: int(result["eval_count"].(float64)),
	}, nil
}

func (l *LLM) generateOpenAI(model, userPrompt, systemPrompt string) (*GenerateResponse, error) {
	if l.OpenAIAPIKey == "" {
		return nil, fmt.Errorf("API key is required for OpenAI")
	}

	client := openai.NewClient(l.OpenAIAPIKey)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	chatReq := openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}
	resp, err := client.CreateChatCompletion(ctx, chatReq)

	if err != nil {
		return nil, fmt.Errorf("error creating summary: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no summary found for the input")
	}

	return &GenerateResponse{
		Text:         resp.Choices[0].Message.Content,
		InputTokens:  resp.Usage.PromptTokens,
		OutputTokens: resp.Usage.CompletionTokens,
	}, nil
}
