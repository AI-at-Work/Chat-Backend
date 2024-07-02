package handlers

import (
	"ai-chat/database/services"
	"ai-chat/database/structures"
	"ai-chat/utils/helper_functions"
	"ai-chat/utils/response_code/error_code"
	"context"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const convertTOMB = 1024 * 1024 // 10 MB

// WebsocketHandler sets up the WebSocket route
func FileUploadHandler(url string, app *fiber.App, database *services.Database) {
	app.Post(url, func(ctx *fiber.Ctx) error {
		return fileUpload(ctx, database)
	})
}

func generateUniqueFileName(originalFileName string) string {
	uniqueID := uuid.New().String()
	fileExt := filepath.Ext(originalFileName)
	return fmt.Sprintf("%s%s", strings.ReplaceAll(uniqueID, "-", ""), fileExt)
}

func fileUpload(c *fiber.Ctx, database *services.Database) error {
	fmt.Println("File Upload")

	formData, err := validateAndExtractFormData(c)
	if err != nil {
		fmt.Println("File Upload Error", formData, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
			"data":    nil,
		})
	}

	// parse incoming image file
	file, err := c.FormFile("file")
	if err != nil {
		log.Println("upload error --> ", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Server error",
			"data":    nil,
		})
	}

	maxFileSize, _ := strconv.Atoi(os.Getenv("MAX_FILE_SIZE"))
	if file.Size > int64(convertTOMB*maxFileSize) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "File size exceeds the maximum limit of 10 MB",
			"data":    nil,
		})
	}

	fileName := generateUniqueFileName(file.Filename)
	log.Printf("File Upload: UserID: %s, SessionID: %s, ModelID: %s, Prompt: %s\n",
		formData.UserId, formData.SessionId, formData.ModelId, formData.Prompt)

	var allSessionFiles []string
	if formData.SessionId == "NEW" {
		var err error
		formData.SessionId, err = fileUploadForNewSession(database, formData.UserId, formData.ModelId, formData.Prompt, fileName)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Unable To Create New Session",
				"data":    nil,
			})
		}
	} else {
		sessionData, err := database.GetUserSessionData(formData.UserId, formData.SessionId)
		if err != nil {
			return err
		}

		err = database.AddNewFileInSessionData(formData.UserId, formData.SessionId, fileName)
		if err != nil {
			log.Println("cache save error --> ", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Server error",
				"data":    nil,
			})
		}

		allSessionFiles = sessionData.FileName
	}

	isNew := "OLD"
	if len(allSessionFiles) == 0 {
		isNew = "NEW"
	}

	fmt.Println("session_id: ", formData.SessionId)
	err = database.SaveFile(c, formData.SessionId, fileName, isNew, file)
	if err != nil {
		log.Println("save error --> ", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Server error",
			"data":    nil,
		})
	}
	fmt.Println("File Saved ..!!")

	// generate image url to serve to client using CDN
	fileUrl := fmt.Sprintf("http://%s/%s/%s", os.Getenv("SERVER_ADDRESS"), os.Getenv("PUBLIC_DIR"), fileName)

	// create metadata and send to client
	data := map[string]interface{}{
		"fileName":  fileName,
		"sessionId": formData.SessionId,
		"imageUrl":  fileUrl,
		"header":    file.Header,
		"size":      file.Size,
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message": "Image uploaded successfully",
		"data":    data,
	})
}

func fileUploadForNewSession(database *services.Database, userId string, modelId string, sessionPrompt, fileName string) (string, error) {
	modelIdInt, err := strconv.Atoi(modelId)
	if err != nil {
		return "", errors.New(string(error_code.Error(error_code.ErrorCodeUnableToCreateSession)))
	}

	sessionData := structures.SessionData{
		ModelId:     modelIdInt,
		SessionName: helper_functions.TruncateText("New Chat", 20),
		Prompt:      sessionPrompt,
		ChatSummary: "",
		FileName:    []string{fileName},
		Chats:       nil,
	}
	sessionId, err := database.CreateNewSession(userId, sessionData)
	if err != nil {
		return "", errors.New(string(error_code.Error(error_code.ErrorCodeUnableToCreateSession)))
	}
	sessionData.SessionId = sessionId

	err = database.AddSession(context.Background(), userId, sessionId, modelId, sessionData.SessionName)
	if err != nil {
		return "", errors.New(string(error_code.Error(error_code.ErrorCodeUnableToCreateSession)))
	}

	err = database.AddChat(context.Background(), sessionId, sessionPrompt, "[]", sessionData.ChatSummary)
	if err != nil {
		return "", errors.New(string(error_code.Error(error_code.ErrorCodeUnableToCreateSession)))
	}

	return sessionId, err
}

func validateAndExtractFormData(c *fiber.Ctx) (*structures.FormData, error) {
	formData := &structures.FormData{
		SessionId: c.FormValue("session_id"),
		UserId:    c.FormValue("user_id"),
		Prompt:    c.FormValue("session_prompt"),
		ModelId:   c.FormValue("model_id"),
	}

	if formData.SessionId == "" {
		return nil, errors.New("session ID is required")
	}
	if formData.UserId == "" {
		return nil, errors.New("user ID is required")
	}
	if formData.ModelId == "" {
		return nil, errors.New("model ID is required")
	}

	return formData, nil
}
