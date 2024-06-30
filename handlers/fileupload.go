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
	"strconv"
	"strings"
)

// WebsocketHandler sets up the WebSocket route
func FileUploadHandler(url string, app *fiber.App, database *services.Database) {
	app.Post(url, func(ctx *fiber.Ctx) error {
		return fileUpload(ctx, database)
	})
}

func fileUpload(c *fiber.Ctx, database *services.Database) error {
	fmt.Println("File Upload")

	// Extract session ID from form value
	sessionId := c.FormValue("session_id")
	if sessionId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Session ID is required",
			"data":    nil,
		})
	}

	fmt.Println("FILE UPLOAD :  ", c.FormValue("user_id"), c.FormValue("model_id"), c.FormValue("session_prompt"))

	if sessionId == "NEW" {
		var err error
		sessionId, err = fileUploadForNewSession(database, c.FormValue("user_id"), c.FormValue("model_id"), c.FormValue("session_prompt"))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Unable To Create New Session",
				"data":    nil,
			})
		}
	}

	fmt.Println("session_id: ", sessionId)

	// parse incoming image file
	file, err := c.FormFile("file")
	if err != nil {
		log.Println("image upload error --> ", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Server error",
			"data":    nil,
		})
	}

	// generate new uuid for image name
	uniqueId := uuid.New()

	// remove "- from imageName"
	filename := strings.Replace(uniqueId.String(), "-", "", -1)

	// extract image extension from original file filename
	fileSplit := strings.Split(file.Filename, ".")
	fileExt := fileSplit[len(fileSplit)-1]

	// generate image from filename and extension
	fileName := fmt.Sprintf("%s.%s", filename, fileExt)

	err = database.SaveFile(c, sessionId, fileName, file)
	if err != nil {
		log.Println("image save error --> ", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Server error",
			"data":    nil,
		})
	}

	fmt.Println("File Saved ..!!")

	// generate image url to serve to client using CDN
	imageUrl := fmt.Sprintf("http://%s/%s/images/%s", os.Getenv("SERVER_ADDRESS"), os.Getenv("PUBLIC_DIR"), fileName)

	// create metadata and send to client
	data := map[string]interface{}{
		"fileName":  fileName,
		"sessionId": sessionId,
		"imageUrl":  imageUrl,
		"header":    file.Header,
		"size":      file.Size,
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message": "Image uploaded successfully",
		"data":    data,
	})
}

func fileUploadForNewSession(database *services.Database, userId string, modelId string, sessionPrompt string) (string, error) {
	modelIdInt, err := strconv.Atoi(modelId)
	if err != nil {
		return "", errors.New(string(error_code.Error(error_code.ErrorCodeUnableToCreateSession)))
	}

	sessionId, err := database.CreateNewSession(userId, sessionPrompt, modelIdInt)
	if err != nil {
		return "", errors.New(string(error_code.Error(error_code.ErrorCodeUnableToCreateSession)))
	}
	sessionData := structures.SessionData{
		ModelId:     modelIdInt,
		SessionName: helper_functions.TruncateText("New Chat", 20),
		SessionId:   sessionId,
		Prompt:      sessionPrompt,
		ChatSummary: "",
		Chats:       nil,
	}

	_ = database.SetSessionValues(userId, sessionData.Prompt, sessionData.ModelId, sessionData.SessionId, sessionData.Chats, sessionData.ChatSummary)
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
