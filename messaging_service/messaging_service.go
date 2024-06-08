package messaging_service

import (
	"ai-chat/api_call"
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
)

func GetChatResponse(database *services.Database, received *structures.UserMessageRequest, messageType int, conn *websocket.Conn) error {

	fmt.Println("File Name: ", received.FileName)

	var isNew bool = false
	if err := database.CheckModelAccess(received.UserId, received.ModelId); err == redis.Nil {
		return errors.New(string(error_code.Error(error_code.ErrorCodeUserDoesNotExists)))
	} else if err != nil {
		return errors.New(string(error_code.Error(error_code.ErrorCodeUserDoesNotHaveModelAccess)))
	}

	fmt.Println("User have the access ..!!")

	var sessionData structures.SessionData
	if received.SessionId == "NEW" {
		// create the session
		sessionId, err := database.CreateNewSession(received.UserId, received.Prompt, received.ModelId)
		if err != nil {
			return errors.New(string(error_code.Error(error_code.ErrorCodeUnableToCreateSession)))
		}
		sessionData = structures.SessionData{
			ModelId:     received.ModelId,
			SessionName: helper_functions.TruncateText(received.Message, 20),
			SessionId:   sessionId,
			Prompt:      received.Prompt,
			Chats:       nil,
		}
		isNew = true
		fmt.Println("ADDED TO NEW SESSION")

	} else {
		var err error
		sessionData, err = database.GetUserSessionData(received.UserId, received.SessionId)
		if err != nil {
			return errors.New(string(error_code.Error(error_code.ErrorCodeUnableToLoadSession)))
		}
	}

	sessionData.Chats = append(sessionData.Chats, structures.Chat{"user", received.Message})
	err := helper_functions.LimitTokenSize(&sessionData, model_data.ModelContextLength(sessionData.ModelId))
	if err != nil {
		return errors.New(string(error_code.Error(error_code.ErrorCodeUnableToTokenizeData)))
	}

	fmt.Println(received)
	fmt.Println("In OPEN AI ")

	// OpenAI API Call
	AiResponse, err := api_call.OpenAIApiCall(sessionData, received.FileName)
	if err != nil {
		return errors.New(string(error_code.Error(error_code.ErrorCodeUnableToReceiveResponseToQuery)))
	}

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
	newConversion = append(newConversion, structures.Chat{"user", received.Message}, structures.Chat{"assistant", AiResponse})
	if received.FileName != "" {
		sessionData.Chats = append(sessionData.Chats, structures.Chat{"assistant", AiResponse}, structures.Chat{"file", received.FileName})
		newConversion = append(newConversion, structures.Chat{"file", received.FileName})
	} else {
		// Load the changes in cache
		sessionData.Chats = append(sessionData.Chats, structures.Chat{"assistant", AiResponse})
	}

	newConversionStr, err := json.Marshal(newConversion)
	if err != nil {
		return errors.New(string(error_code.Error(error_code.ErrorCodeJSONMarshal)))
	}

	_ = database.SetSessionValues(received.UserId, sessionData.Prompt, sessionData.ModelId, sessionData.SessionId, sessionData.Chats)
	_ = database.Stream.AddToStream(
		context.Background(),
		received.UserId,
		sessionData.SessionId,
		fmt.Sprintf("%d", sessionData.ModelId),
		sessionData.Prompt,
		string(newConversionStr),
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
	data, err := database.GetUserSessionChat(received.UserId, received.SessionId)
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
