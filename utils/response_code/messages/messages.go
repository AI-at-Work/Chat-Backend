package messages

const (
	MessageCodeUserDetails      = 0
	MessageCodeListSessions     = 1
	MessageCodeChatsBySessionId = 2
	MessageCodeChatMessage      = 3
	MessageCodeSessionDelete    = 4
	MessageCodeGetAIModels      = 5
)

var messageCodeMapping = map[int]string{
	0: "User Details",
	1: "Message Listing",
	2: "Chats By SessionId",
	3: "Chat Message",
	4: "Session Delete",
	5: "Get AI Models",
}

func Message(num int) []byte {
	return []byte(messageCodeMapping[num])
}
