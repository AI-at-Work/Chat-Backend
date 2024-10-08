package services

import (
	"ai-chat/database/structures"
	"ai-chat/utils/model_data"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"os"
	"strconv"
	"strings"
)

func LoadAllModels(db *sqlx.DB) error {
	// Load all the models details
	insertQuery := "INSERT INTO Model_Details (Model_Id, Model_Name, context_length) VALUES (:id, :name, :len);"
	for id, name := range model_data.GetModelNumberMapping() {
		contextLength := model_data.ModelContextLength(id)
		_, err := db.NamedExec(insertQuery, map[string]interface{}{"id": id, "name": name, "len": contextLength})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate") {
				continue
			}
			return fmt.Errorf("error inserting model with id %d: %v", id, err)
		}
	}

	return nil
}

func LoadAllUsers(db *Database) error {
	query := `SELECT User_Id, UserName, Models, Balance FROM User_Data;`
	rows, err := db.Db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var userIDTemp, userNameTemp sql.NullString
		var models []uint8
		var balance float64
		if err := rows.Scan(&userIDTemp, &userNameTemp, &models, &balance); err != nil {
			return err
		}

		if !userNameTemp.Valid || !userIDTemp.Valid {
			continue
		}

		var userID, userName string
		userID = userIDTemp.String
		userName = userNameTemp.String

		// Redis key for storing user data
		userKey := fmt.Sprintf("user:%s", userID)

		_, err = db.Cache.HSet(context.Background(), userKey, map[string]interface{}{
			"username": userName,
			"models":   models,
			"balance":  balance,
		}).Result()
		if err != nil {
			return err
		}

		fmt.Println("Loaded user:", userID)
	}
	return nil
}

func PopulateRedisCache(db *Database) error {
	maxHistoryLength, err := strconv.Atoi(os.Getenv("MAX_CHAT_HISTORY_CONTEXT"))
	if err != nil {
		return err
	}

	query := `
	SELECT sd.Session_Id, sd.Session_Name, sd.User_Id, sd.Model_Id, cd.Session_Prompt, cd.Chats, cd.Chats_Summary, fd.File_Name
	FROM Session_Details sd 
	LEFT JOIN Chat_Details cd ON sd.Session_Id = cd.Session_Id 
	LEFT JOIN File_Data fd ON sd.Session_Id = fd.Session_Id
	ORDER BY sd.Created_At DESC
	`
	rows, err := db.Db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var sessionIDTemp, sessionNameTemp, userIDTemp, modelIDTemp, sessionPromptTemp, chatsTemp, chatsSummaryTemp sql.NullString
		var fileName []string
		if err := rows.Scan(&sessionIDTemp, &sessionNameTemp, &userIDTemp, &modelIDTemp, &sessionPromptTemp, &chatsTemp, &chatsSummaryTemp, pq.Array(&fileName)); err != nil {
			return err
		}

		if !modelIDTemp.Valid || !sessionNameTemp.Valid || !userIDTemp.Valid || !sessionIDTemp.Valid {
			continue
		}
		var sessionID, userID, modelID, sessionPrompt, chats, chatsSummary, sessionName string
		sessionID = sessionIDTemp.String
		userID = userIDTemp.String
		modelID = modelIDTemp.String
		sessionPrompt = sessionPromptTemp.String
		chats = chatsTemp.String
		chatsSummary = chatsSummaryTemp.String
		sessionName = sessionNameTemp.String

		// Redis key for storing session data
		key := fmt.Sprintf("user:%s:session:%s", userID, sessionID)

		fmt.Println("DATA: ", sessionID, sessionPrompt, chats)

		fileNameJSON, err := json.Marshal(fileName)
		if err != nil {
			return err
		}

		var chatsList []structures.Chat
		if err := json.Unmarshal([]byte(chats), &chatsList); err != nil {
			return fmt.Errorf("error parsing chats data: %w", err)
		}

		// Keep only the latest 10 chats
		if len(chatsList) > maxHistoryLength {
			chatsList = chatsList[len(chatsList)-maxHistoryLength:]
		}

		// Marshal the updated chats list back to JSON
		chatsUpdated, err := json.Marshal(chatsList)
		if err != nil {
			return fmt.Errorf("error marshaling updated chats data: %w", err)
		}

		// Redis hash fields
		sessionData := map[string]interface{}{
			"session_name":   sessionName,
			"model_id":       modelID,
			"session_prompt": sessionPrompt,
			"chat_summary":   chatsSummary,
			"file_name":      fileNameJSON,
			"chats":          chatsUpdated,
		}

		_, err = db.Cache.HSet(context.Background(), key, sessionData).Result()
		if err != nil {
			return err
		}
		fmt.Println("Loaded session:", sessionID)

	}
	return nil
}
