package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"log"
	"mime/multipart"
	"os"
)

func (dataBase *Database) SaveFile(conn *fiber.Ctx, sessionId string, fileName string, file *multipart.FileHeader) error {
	var query string
	var err error

	// Start a transaction
	tx, err := dataBase.Db.BeginTxx(conn.Context(), nil) // Notice the use of BeginTxx for better context support
	if err != nil {
		return fmt.Errorf("failed to start file save transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// save image to ./images dir
	err = conn.SaveFile(file, fmt.Sprintf("./%s/images/%s", os.Getenv("PUBLIC_DIR"), fileName))
	if err != nil {
		log.Println("image save error --> ", err)
		return conn.JSON(fiber.Map{"status": 500, "message": "Server error", "data": nil})
	}

	// Prepare the SQL query using named parameters
	query = `INSERT INTO File_Data (Session_Id, File_Name) VALUES (:session_id, :file_name)`
	params := map[string]interface{}{
		"session_id": sessionId,
		"file_name":  fileName,
	}

	// Execute the query
	result, err := tx.NamedExecContext(conn.Context(), query, params)
	if err != nil {
		return fmt.Errorf("failed to execute file save query: %w", err)
	}

	// Check how many rows were affected
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("in file save failed to get affected rows: %w", err)
	}
	if affected == 0 {
		return errors.New("in file save no rows were affected, possible invalid user_id or session_id")
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("in file save failed to commit transaction: %w", err)
	}

	return nil
}

func (dataBase *Database) AddSession(ctx context.Context, userId string, sessionId string, modelId string, sessionName string) error {
	var query string
	var err error

	// Start a transaction
	tx, err := dataBase.Db.BeginTxx(ctx, nil) // Notice the use of BeginTxx for better context support
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Prepare the SQL query using named parameters
	query = `INSERT INTO Session_Details (Session_Id, User_Id, Model_Id, Session_Name) VALUES (:session_id, :user_id, :model_id, :session_name)`
	params := map[string]interface{}{
		"session_id":   sessionId,
		"user_id":      userId,
		"model_id":     modelId,
		"session_name": sessionName,
	}

	// Execute the query
	result, err := tx.NamedExecContext(ctx, query, params)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	// Check how many rows were affected
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if affected == 0 {
		return errors.New("no rows were affected, possible invalid user_id or session_id")
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (dataBase *Database) AddChat(ctx context.Context, sessionId string, prompt string, chats string) error {
	var query string
	var rows sql.Result
	var err error = nil

	// Start a transaction
	tx, err := dataBase.Db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.New("unable to begin transaction")
	}

	// Use defer to roll back transaction if anything goes wrong before commit.
	defer func() {
		if err != nil {
			log.Println("Doing RollBack : ", err)
			tx.Rollback()
		}
	}()

	query = `INSERT INTO Chat_Details (Session_Id, Session_Prompt, Chats) VALUES ($1, $2, $3::JSONB)`
	rows, err = tx.ExecContext(ctx, query, sessionId, prompt, chats)

	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("error executing query: %v", err)
	}

	// Check if the operation affected any rows
	affected, err := rows.RowsAffected()
	if err != nil {
		return errors.New("unable to get affected rows")
	}
	if affected == 0 {
		return errors.New("no rows were affected, check session ID")
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return errors.New("unable to commit the transaction")
	}

	return nil
}
