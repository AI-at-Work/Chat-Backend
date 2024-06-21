package services

import (
	"fmt"
	"github.com/RediSearch/redisearch-go/v2/redisearch"
	"os"
	"strconv"
)

func (dataBase *Database) CreateChatSchemaInCash() error {
	dims, _ := strconv.Atoi(os.Getenv("EMBEDDING_DIMENSION"))

	// FT.CREATE chat ON HASH PREFIX 1 chat: SCHEMA timestamp NUMERIC SORTABLE
	//chat TEXT NOSTEM NOINDEX chat_embeddings VECTOR FLAT 6 DIM 128 TYPE FLOAT32 DISTANCE_METRIC L2

	// Create a schema
	sc := redisearch.NewSchema(redisearch.DefaultOptions).
		AddField(redisearch.NewNumericFieldOptions("timestamp", redisearch.NumericFieldOptions{Sortable: true, NoIndex: false})).
		AddField(redisearch.NewTextFieldOptions("chat", redisearch.TextFieldOptions{NoStem: true, NoIndex: true})).
		AddField(redisearch.NewVectorFieldOptions("chat_embeddings", redisearch.VectorFieldOptions{
			Algorithm: redisearch.Flat,
			Attributes: map[string]interface{}{
				"DIM":             dims,
				"DISTANCE_METRIC": "COSINE",
				"TYPE":            "FLOAT32",
			},
		}))
	// Drop an existing index. If the index does not exist an error is returned
	_ = dataBase.Vector.Drop()

	// Create the index with the given schema
	if err := dataBase.Vector.CreateIndex(sc); err != nil {
		return err
	}

	return nil
}

func (dataBase *Database) AddToVectorCache(userId, sessionId, timestamp, chat, vector string) error {
	key := fmt.Sprintf("user:%s:session:%s:chat:%s", userId, sessionId, timestamp)

	doc := redisearch.NewDocument(key, 1.0)
	doc.Set("timestamp", timestamp).
		Set("chat", chat).
		Set("chat_embeddings", vector)

	if err := dataBase.Vector.Index([]redisearch.Document{doc}...); err != nil {
		return err
	}
	return nil
}

func (dataBase *Database) SearchInVectorCache(userId, sessionId, userQueryEmbedding string) ([]redisearch.Document, error) {
	// Constructing the key pattern that includes userId and sessionId
	keyPattern := fmt.Sprintf("user:%s:session:%s:chat:*", userId, sessionId)

	// Limit the results to 10 for performance reasons
	limit := 10
	query := redisearch.Query{
		// Using TAG field for efficient filtering and KNN search
		Raw: fmt.Sprintf("@__key:{%s} =>[KNN 5 @chat_embeddings $vector AS score]", keyPattern),
		Params: map[string]interface{}{
			"vector": userQueryEmbedding,
		},
		SortBy: &redisearch.SortingKey{
			Field:     "timestamp",
			Ascending: false,
		},
	}

	docs, total, err := dataBase.Vector.Search(
		query.Limit(0, limit),
	)

	if err != nil {
		return nil, err
	}

	fmt.Println("Total:", total)
	fmt.Println("Docs:", docs)

	return docs, nil
}
