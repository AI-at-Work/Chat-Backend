package handlers

import (
	"ai-chat/database/services"
	"ai-chat/database/structures"
	"ai-chat/messaging_service"
	"ai-chat/utils/response_code/error_code"
	"ai-chat/utils/response_code/messages"
	"encoding/json"
	"fmt"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"log"
)

// WebsocketHandler sets up the WebSocket route
func WebsocketHandler(url string, app *fiber.App, database *services.Database) {
	app.Use(url, websocket.New(func(c *websocket.Conn) {
		fmt.Println("New WebSocket Connection")
		defer c.Close() // Ensure the connection is closed after return1

		NewConnection(c, database)
	}))
}

// Send error message over WebSocket connection
func sendErrorOverWebSocket(c *websocket.Conn, errMsg string) {
	if err := c.WriteMessage(websocket.TextMessage, []byte(errMsg)); err != nil {
		log.Printf("Failed to send error message over WebSocket: %v", err)
	}
}

// NewConnection handles incoming messages and sends responses
func NewConnection(conn *websocket.Conn, database *services.Database) {
	msg := &structures.ClientRequest{}
	for {
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			break // Exit the loop on read error
		}

		if err := json.Unmarshal(data, msg); err != nil {
			log.Println("Unmarshal error:", err)
			sendErrorOverWebSocket(conn, string(error_code.Error(error_code.ErrorCodeJSONUnmarshal)))
			continue
		}

		switch msg.MessageType {
		// Define your cases here as in your original handler
		case messages.MessageCodeUserDetails:
			var dataReceived structures.UserDataRequest
			dataReceived.Unmarshal(msg.Data)
			err = messaging_service.GetUserDetails(database, &dataReceived, messageType, conn)
		case messages.MessageCodeListSessions:
			var dataReceived structures.UserSessionsRequest
			dataReceived.Unmarshal(msg.Data)
			err = messaging_service.GetListOfSessions(database, &dataReceived, messageType, conn)
		case messages.MessageCodeChatsBySessionId:
			var dataReceived structures.SessionChatsRequest
			dataReceived.Unmarshal(msg.Data)
			err = messaging_service.GetChatsBySessionId(database, &dataReceived, messageType, conn)
		case messages.MessageCodeChatMessage:
			var dataReceived structures.UserMessageRequest
			dataReceived.Unmarshal(msg.Data)
			err = messaging_service.GetChatResponse(database, &dataReceived, messageType, conn)
		case messages.MessageCodeSessionDelete:
			var dataReceived structures.SessionDeleteRequest
			dataReceived.Unmarshal(msg.Data)
			err = messaging_service.DeleteSession(database, &dataReceived, messageType, conn)
		case messages.MessageCodeGetAIModels:
			var dataReceived structures.AIModelsRequest
			dataReceived.Unmarshal(msg.Data)
			err = messaging_service.AIModesList(database, &dataReceived, messageType, conn)
		default:
			sendErrorOverWebSocket(conn, string(error_code.Error(error_code.ErrorCodeUnknownMessage)))
			return
		}

		if err != nil {
			fmt.Println("Connection Closed ..!!")
			sendErrorOverWebSocket(conn, err.Error())
		}
	}
}
