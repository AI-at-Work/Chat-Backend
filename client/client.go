package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"log"
)

const (
	MessageCodeUserDetails      = 0
	MessageCodeListSessions     = 1
	MessageCodeChatsBySessionId = 2
	MessageCodeChatMessage      = 3
	MessageCodeSessionDelete    = 4
)

type ClientRequest struct {
	MessageType int             `json:"type"`
	Data        json.RawMessage `json:"data"`
}

type UserData struct {
	UserId string `json:"user_id"`
}

type UserMessage struct {
	UserId    string `json:"user_id" db:"user_id"`
	SessionId string `json:"session_id" db:"session_id"`
	ModelId   int    `json:"model_id" db:"model_id"`
	Message   string `json:"message" db:"message"`
	Prompt    string `json:"session_prompt" db:"session_prompt"`
}

type ClientResponse struct {
	MessageType int             `json:"type"`
	Data        json.RawMessage `json:"data"`
}

func Unmarshal(data []byte, m *ClientResponse) {
	err := json.Unmarshal(data, m)
	if err != nil {
		log.Println(err)
	}
}

func main() {
	// Connect to the WebSocket server
	conn, _, _, err := ws.Dial(context.Background(), "ws://localhost:8080/ws")
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Prepare the user details request
	userDetails := UserMessage{
		UserId: "d3f01d09-e4cc-46a9-9a92-e84bb8b6bd6f",
		//SessionId: "33deeb43-8c61-4dc8-8f8d-69a2139b9bd9",
		//ModelId:   48,
		//Message:   "Ok AB Chal Raha Hai ..!!",
		//Prompt:    "YOU are coder",
	} // Specify the user ID you want to fetch
	userDataBytes, err := json.Marshal(userDetails)
	if err != nil {
		log.Fatalf("Failed to marshal request: %v", err)
	}

	request := ClientRequest{
		MessageType: MessageCodeUserDetails, // MessageType for user details as per your server code
		Data:        userDataBytes,
	}

	// Marshal the request into JSON
	requestData, err := json.Marshal(request)
	if err != nil {
		log.Fatalf("Failed to marshal request: %v", err)
	}

	// Send the request
	err = wsutil.WriteClientMessage(conn, ws.OpText, requestData)
	if err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}

	// Read the response
	response, _, err := wsutil.ReadServerData(conn)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}
	fmt.Printf("Response from server: %s\n", response)

	var clientResponse ClientResponse
	Unmarshal(response, &clientResponse)

	fmt.Printf("ClientResponse %s\n", clientResponse.Data)

}
