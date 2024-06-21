package services

import (
	"ai-chat/database/initialize"
	"ai-chat/database/structures"
	"ai-chat/sync_worker/worker"
	"ai-chat/utils/model_data"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/RediSearch/redisearch-go/v2/redisearch"
	bert "github.com/go-skynet/go-bert.cpp"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"sort"
	"strconv"
	"strings"
)

type Database struct {
	Db     *sqlx.DB
	Cache  *redis.Client
	Stream *worker.StreamDataBase
	Vector *redisearch.Client
	Model  *bert.Bert
}

func GetDataBase() *Database {
	return &Database{
		Db:     initialize.InitPostgres(),
		Cache:  initialize.InitRedis(),
		Stream: worker.GetStreamDataBase(),
		Vector: initialize.InitRedisChatVector(),
		Model:  initialize.InitModel(),
	}
}

func (dataBase *Database) DeleteSession(userId string, sessionId string) (structures.SessionDeleteResponse, error) {
	key := fmt.Sprintf("user:%s:session:%s", userId, sessionId)
	_, err := dataBase.Cache.Del(context.Background(), key).Result()
	if err != nil {
		return structures.SessionDeleteResponse{}, err
	}

	tx := dataBase.Db.MustBegin()
	query := `DELETE FROM Session_Details WHERE Session_Id = $1`
	rows := tx.MustExec(query, sessionId)

	affected, err := rows.RowsAffected()
	if err != nil {
		return structures.SessionDeleteResponse{}, errors.New("unable To Get Affected Rows")
	}

	if affected <= 0 {
		fmt.Println("User Not Exist")
		return structures.SessionDeleteResponse{}, errors.New("no Rows Get Affected")
	}

	err = tx.Commit()
	if err != nil {
		return structures.SessionDeleteResponse{}, errors.New("unable To Commit")
	}

	return structures.SessionDeleteResponse{
		UserId: userId,
	}, nil
}

func (dataBase *Database) GetUserDetails(userId string) (*structures.UserDataResponse, error) {
	var data structures.UserDataResponse
	err := dataBase.Db.Get(&data, "select user_id, username from user_data where user_id=$1", userId)

	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (dataBase *Database) CreateNewSession(userId string, prompt string, modelId int) (string, error) {
	// Generate a new UUID for the session
	sessionId := uuid.New().String()

	// Update Redis with the new session information
	key := fmt.Sprintf("user:%s:session:%s", userId, sessionId)
	sessionData := map[string]interface{}{
		"model_id":       modelId,
		"session_prompt": prompt,
		"chats":          "[]", // Start with an empty chats array
	}

	_, err := dataBase.Cache.HSet(context.Background(), key, sessionData).Result()
	if err != nil {
		return "", err
	}

	return sessionId, nil
}

func (dataBase *Database) SetSessionValues(userId string, prompt string, modelId int, sessionId string, chats []structures.Chat) error {
	chatsJSON, err := json.Marshal(chats)
	if err != nil {
		return err
	}

	// Update Redis with the new session information
	key := fmt.Sprintf("user:%s:session:%s", userId, sessionId)
	sessionData := map[string]interface{}{
		"model_id":       modelId,
		"session_prompt": prompt,
		"chats":          chatsJSON, // Start with an empty chats array
	}

	_, err = dataBase.Cache.HSet(context.Background(), key, sessionData).Result()
	if err != nil {
		return err
	}
	return nil
}

func (dataBase *Database) GetUserSessionData(userId string, sessionId string) (structures.SessionData, error) {
	// Construct the key to access the session data in Redis
	key := fmt.Sprintf("user:%s:session:%s", userId, sessionId)

	// Retrieve session data from Redis
	values, err := dataBase.Cache.HGetAll(context.Background(), key).Result()
	if err != nil {
		return structures.SessionData{}, fmt.Errorf("error retrieving session data from Redis: %w", err)
	}

	if len(values) == 0 {
		return structures.SessionData{}, errors.New("no session data found")
	}

	// Parse the chats from the retrieved data
	var chats []structures.Chat
	if err := json.Unmarshal([]byte(values["chats"]), &chats); err != nil {
		return structures.SessionData{}, fmt.Errorf("error parsing chats data: %w", err)
	}

	modelId, err := strconv.Atoi(values["model_id"])
	if err != nil {
		return structures.SessionData{}, fmt.Errorf("error parsing modelId: %w", err)
	}

	// Construct the session data structure
	sessionData := structures.SessionData{
		SessionId: sessionId,
		ModelId:   modelId,
		Prompt:    values["session_prompt"],
		Chats:     chats,
	}

	return sessionData, nil
}

func (dataBase *Database) CheckModelAccess(userId string, modelId int) error {
	// Construct the key to access the user's data in Redis
	userKey := fmt.Sprintf("user:%s", userId)

	// Retrieve the models data from Redis
	modelsJSON, err := dataBase.Cache.HGet(context.Background(), userKey, "models").Result()
	if err != nil {
		return err
	}

	// Parse the JSON array of model IDs
	trimmed := strings.Trim(modelsJSON, "{}")
	// Split the string by comma
	parts := strings.Split(trimmed, ",")

	// Check if the user has access to the specified modelId
	for _, part := range parts {
		if part == strconv.Itoa(modelId) {
			return nil
		}
	}
	return errors.New("access denied: user does not have access to this model")
}

func (dataBase *Database) GetSessionsByUserId(userId string) (structures.SessionListResponse, error) {
	// Use parameterized query to prevent SQL injection
	query := "SELECT session_id, session_name FROM session_details WHERE user_id=$1 ORDER BY session_name DESC"

	rows, err := dataBase.Db.Query(query, userId)
	if err != nil {
		return structures.SessionListResponse{}, err
	}
	defer rows.Close()

	var sessionInfo []structures.SessionInfo
	for rows.Next() {
		var sessionIDTemp, sessionNameTemp sql.NullString
		if err := rows.Scan(&sessionIDTemp, &sessionNameTemp); err != nil {
			return structures.SessionListResponse{}, err
		}

		// Check if SQL values are not null before appending
		sessionId := ""
		if sessionIDTemp.Valid {
			sessionId = sessionIDTemp.String
		}
		sessionName := ""
		if sessionNameTemp.Valid {
			sessionName = sessionNameTemp.String
		}

		sessionInfo = append(sessionInfo, structures.SessionInfo{
			SessionId:   sessionId,
			SessionName: sessionName,
		})
	}

	return structures.SessionListResponse{
		UserId:  userId,
		Session: sessionInfo,
	}, nil
}

func (dataBase *Database) GetUserSessionChat(userId string, sessionId string) (string, error) {
	// Construct the key to access the session data in Redis
	key := fmt.Sprintf("user:%s:session:%s", userId, sessionId)

	// Retrieve session data from Redis
	values, err := dataBase.Cache.HGet(context.Background(), key, "chats").Result()
	if err != nil {
		return "", fmt.Errorf("error retrieving session data from Redis: %w", err)
	}

	if len(values) == 0 {
		return "", errors.New("no session data found")
	}

	return values, nil
}

func (dataBase *Database) GetAIModel() (structures.AIModelsResponse, error) {
	modelNames := make([]string, 0, len(model_data.ModelNameMapping))
	for name := range model_data.ModelNameMapping {
		modelNames = append(modelNames, name)
	}
	sort.Strings(modelNames) // Sorting the slice to ensure the order is consistent
	return structures.AIModelsResponse{Models: modelNames}, nil

}
