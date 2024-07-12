package error_code

const (
	ErrorCodeJSONUnmarshal                  = 0
	ErrorCodeUnknownMessage                 = 1
	ErrorCodeJSONMarshal                    = 2
	ErrorCodeUserDoesNotExists              = 3
	ErrorCodeUnableToCreateSession          = 4
	ErrorCodeUnableToLoadSession            = 5
	ErrorCodeUnableToReceiveResponseToQuery = 6
	ErrorCodeUserDoesNotHaveModelAccess     = 7
	ErrorCodeUnableToTokenizeData           = 8
	ErrorCodeUnableToLoadChats              = 9
	ErrorCodeUnableToDeleteSession          = 10
	ErrorCodeUnableToGenerateAIModelList    = 11
	ErrorCodeUnableToCreateEmbedding        = 12
	ErrorCodeUnableToSearchForEmbedding     = 13
	ErrorCodeInternalServerError            = 14
	ErrorCodeInSufficientBalance            = 15
	ErrorCodeUnableToGetBalanceDetails      = 16
)

var errorCodeMapping = map[int]string{
	0:  "JSON Parsing Error",
	1:  "Unknown Message",
	2:  "JSON Marshal Error",
	3:  "User Does Not Exists",
	4:  "Unable to Create Session",
	5:  "Unable to Load Session",
	6:  "Unable to Receive Response To Query",
	7:  "User Does Not Have Model Access",
	8:  "Unable to Tokenize Data",
	9:  "Unable to Load Chats",
	10: "Unable to Delete Session",
	11: "Unable to Generate AI Model List",
	12: "Unable to Create Embedding",
	13: "Unable to Search For Embedding",
	14: "Internal Server Error",
	15: "Insufficient balance",
	16: "Unable to get balance details",
}

func Error(num int) []byte {
	return []byte("{\"error\": \"" + errorCodeMapping[num] + "\"}")
}
