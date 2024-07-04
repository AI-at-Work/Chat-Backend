package api_call

import (
	"ai-chat/database/structures"
	"ai-chat/utils/response_code/error_code"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "ai-chat/pb"
	"github.com/sashabaranov/go-openai"
	"os"
	"strings"
	"time"
)

const (
	timeout = 30 * time.Second
)

type AIClient struct {
	client pb.AIServiceClient
	conn   *grpc.ClientConn
}

func InitAIClient() *AIClient {
	address := os.Getenv("AI_SERVER_ADDRESS")
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	return &AIClient{
		client: pb.NewAIServiceClient(conn),
		conn:   conn,
	}
}

func (c *AIClient) Close() {
	c.conn.Close()
}

func (c *AIClient) AIApiCall(userId, sessionId, chat string, fileName []string, sessionPrompt string, chatHistory []structures.Chat, chatSummary, modelName string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer func() {
		cancel()
	}()

	chatHistoryStr, err := json.Marshal(chatHistory)
	if err != nil {
		return "", errors.New(string(error_code.Error(error_code.ErrorCodeJSONMarshal)))
	}

	r, err := c.client.Process(ctx, &pb.Request{
		UserId:        userId,
		SessionId:     sessionId,
		ChatMessage:   chat,
		FileName:      fileName,
		ModelName:     modelName,
		SessionPrompt: sessionPrompt,
		ChatSummary:   chatSummary,
		ChatHistory:   string(chatHistoryStr),
		Timestamp:     timestamppb.Now(),
	})
	if err != nil {
		fmt.Println("API ERR: ", err)
		return "", err
	}
	return r.GetResponseText(), nil
}

func (c *AIClient) ApiSummary(summary, chats, model string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")

	client := openai.NewClient(apiKey)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	prompt := GetSummaryPrompt(summary, chats)

	// Determine the type of completion based on the model
	chatReq := openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{Role: "user", Content: prompt},
		},
	}
	resp, err := client.CreateChatCompletion(ctx, chatReq)

	if err != nil {
		return "", fmt.Errorf("error creating summary: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no summary found for the input")
	}

	newSummary := resp.Choices[0].Message.Content
	return newSummary, nil
}

func ApiEmbedding(input string) ([]float32, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("input is empty")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")

	client := openai.NewClient(apiKey)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req := openai.EmbeddingRequest{
		Input: []string{input},
		Model: openai.AdaEmbeddingV2,
	}

	resp, err := client.CreateEmbeddings(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error creating embeddings: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embeddings found for the input")
	}

	embedding := resp.Data[0].Embedding
	return embedding, nil
}
