package model_data

const (
	GPTTurbo125      = 0
	GPTTurbo         = 1
	GPTTurbo1106     = 2
	GPTTurboInstruct = 3
	GPT4Turbo        = 4
	GPT4Turbo09      = 5
	GPT4             = 6
	GPT432k          = 7
)

var ModelNumberMapping = map[int]string{
	0: "gpt-3.5-turbo-0125",
	1: "gpt-3.5-turbo",
	2: "gpt-3.5-turbo-1106",
	3: "gpt-3.5-turbo-instruct",
	4: "gpt-4-turbo",
	5: "gpt-4-turbo-2024-04-09",
	6: "gpt-4",
	7: "gpt-4-32k",
}

var ModelNameMapping = map[string]int{
	"gpt-3.5-turbo-0125":     0,
	"gpt-3.5-turbo":          1,
	"gpt-3.5-turbo-1106":     2,
	"gpt-3.5-turbo-instruct": 3,
	"gpt-4-turbo":            4,
	"gpt-4-turbo-2024-04-09": 5,
	"gpt-4":                  6,
	"gpt-4-32k":              7,
}

var modelNumberMappingContextLength = map[int]int{
	0: 16385,
	1: 16385,
	2: 16385,
	3: 4096,
	4: 128000,
	5: 128000,
	6: 8192,
	7: 32768,
}

var ModelPricing = map[string]struct {
    Input  float64
    Output float64
}{
    "gpt-3.5-turbo-0125":     {Input: 0.00050, Output: 0.00150},
    "gpt-3.5-turbo":          {Input: 0.0030, Output: 0.0060},
    "gpt-3.5-turbo-1106":     {Input: 0.0010, Output: 0.0020},
    "gpt-3.5-turbo-instruct": {Input: 0.00150, Output: 0.00200},
    "gpt-4-turbo":            {Input: 0.0100, Output: 0.0300},
    "gpt-4-turbo-2024-04-09": {Input: 0.0100, Output: 0.0300},
    "gpt-4":                  {Input: 0.0300, Output: 0.0600},
    "gpt-4-32k":              {Input: 0.0600, Output: 0.1200},
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
