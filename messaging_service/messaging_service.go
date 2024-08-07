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
	fmt.Println("Received Model : ", received.ModelName)

	maxHistoryLength, err := strconv.Atoi(os.Getenv("MAX_CHAT_HISTORY_CONTEXT"))
	if err != nil {
		return errors.New(string(error_code.Error(error_code.ErrorCodeInternalServerError)))
	}

	var isNew bool = false
	var balance float64 = 0
	if balance, err = database.CheckModelAccessAndGetBalance(received.UserId, model_data.ModelNumber(received.ModelName)); err == redis.Nil {
		fmt.Println("User Not Exists ..!!")
		return errors.New(string(error_code.Error(error_code.ErrorCodeUserDoesNotExists)))
	} else if err != nil {
		fmt.Println("User dont have access ..!!")
		return errors.New(string(error_code.Error(error_code.ErrorCodeUserDoesNotHaveModelAccess)))
	} else if balance <= 0 {
		fmt.Println("Insufficient balance ..!!")
		return errors.New(string(error_code.Error(error_code.ErrorCodeInSufficientBalance)))
	}

	fmt.Println("User have the access ..!!")

	var sessionData structures.SessionData
	if received.SessionId == "NEW" {
		// create the session
		sessionData = structures.SessionData{
			ModelId:     model_data.ModelNumber(received.ModelName),
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
	AiResponse, sessionCost, err := database.AIService.AIApiCall(received.UserId, sessionData.SessionId,
		received.Message, sessionData.FileName, sessionData.Prompt, sessionData.Chats, sessionData.ChatSummary, model_data.ModelName(sessionData.ModelId), model_data.GetModelProvider(model_data.ModelName(sessionData.ModelId)), balance)
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

	var summaryCost float64 = 0
	sessionData.ChatSummary, summaryCost, err = database.GetUpdatedSummary(sessionData.ChatSummary, fmt.Sprintf("User: %s\n\nAssistant: %s", received.Message, AiResponse), model_data.ModelName(sessionData.ModelId))
	fmt.Println("Summary Generation Error: ", err)
	err = database.SetSessionValues(received.UserId, sessionData)
	fmt.Println("Session Value Update Error: ", err)

	fmt.Printf("API Cost: %f, summary cost: %f, total cost: %f, remaining balance: %f",
		sessionCost, summaryCost, sessionCost+summaryCost, balance-(sessionCost+summaryCost))

	sessionCost = sessionCost + summaryCost
	balance = balance - sessionCost
	err = database.SetUserValues(received.UserId, balance)
	fmt.Println("Session Value Balance Update Error: ", err)

	err = database.Stream.AddToStream(
		context.Background(),
		received.UserId,
		sessionData.SessionId,
		fmt.Sprintf("%d", sessionData.ModelId),
		sessionData.Prompt,
		string(newConversionStr),
		sessionData.ChatSummary,
		sessionData.SessionName,
		isNew,
		balance)
	fmt.Println("Add To Stream Error: ", err)
	return nil
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

func GetBalance(database *services.Database, request *structures.GetBalanceRequest, messageType int, conn *websocket.Conn) error {
	balance, err := database.GetBalance(request.UserId)
	if err != nil {
		return errors.New(string(error_code.Error(error_code.ErrorCodeUnableToGetBalanceDetails)))
	}

	data := structures.GetBalanceResponse{
		Balance: balance,
	}

	var response []byte
	if response, err = data.Marshal(); err != nil {
		err = conn.WriteMessage(messageType, error_code.Error(error_code.ErrorCodeJSONMarshal))
	} else {
		toSend := structures.ClientResponse{
			MessageType: messages.MessageCodeGetBalance,
			Data:        response,
		}

		response, _ = toSend.Marshal()
		err = conn.WriteMessage(messageType, response)
	}
	return err
}
