package services

import (
	"ai-chat/api_call"
	"ai-chat/database/initialize"
	"ai-chat/database/structures"
	"ai-chat/sync_worker/worker"
	"ai-chat/utils/model_data"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"sort"
	"strconv"
	"strings"
)

type Database struct {
	Db        *sqlx.DB
	Cache     *redis.Client
	Stream    *worker.StreamDataBase
	AIService *api_call.AIClient
}

func GetDataBase() *Database {
	return &Database{
		Db:        initialize.InitPostgres(),
		Cache:     initialize.InitRedis(),
		Stream:    worker.GetStreamDataBase(),
		AIService: api_call.InitAIClient(),
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

func (dataBase *Database) CreateNewSession(userId string, sessionData structures.SessionData) (string, error) {
	// Generate a new UUID for the session
	sessionId := uuid.New().String()

	chatsJSON, err := json.Marshal(sessionData.Chats)
	if err != nil {
		return "", err
	}

	fileNameJSON, err := json.Marshal(sessionData.FileName)
	if err != nil {
		return "", err
	}

	// Update Redis with the new session information
	key := fmt.Sprintf("user:%s:session:%s", userId, sessionId)
	data := map[string]interface{}{
		"model_id":       sessionData.ModelId,
		"session_prompt": sessionData.Prompt,
		"chats":          "[]", // Start with an empty chats array
		"session_name":   sessionData.SessionName,
		"chat_summary":   chatsJSON,
		"file_name":      fileNameJSON,
	}

	_, err = dataBase.Cache.HSet(context.Background(), key, data).Result()
	if err != nil {
		return "", err
	}

	return sessionId, nil
}

func (dataBase *Database) SetUserValues(userId string, balance float64) error {
	// Redis key for storing user data
	userKey := fmt.Sprintf("user:%s", userId)

	_, err := dataBase.Cache.HMSet(context.Background(), userKey, map[string]interface{}{
		"balance": balance,
	}).Result()
	if err != nil {
		return err
	}

	return nil
}

func (dataBase *Database) SetSessionValues(userId string, sessionData structures.SessionData) error {
	chatsJSON, err := json.Marshal(sessionData.Chats)
	if err != nil {
		return err
	}

	fileNameJSON, err := json.Marshal(sessionData.FileName)
	if err != nil {
		return err
	}

	// Update Redis with the new session information
	key := fmt.Sprintf("user:%s:session:%s", userId, sessionData.SessionId)
	data := map[string]interface{}{
		"session_name":   sessionData.SessionName,
		"model_id":       sessionData.ModelId,
		"session_prompt": sessionData.Prompt,
		"chats":          chatsJSON, // Start with an empty chats array
		"chat_summary":   sessionData.ChatSummary,
		"file_name":      fileNameJSON,
	}

	_, err = dataBase.Cache.HSet(context.Background(), key, data).Result()
	if err != nil {
		return err
	}
	return nil
}

func (dataBase *Database) AddNewFileInSessionData(userId string, sessionId string, fileName string) error {
	sessionData, err := dataBase.GetUserSessionData(userId, sessionId)
	if err != nil {
		return err
	}

	sessionData.FileName = append(sessionData.FileName, fileName)
	err = dataBase.SetSessionValues(userId, sessionData)
	if err != nil {
		return err
	}

	return nil
}

func (dataBase *Database) DeleteFileFromSessionData(userId string, sessionId string, fileName string) error {
	sessionData, err := dataBase.GetUserSessionData(userId, sessionId)
	if err != nil {
		return err
	}

	// Start searching from the end of the slice
	for i := len(sessionData.FileName) - 1; i >= 0; i-- {
		if sessionData.FileName[i] == fileName {
			// Remove the file by slicing
			sessionData.FileName = append(sessionData.FileName[:i], sessionData.FileName[i+1:]...)
			break
		}
	}

	// Update the session data
	err = dataBase.SetSessionValues(userId, sessionData)
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

	var fileName []string
	if err := json.Unmarshal([]byte(values["file_name"]), &fileName); err != nil {
		return structures.SessionData{}, fmt.Errorf("error parsing file_name data: %w", err)
	}

	modelId, err := strconv.Atoi(values["model_id"])
	if err != nil {
		return structures.SessionData{}, fmt.Errorf("error parsing modelId: %w", err)
	}

	// Construct the session data structure
	sessionData := structures.SessionData{
		SessionName: values["session_name"],
		SessionId:   sessionId,
		ModelId:     modelId,
		Prompt:      values["session_prompt"],
		ChatSummary: values["chat_summary"],
		FileName:    fileName,
		Chats:       chats,
	}

	return sessionData, nil
}

func (dataBase *Database) CheckModelAccessAndGetBalance(userId string, modelId int) (float64, error) {
	// Construct the key to access the user's data in Redis
	userKey := fmt.Sprintf("user:%s", userId)

	// Retrieve the models data from Redis
	userData, err := dataBase.Cache.HGetAll(context.Background(), userKey).Result()
	if err != nil {
		return 0, err
	}

	modelsJSON := userData["models"]
	balance, err := strconv.ParseFloat(userData["balance"], 64)
	if err != nil {
		return 0, err
	}

	// Parse the JSON array of model IDs
	trimmed := strings.Trim(modelsJSON, "{}")
	// Split the string by comma
	parts := strings.Split(trimmed, ",")

	// Check if the user has access to the specified modelId
	for _, part := range parts {
		if part == strconv.Itoa(modelId) {
			return balance, nil
		}
	}
	return 0, errors.New("access denied: user does not have access to this model")
}

func (dataBase *Database) GetSessionsByUserId(userId string) (structures.UserSessionResponse, error) {
	// Use parameterized query to prevent SQL injection
	query := "SELECT session_id, session_name FROM session_details WHERE user_id=$1 ORDER BY session_name DESC"

	rows, err := dataBase.Db.Query(query, userId)
	if err != nil {
		return structures.UserSessionResponse{}, err
	}
	defer rows.Close()

	var sessionInfo []structures.SessionInfo
	for rows.Next() {
		var sessionIDTemp, sessionNameTemp sql.NullString
		if err := rows.Scan(&sessionIDTemp, &sessionNameTemp); err != nil {
			return structures.UserSessionResponse{}, err
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

	return structures.UserSessionResponse{
		UserId:  userId,
		Session: sessionInfo,
	}, nil
}

func (dataBase *Database) GetUserSessionChat(sessionId string) (string, error) {
	// Use parameterized query to prevent SQL injection
	query := "SELECT Chats FROM Chat_Details WHERE Session_Id=$1"

	rows, err := dataBase.Db.Query(query, sessionId)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var chats string = "[]"
	for rows.Next() {
		var chatsTemp sql.NullString
		if err := rows.Scan(&chatsTemp); err != nil {
			return "", err
		}

		// Check if SQL values are not null before appending
		if chatsTemp.Valid {
			chats = chatsTemp.String
		}
	}
	return chats, nil
}

func (dataBase *Database) GetAIModel() (structures.AIModelsResponse, error) {
	modelNames := model_data.GetModelsName()
	sort.Strings(modelNames) // Sorting the slice to ensure the order is consistent
	return structures.AIModelsResponse{Models: modelNames}, nil

}

func (dataBase *Database) GetUpdatedSummary(existingSummary string, chat, modelName string) (string, float64, error) {
	chatSummary, cost, err := dataBase.AIService.ApiSummary(existingSummary, chat, modelName)
	if err != nil {
		return "", 0, fmt.Errorf("failed to generate new summary: %w", err)
	}
	return chatSummary, cost, nil
}

func (dataBase *Database) GetBalance(userId string) (float64, error) {
	// Construct the key to access the user's data in Redis
	userKey := fmt.Sprintf("user:%s", userId)

	// Retrieve the models data from Redis
	balance, err := dataBase.Cache.HGet(context.Background(), userKey, "balance").Result()
	if err != nil {
		return 0, err
	}

	balanceFloat, err := strconv.ParseFloat(balance, 64)
	if err != nil {
		return 0, err
	}

	return balanceFloat, nil
}

func (dataBase *Database) DeleteSessionFile(userId, sessionId string, fileName string) error {
	if fileName == "" {
		return nil
	}

	// delete the file from the session cache
	err := dataBase.DeleteFileFromSessionData(userId, sessionId, fileName)
	if err != nil {
		return fmt.Errorf("error while deleting file from cache: %w", err)
	}

	// delete the file from the database
	err = dataBase.DeleteFile(context.Background(), sessionId, fileName)
	if err != nil {
		return fmt.Errorf("error while deleting file from database: %w", err)
	}

	return nil
}
