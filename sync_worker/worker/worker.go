package worker

import (
	"ai-chat/database/initialize"
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"os"
)

type StreamDataBase struct {
	Cache *redis.Client
}

func GetStreamDataBase() *StreamDataBase {
	return &StreamDataBase{
		Cache: initialize.InitRedis(),
	}
}

func (dataBase *StreamDataBase) AddToStream(ctx context.Context, userId string, sessionId string, modelId string, sessionPrompt string, chats string, chatsSummary string, sessionName string, isNew bool, balance float64) error {
	var isNewStr string
	if isNew {
		isNewStr = "new"
	} else {
		isNewStr = "old"
	}

	err := dataBase.Cache.XAdd(ctx, &redis.XAddArgs{
		Stream: os.Getenv("REDIS_STREAM"),
		MaxLen: 0,
		ID:     "",
		Values: []string{"userId", userId, "sessionId", sessionId, "sessionPrompt", sessionPrompt, "modelId", modelId, "chats", chats, "chatsSummary", chatsSummary, "sessionName", sessionName, "isNew", isNewStr, "balance", fmt.Sprintf("%f", balance)},
	}).Err()
	if err != nil {
		return err
	}

	return nil
}
