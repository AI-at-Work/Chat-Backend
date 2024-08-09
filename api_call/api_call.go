package api_call

import (
	"ai-chat/database/structures"
	"ai-chat/utils/helper_functions"
	"ai-chat/utils/model_data"
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
	timeout = 10 * time.Minute
)

type AIClient struct {
	client pb.AIServiceClient
	conn   *grpc.ClientConn
	llm    *LLM
}

func InitAIClient() *AIClient {
	address := fmt.Sprintf("%s:%s", os.Getenv("AI_SERVER_HOST"), os.Getenv("AI_SERVER_PORT"))
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	return &AIClient{
		client: pb.NewAIServiceClient(conn),
		conn:   conn,
		llm:    NewLLM(os.Getenv("OPENAI_API_KEY")),
	}
}

func (c *AIClient) Close() {
	c.conn.Close()
}

func (c *AIClient) AIApiCall(userId, sessionId, chat string, fileName []string, sessionPrompt string, chatHistory []structures.Chat, chatSummary, modelName, modelProvider string, balance float64) (string, float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer func() {
		cancel()
	}()

	chatHistoryStr, err := json.Marshal(chatHistory)
	if err != nil {
		return "", 0, errors.New(string(error_code.Error(error_code.ErrorCodeJSONMarshal)))
	}

	r, err := c.client.Process(ctx, &pb.Request{
		UserId:        userId,
		SessionId:     sessionId,
		ChatMessage:   chat,
		FileName:      fileName,
		ModelName:     modelName,
		ModelProvider: modelProvider,
		SessionPrompt: sessionPrompt,
		ChatSummary:   chatSummary,
		ChatHistory:   string(chatHistoryStr),
		Balance:       float32(balance),
		Timestamp:     timestamppb.Now(),
	})
	if err != nil {
		fmt.Println("API ERR: ", err)
		return "", 0, err
	}
	return r.GetResponseText(), float64(r.GetCost()), nil
}

func (c *AIClient) ApiSummary(summary, chats, model string) (string, float64, error) {
	prompt := GetSummaryPrompt(summary, chats)

	resp, err := c.llm.Generate(model_data.GetModelProvider(model), model, prompt, "")
	if err != nil {
		return "", 0, fmt.Errorf("error while calling llm : %w", err)
	}

	cost, err := helper_functions.EstimateOpenAIAPICost(model, resp.InputTokens, resp.OutputTokens)
	if err != nil {
		return "", 0, fmt.Errorf("error while estimating cost: %w", err)
	}
	return resp.Text, cost, nil
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
