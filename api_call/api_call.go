package api_call

import (
	"ai-chat/database/structures"
	"fmt"
	bert "github.com/go-skynet/go-bert.cpp"
	"strings"
)

const (
	threads = 10
)

func OpenAIApiCall(data structures.SessionData, fileName string) (string, error) {

	return "The issue with the provided function is that it does not properly handle file messages that might come after a user's message and before an assistant's message. Additionally, the `fileName` variable is not being reset for each user message, which could lead to incorrect data being carried over from previous iterations.\n\nTo fix the problem, you should ensure that the `fileName` is reset to an empty string at the start of each iteration for a user message. Moreover, the loop that checks for the assistant's response should correctly handle file messages by continuing the search for the assistant's response, instead of breaking the loop when encountering a file. Here's the corrected version of the function:\n\n```javascript\nfunction convertMessages(messages) {\n    const result = [];\n    \n    messages.forEach((message, index) => {\n        if (message.role === \"user\") {\n            let request = message.content;\n            let response = \"\";\n            let fileName = \"\";\n\n            // Find the next 'assistant' message and possibly a file message\n            for (let i = index + 1; i < messages.length; i++) {\n                if (messages[i].role === \"assistant\") {\n                    response = messages[i].content;\n                    break;\n                } else if (messages[i].role === \"file\") {\n                    fileName = messages[i].content;\n                    // Continue searching for the assistant's response after the file\n                }\n            }\n\n            // Push the request-response pair into result\n            result.push({ request, response, loading: false, fileName: fileName });\n        }\n    });\n\n    return result;\n}\n```\n\nThis adjustment ensures that:\n1. `fileName` is set to the most recent file message that appears immediately after the user's request and before the assistant's response.\n2. If a file message is present, the loop continues to search for the assistant's response instead of stopping. This is crucial for handling scenarios where a file might be referenced in the interaction between the user and the assistant.", nil
}

func ApiEmbedding(input string, model *bert.Bert) (string, error) {
	// Preprocess the input (trim whitespace, etc.)
	input = strings.TrimSpace(input)

	// Generate embeddings
	embeddings, err := model.Embeddings(input, bert.SetThreads(threads))
	if err != nil {
		return "", fmt.Errorf("error generating embeddings: %w", err)
	}

	// Convert the embeddings to a string representation
	embeddingStr := fmt.Sprintf("%v", embeddings)

	fmt.Printf("\n\nEmbedding: %s\n\n", embeddingStr)

	return embeddingStr, nil
}
