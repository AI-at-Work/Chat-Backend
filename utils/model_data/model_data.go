package model_data

const (
	GPTTurbo125      = 0
	GPTTurbo         = 1
	GPTTurbo1106     = 2
	GPTTurboInstruct = 3
	GPT4Turbo        = 4
	GPT4Turbo09      = 5
	GPT4             = 6
	GPT40613         = 7
	GPT432k          = 8
	GPT432k0613      = 9
)

var ModelNumberMapping = map[int]string{
	0: "gpt-3.5-turbo-0125",
	1: "gpt-3.5-turbo",
	2: "gpt-3.5-turbo-1106",
	3: "gpt-3.5-turbo-instruct",
	4: "gpt-4-turbo",
	5: "gpt-4-turbo-2024-04-09",
	6: "gpt-4",
	7: "gpt-4-0613",
	8: "gpt-4-32k",
	9: "gpt-4-32k-0613",
}

var ModelNameMapping = map[string]int{
	"gpt-3.5-turbo-0125":     0,
	"gpt-3.5-turbo":          1,
	"gpt-3.5-turbo-1106":     2,
	"gpt-3.5-turbo-instruct": 3,
	"gpt-4-turbo":            4,
	"gpt-4-turbo-2024-04-09": 5,
	"gpt-4":                  6,
	"gpt-4-0613":             7,
	"gpt-4-32k":              8,
	"gpt-4-32k-0613":         9,
}

var modelNumberMappingContextLength = map[int]int{
	0: 16385,
	1: 16385,
	2: 16385,
	3: 4096,
	4: 128000,
	5: 128000,
	6: 8192,
	7: 8192,
	8: 32768,
	9: 32768,
}

func ModelName(num int) string {
	return ModelNumberMapping[num]
}

func ModelNumber(name string) int {
	return ModelNameMapping[name]
}

func ModelContextLength(num int) int {
	return modelNumberMappingContextLength[num]
}
