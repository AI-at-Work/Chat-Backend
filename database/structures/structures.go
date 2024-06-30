package structures

import (
	"encoding/json"
	"log"
)

type ClientRequest struct {
	MessageType int             `json:"type"`
	Data        json.RawMessage `json:"data"`
}

type ClientResponse struct {
	MessageType int             `json:"type"`
	Data        json.RawMessage `json:"data"`
}

type UserDataRequest struct {
	UserId   string `json:"user_id" db:"user_id"`
	Username string `json:"username" db:"username"`
}

type UserDataResponse struct {
	UserId   string `json:"user_id" db:"user_id"`
	Username string `json:"username" db:"username"`
}

type UserSessionsRequest struct {
	UserId string `json:"user_id"`
}

type SessionChatsRequest struct {
	UserId    string `json:"user_id"`
	SessionId string `json:"session_id"`
}

type SessionChatsResponse struct {
	UserId    string `json:"user_id"`
	SessionId string `json:"session_id"`
	Chats     string `json:"chat"`
}

type SessionDeleteRequest struct {
	UserId    string `json:"user_id"`
	SessionId string `json:"session_id"`
}

type SessionDeleteResponse struct {
	UserId string `json:"user_id"`
}

type AIModelsRequest struct {
	UserId string `json:"user_id"`
}

type AIModelsResponse struct {
	Models []string `json:"models"`
}

type UserMessageRequest struct {
	UserId    string `json:"user_id" db:"user_id"`
	SessionId string `json:"session_id" db:"session_id"`
	ModelId   int    `json:"model_id" db:"model_id"`
	Message   string `json:"message" db:"message"`
	Prompt    string `json:"session_prompt" db:"session_prompt"`
	FileName  string `json:"file_name" db:"file_name"`
}

type UserMessageResponse struct {
	UserId      string `json:"user_id" db:"user_id"`
	SessionId   string `json:"session_id" db:"session_id"`
	SessionName string `json:"session_name" db:"session_name"`
	Message     string `json:"message" db:"message"`
}

type Chat struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Vector struct {
	Data []float32
}

type SessionData struct {
	SessionId   string `json:"session_id" db:"session_id"`
	SessionName string `json:"session_name" db:"session_name"`
	ModelId     int    `json:"model_id" db:"model_id"`
	Prompt      string `json:"session_prompt" db:"session_prompt"`
	ChatSummary string `json:"chat_summary" db:"chat_summary"`
	Chats       []Chat `json:"chats" db:"chats"`
}

type SessionInfo struct {
	SessionId   string `json:"session_id"`
	SessionName string `json:"session_name"`
}

type SessionListResponse struct {
	UserId  string        `json:"user_id"`
	Session []SessionInfo `json:"session_info"`
}

func (m *AIModelsRequest) Unmarshal(data []byte) {
	err := json.Unmarshal(data, &m)
	if err != nil {
		log.Println(err)
	}
}

func (m *AIModelsResponse) Marshal() ([]byte, error) {
	data, err := json.Marshal(m)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return data, err
}

func (m *SessionDeleteRequest) Unmarshal(data []byte) {
	err := json.Unmarshal(data, &m)
	if err != nil {
		log.Println(err)
	}
}

func (m *SessionDeleteResponse) Marshal() ([]byte, error) {
	data, err := json.Marshal(m)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return data, err
}

func (m *ClientResponse) Marshal() ([]byte, error) {
	data, err := json.Marshal(m)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return data, err
}

func (m *UserMessageResponse) Unmarshal(data []byte) {
	err := json.Unmarshal(data, &m)
	if err != nil {
		log.Println(err)
	}
}

func (m *UserMessageResponse) Marshal() ([]byte, error) {
	data, err := json.Marshal(m)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return data, err
}

func (m *SessionChatsResponse) Unmarshal(data []byte) {
	err := json.Unmarshal(data, &m)
	if err != nil {
		log.Println(err)
	}
}

func (m *SessionChatsResponse) Marshal() ([]byte, error) {
	data, err := json.Marshal(m)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return data, err
}

func (m *UserDataResponse) Unmarshal(data []byte) {
	err := json.Unmarshal(data, &m)
	if err != nil {
		log.Println(err)
	}
}

func (m *UserDataResponse) Marshal() ([]byte, error) {
	data, err := json.Marshal(m)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return data, err
}

func (m *SessionListResponse) Unmarshal(data []byte) {
	err := json.Unmarshal(data, &m)
	if err != nil {
		log.Println(err)
	}
}

func (m *SessionListResponse) Marshal() ([]byte, error) {
	data, err := json.Marshal(m)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return data, err
}

func (m *UserDataRequest) Marshal() ([]byte, error) {
	data, err := json.Marshal(m)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return data, nil
}

func (m *UserDataRequest) Unmarshal(data []byte) {
	err := json.Unmarshal(data, &m)
	if err != nil {
		log.Println(err)
	}
}

func (m *UserSessionsRequest) Marshal() ([]byte, error) {
	data, err := json.Marshal(m)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return data, err
}

func (m *UserSessionsRequest) Unmarshal(data []byte) {
	err := json.Unmarshal(data, &m)
	if err != nil {
		log.Println(err)
	}
}

func (m *SessionChatsRequest) Marshal() ([]byte, error) {
	data, err := json.Marshal(m)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return data, err
}

func (m *SessionChatsRequest) Unmarshal(data []byte) {
	err := json.Unmarshal(data, &m)
	if err != nil {
		log.Println(err)
	}
}

func (m *UserMessageRequest) Marshal() ([]byte, error) {
	data, err := json.Marshal(m)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return data, err
}

func (m *UserMessageRequest) Unmarshal(data []byte) {
	err := json.Unmarshal(data, &m)
	if err != nil {
		log.Println(err)
	}
}
