package messaging_service

import (
	"ai-chat/database/services"
	"ai-chat/database/structures"
	"ai-chat/utils/helper_functions"
	"ai-chat/utils/model_data"
	"ai-chat/utils/response_code/error_code"
	"ai-chat/utils/response_code/messages"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gofiber/contrib/websocket"
	"github.com/redis/go-redis/v9"
	"os"
	"strconv"
)

func GetChatResponse(database *services.Database, received *structures.UserMessageRequest, messageType int, conn *websocket.Conn) error {
	fmt.Println("Received File Name: ", received.FileName)
	fmt.Println("Received Session Id: ", received.SessionId)

	maxHistoryLength, err := strconv.Atoi(os.Getenv("MAX_CHAT_HISTORY_CONTEXT"))
	if err != nil {
		return errors.New(string(error_code.Error(error_code.ErrorCodeInternalServerError)))
	}

	var isNew bool = false
	if err := database.CheckModelAccess(received.UserId, received.ModelId); err == redis.Nil {
		fmt.Println("User Not Exists ..!!")
		return errors.New(string(error_code.Error(error_code.ErrorCodeUserDoesNotExists)))
	} else if err != nil {
		fmt.Println("User dont have access ..!!")
		return errors.New(string(error_code.Error(error_code.ErrorCodeUserDoesNotHaveModelAccess)))
	}

	fmt.Println("User have the access ..!!")

	var sessionData structures.SessionData
	if received.SessionId == "NEW" {
		// create the session
		sessionData = structures.SessionData{
			ModelId:     received.ModelId,
			SessionName: helper_functions.TruncateText(received.Message, 20),
			Prompt:      received.Prompt,
			FileName:    nil,
			ChatSummary: "",
			Chats:       nil,
		}
		sessionId, err := database.CreateNewSession(received.UserId, sessionData)
		if err != nil {
			return errors.New(string(error_code.Error(error_code.ErrorCodeUnableToCreateSession)))
		}
		sessionData.SessionId = sessionId

		isNew = true
		fmt.Println("ADDED TO NEW SESSION")

	} else {
		var err error
		sessionData, err = database.GetUserSessionData(received.UserId, received.SessionId)
		if err != nil {
			return errors.New(string(error_code.Error(error_code.ErrorCodeUnableToLoadSession)))
		}
	}

	fmt.Println("SESSION ID: ", sessionData.SessionId)
	fmt.Println("In OPEN AI ")

	// OpenAI API Call
	AiResponse, err := database.AIService.AIApiCall(received.UserId, sessionData.SessionId,
		received.Message, sessionData.FileName, sessionData.Prompt, sessionData.Chats, sessionData.ChatSummary, model_data.ModelNumberMapping[sessionData.ModelId])
	if err != nil {
		return errors.New(string(error_code.Error(error_code.ErrorCodeUnableToReceiveResponseToQuery)))
	}

	sessionData.Chats = append(sessionData.Chats, structures.Chat{"user", received.Message})

	data := structures.UserMessageResponse{
		UserId:      received.UserId,
		SessionId:   sessionData.SessionId,
		SessionName: helper_functions.TruncateText(received.Message, 20),
		Message:     AiResponse,
	}

	var response []byte
	if response, err = data.Marshal(); err != nil {
		err = conn.WriteMessage(messageType, error_code.Error(error_code.ErrorCodeJSONMarshal))
	} else {
		toSend := structures.ClientResponse{
			MessageType: messages.MessageCodeChatMessage,
			Data:        response,
		}

		response, _ = toSend.Marshal()
		err = conn.WriteMessage(messageType, response)
	}

	var newConversion []structures.Chat
	newConversion = append(newConversion, structures.Chat{Role: "user", Content: received.Message}, structures.Chat{Role: "assistant", Content: AiResponse})
	if received.FileName != "" {
		sessionData.Chats = append(sessionData.Chats, structures.Chat{Role: "assistant", Content: AiResponse}, structures.Chat{Role: "file", Content: received.FileName})
		newConversion = append(newConversion, structures.Chat{Role: "file", Content: received.FileName})
	} else {
		// Load the changes in cache
		sessionData.Chats = append(sessionData.Chats, structures.Chat{Role: "assistant", Content: AiResponse})
	}

	// Keep only the latest 10 chats
	if len(sessionData.Chats) > maxHistoryLength {
		sessionData.Chats = sessionData.Chats[len(sessionData.Chats)-maxHistoryLength:]
	}

	newConversionStr, err := json.Marshal(newConversion)
	if err != nil {
		return errors.New(string(error_code.Error(error_code.ErrorCodeJSONMarshal)))
	}

	sessionData.ChatSummary, err = database.GetUpdatedSummary(sessionData.ChatSummary, fmt.Sprintf("User: %s\n\nAssistant: %s", received.Message, AiResponse), model_data.ModelNumberMapping[sessionData.ModelId])
	_ = database.SetSessionValues(received.UserId, sessionData)

	_ = database.Stream.AddToStream(
		context.Background(),
		received.UserId,
		sessionData.SessionId,
		fmt.Sprintf("%d", sessionData.ModelId),
		sessionData.Prompt,
		string(newConversionStr),
		sessionData.ChatSummary,
		sessionData.SessionName,
		isNew)
	return err
}

func GetUserDetails(database *services.Database, received *structures.UserDataRequest, messageType int, conn *websocket.Conn) error {
	data, err := database.GetUserDetails(received.UserId)
	if err != nil {
		return errors.New(string(error_code.Error(error_code.ErrorCodeUserDoesNotExists)))
	}

	var response []byte
	if response, err = data.Marshal(); err != nil {
		err = conn.WriteMessage(messageType, error_code.Error(error_code.ErrorCodeJSONMarshal))
	} else {
		toSend := structures.ClientResponse{
			MessageType: messages.MessageCodeUserDetails,
			Data:        response,
		}

		response, _ = toSend.Marshal()
		err = conn.WriteMessage(messageType, response)
	}
	return err
}

func GetChatsBySessionId(database *services.Database, received *structures.SessionChatsRequest, messageType int, conn *websocket.Conn) error {
	data, err := database.GetUserSessionChat(received.SessionId)
	if err != nil {
		return errors.New(string(error_code.Error(error_code.ErrorCodeUnableToLoadChats)))
	}

	resp := structures.SessionChatsResponse{UserId: received.UserId, SessionId: received.SessionId, Chats: data}

	var response []byte
	if response, err = resp.Marshal(); err != nil {
		err = conn.WriteMessage(messageType, error_code.Error(error_code.ErrorCodeJSONMarshal))
	} else {
		toSend := structures.ClientResponse{
			MessageType: messages.MessageCodeChatsBySessionId,
			Data:        response,
		}

		response, _ = toSend.Marshal()
		err = conn.WriteMessage(messageType, response)
	}
	return err
}

func GetListOfSessions(database *services.Database, received *structures.UserSessionsRequest, messageType int, conn *websocket.Conn) error {
	data, err := database.GetSessionsByUserId(received.UserId)
	if err != nil {
		return errors.New(string(error_code.Error(error_code.ErrorCodeUserDoesNotExists)))
	}

	var response []byte
	if response, err = data.Marshal(); err != nil {
		err = conn.WriteMessage(messageType, error_code.Error(error_code.ErrorCodeJSONMarshal))
	} else {
		toSend := structures.ClientResponse{
			MessageType: messages.MessageCodeListSessions,
			Data:        response,
		}

		response, _ = toSend.Marshal()
		err = conn.WriteMessage(messageType, response)
	}
	return err
}

func DeleteSession(database *services.Database, received *structures.SessionDeleteRequest, messageType int, conn *websocket.Conn) error {
	data, err := database.DeleteSession(received.UserId, received.SessionId)
	if err != nil {
		return errors.New(string(error_code.Error(error_code.ErrorCodeUnableToDeleteSession)))
	}

	var response []byte
	if response, err = data.Marshal(); err != nil {
		err = conn.WriteMessage(messageType, error_code.Error(error_code.ErrorCodeJSONMarshal))
	} else {
		toSend := structures.ClientResponse{
			MessageType: messages.MessageCodeSessionDelete,
			Data:        response,
		}

		response, _ = toSend.Marshal()
		err = conn.WriteMessage(messageType, response)
	}
	return err
}

func AIModesList(database *services.Database, s *structures.AIModelsRequest, messageType int, conn *websocket.Conn) error {
	data, err := database.GetAIModel()
	if err != nil {
		return errors.New(string(error_code.Error(error_code.ErrorCodeUnableToGenerateAIModelList)))
	}

	var response []byte
	if response, err = data.Marshal(); err != nil {
		err = conn.WriteMessage(messageType, error_code.Error(error_code.ErrorCodeJSONMarshal))
	} else {
		toSend := structures.ClientResponse{
			MessageType: messages.MessageCodeGetAIModels,
			Data:        response,
		}

		response, _ = toSend.Marshal()
		err = conn.WriteMessage(messageType, response)
	}
	return err
}
