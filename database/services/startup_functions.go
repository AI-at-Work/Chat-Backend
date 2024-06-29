package services

import (
	"ai-chat/database/structures"
	"ai-chat/utils/model_data"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"strings"
	"time"
)

func LoadAllModels(db *sqlx.DB) error {
	// Load all the models details
	insertQuery := "INSERT INTO Model_Details (Model_Id, Model_Name, context_length) VALUES (:id, :name, :len);"
	for id, name := range model_data.ModelNumberMapping {
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
	query := `SELECT User_Id, UserName, Models FROM User_Data`
	rows, err := db.Db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var userIDTemp, userNameTemp sql.NullString
		var models []uint8
		if err := rows.Scan(&userIDTemp, &userNameTemp, &models); err != nil {
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
		}).Result()
		if err != nil {
			return err
		}

		err = db.CreateChatSchemaInCache(userID)
		if err != nil {
			return err
		}

		fmt.Println("Loaded user:", userID)
	}
	return nil
}

func PopulateRedisCache(db *Database) error {
	query := `
	SELECT sd.Session_Id, sd.Session_Name, sd.User_Id, sd.Model_Id, cd.Session_Prompt, cd.Chats, cd.Chats_Vector
	FROM Session_Details sd
	LEFT JOIN Chat_Details cd ON sd.Session_Id = cd.Session_Id ORDER BY sd.Created_At DESC
	`
	rows, err := db.Db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var sessionIDTemp, sessionNameTemp, userIDTemp, modelIDTemp, sessionPromptTemp, chatsTemp, chatsVectorTemp sql.NullString
		if err := rows.Scan(&sessionIDTemp, &sessionNameTemp, &userIDTemp, &modelIDTemp, &sessionPromptTemp, &chatsTemp, &chatsVectorTemp); err != nil {
			return err
		}

		if !modelIDTemp.Valid || !sessionNameTemp.Valid || !userIDTemp.Valid || !sessionIDTemp.Valid {
			continue
		}
		var sessionID, userID, modelID, sessionPrompt, chats, chatsVector, sessionName string
		sessionID = sessionIDTemp.String
		userID = userIDTemp.String
		modelID = modelIDTemp.String
		sessionPrompt = sessionPromptTemp.String
		chats = chatsTemp.String
		chatsVector = chatsVectorTemp.String
		sessionName = sessionNameTemp.String

		// Redis key for storing session data
		key := fmt.Sprintf("user:%s:session:%s", userID, sessionID)

		fmt.Println("DATA: ", sessionID, sessionPrompt, chats)

		// Redis hash fields
		sessionData := map[string]interface{}{
			"session_name":   sessionName,
			"model_id":       modelID,
			"session_prompt": sessionPrompt,
			"chats":          chats,
		}

		_, err := db.Cache.HSet(context.Background(), key, sessionData).Result()
		if err != nil {
			return err
		}
		fmt.Println("Loaded session:", sessionID)

		var chatsList []structures.Chat
		if err := json.Unmarshal([]byte(chats), &chatsList); err != nil {
			return fmt.Errorf("error parsing chats data: %w", err)
		}

		var vectorList [][]float32
		if err := json.Unmarshal([]byte(chatsVector), &vectorList); err != nil {
			return fmt.Errorf("error parsing chats vector data: %w", err)
		}

		fmt.Println("LEN : ", len(chatsList))
		fmt.Println("LEN : ", len(vectorList))

		indexChat := 0
		indexVector := 0
		numFiles := 0
		for indexChat < len(chatsList) {
			fmt.Println("ROLE:", chatsList[indexChat].Role)
			if chatsList[indexChat].Role == "file" {
				indexChat++
				numFiles++
				continue
			}
			err := db.AddToVectorCache(userID, sessionID, time.Now().UnixMilli(), fmt.Sprintf("{\"role\":\"%s\", \"content\":\"%s\"}",
				chatsList[indexChat].Role, chatsList[indexChat].Content), vectorList[indexVector])
			if err != nil {
				fmt.Println("ERRR: ", err)
				return err
			}
			indexVector++
			indexChat++
		}
		if indexChat-numFiles != indexVector {
			panic("Data Inconsistency Found ..!!. \n Len of chatList must be equal to vectorList")
		}
	}
	return nil
}
