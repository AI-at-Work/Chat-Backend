package initialize

import (
	"fmt"
	bert "github.com/go-skynet/go-bert.cpp"
	"os"
)

func InitModel() *bert.Bert {
	var err error
	modelPath := os.Getenv("BERT_MODEL_PATH")
	if modelPath == "" {
		modelPath = "./embedding_model/ggml-model-q4_0.bin" // Default path, adjust as needed
	}

	model, err := bert.New(modelPath)
	if err != nil {
		fmt.Println("Error loading the BERT model:", err)
		os.Exit(1)
	}
	fmt.Println("BERT model loaded successfully.")

	return model
}
