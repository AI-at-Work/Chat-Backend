package helper_functions

import (
	"ai-chat/database/structures"
	"ai-chat/utils/model_data"
	"fmt"
	"github.com/pkoukk/tiktoken-go"
	"log"
	"math"
	"strings"
)

func EstimateOpenAIAPICost(model string, numTokensInput, numTokensOutput int) (float64, error) {
	pricing, ok := model_data.ModelPricing[model]
	if !ok {
		return 0, fmt.Errorf("unknown model: %s", model)
	}

	inputCost := float64(numTokensInput/1000) * pricing.Input
	outputCost := float64(numTokensOutput/1000) * pricing.Output
	totalCost := inputCost + outputCost

	return math.Round(totalCost), nil
}

// Function to simulate token counting.
func countTokens(content string, model string) (int, error) {
	tkm, err := tiktoken.EncodingForModel(model)
	if err != nil {
		err = fmt.Errorf("encoding for model: %v", err)
		log.Println(err)
		return 0, err
	}

	var tokensPerMessage = 4 // every message follows <|start|>{role/name}\n{content}<|end|>\n
	var tokensPerName = -1   // if there's a name, the role is omitted

	return len(tkm.Encode(content, nil, nil)) + tokensPerMessage + tokensPerName + 4, nil
}

// Ensure the sessionData.Chats fit within the token limit.
func LimitTokenSize(sessionData *structures.SessionData, maxTokens int) error {
	// To make sure proper token encoding get applied
	model := model_data.ModelName(sessionData.ModelId)
	if strings.Contains(model, "gpt-4o") {
		model = "gpt-4o"
	} else if strings.Contains(model, "gpt-4") || strings.Contains(model, "gpt-3") {
		model = "gpt-3.5-turbo"
	} else if strings.Contains(model, "text-davinci-003") || strings.Contains(model, "text-davinci-002") {
		model = "text-davinci-002"
	} else {
		model = "text-davinci-001"
	}

	totalTokens := 0
	startIndex := len(sessionData.Chats)

	// Count tokens from the end to start and find the index where tokens exceed the limit.
	for i := len(sessionData.Chats) - 1; i >= 0; i-- {
		chat := sessionData.Chats[i]

		var contentTokens, roleTokens int
		var err error
		if contentTokens, err = countTokens(chat.Content+" content", model); err != nil {
			return err
		}

		if roleTokens, err = countTokens(chat.Role+" role", model); err != nil {
			return err
		}

		tokens := contentTokens + roleTokens // Sum tokens of content and role.

		if totalTokens+tokens > maxTokens {
			break
		}
		totalTokens += tokens
		startIndex = i
	}

	// Slice the array to contain only the latest fitting entries.
	sessionData.Chats = sessionData.Chats[startIndex:]
	return nil
}

func TruncateText(s string, max int) string {
	if max > len(s) {
		return s
	}
	return s[:strings.LastIndex(s[:max], " ")]
}
