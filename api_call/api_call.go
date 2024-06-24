package api_call

import (
	"ai-chat/database/structures"
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"os"
	"strings"
	"time"
)

const (
	timeout = 30 * time.Second
)

func OpenAIApiCall(data structures.SessionData, fileName string) (string, error) {
	return "The issue with the provided function is that it does not properly handle file messages that might come after a user's message and before an assistant's message. Additionally, the `fileName` variable is not being reset for each user message, which could lead to incorrect data being carried over from previous iterations.\n\nTo fix the problem, you should ensure that the `fileName` is reset to an empty string at the start of each iteration for a user message. Moreover, the loop that checks for the assistant's response should correctly handle file messages by continuing the search for the assistant's response, instead of breaking the loop when encountering a file. Here's the corrected version of the function:\n\n```javascript\nfunction convertMessages(messages) {\n    const result = [];\n    \n    messages.forEach((message, index) => {\n        if (message.role === \"user\") {\n            let request = message.content;\n            let response = \"\";\n            let fileName = \"\";\n\n            // Find the next 'assistant' message and possibly a file message\n            for (let i = index + 1; i < messages.length; i++) {\n                if (messages[i].role === \"assistant\") {\n                    response = messages[i].content;\n                    break;\n                } else if (messages[i].role === \"file\") {\n                    fileName = messages[i].content;\n                    // Continue searching for the assistant's response after the file\n                }\n            }\n\n            // Push the request-response pair into result\n            result.push({ request, response, loading: false, fileName: fileName });\n        }\n    });\n\n    return result;\n}\n```\n\nThis adjustment ensures that:\n1. `fileName` is set to the most recent file message that appears immediately after the user's request and before the assistant's response.\n2. If a file message is present, the loop continues to search for the assistant's response instead of stopping. This is crucial for handling scenarios where a file might be referenced in the interaction between the user and the assistant.", nil
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
