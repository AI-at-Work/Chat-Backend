package initialize

import (
	"ai-chat/utils/model_data"
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func InitRedis() *redis.Client {
	dbName, _ := strconv.Atoi(os.Getenv("REDIS_DB"))

	redisDb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDRESS"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       dbName,
	})
	return redisDb
}

func InitPostgres() *sqlx.DB {
	databaseString := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PASSWORD"),
	)

	fmt.Println(databaseString)

	db, err := sqlx.Connect("postgres", databaseString)
	if err != nil {
		panic(err)
	}

	return db
}

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

func LoadAllUsers(db *sqlx.DB, rdb *redis.Client) error {
	query := `SELECT User_Id, UserName, Models FROM User_Data`
	rows, err := db.Query(query)
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

		_, err = rdb.HSet(context.Background(), userKey, map[string]interface{}{
			"username": userName,
			"models":   models,
		}).Result()
		if err != nil {
			return err
		}
		fmt.Println("Loaded user:", userID)
	}
	return nil
}

func LoadSessionDetails(db *sqlx.DB, rdb *redis.Client) error {
	query := `
	SELECT sd.Session_Id, sd.Session_Name, sd.User_Id, sd.Model_Id, cd.Session_Prompt, cd.Chats
	FROM Session_Details sd
	LEFT JOIN Chat_Details cd ON sd.Session_Id = cd.Session_Id ORDER BY sd.Created_At DESC
	`
	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var sessionIDTemp, sessionNameTemp, userIDTemp, modelIDTemp, sessionPromptTemp, chatsTemp sql.NullString
		if err := rows.Scan(&sessionIDTemp, &sessionNameTemp, &userIDTemp, &modelIDTemp, &sessionPromptTemp, &chatsTemp); err != nil {
			return err
		}

		if !modelIDTemp.Valid || !sessionNameTemp.Valid || !userIDTemp.Valid || !sessionIDTemp.Valid {
			continue
		}
		var sessionID, userID, modelID, sessionPrompt, chats, sessionName string
		sessionID = sessionIDTemp.String
		userID = userIDTemp.String
		modelID = modelIDTemp.String
		sessionPrompt = sessionPromptTemp.String
		chats = chatsTemp.String
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

		_, err := rdb.HSet(context.Background(), key, sessionData).Result()
		if err != nil {
			return err
		}
		fmt.Println("Loaded session:", sessionID)
	}
	return nil
}
