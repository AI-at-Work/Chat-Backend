package services

import (
	"fmt"
	"github.com/RediSearch/redisearch-go/v2/redisearch"
	"github.com/google/uuid"
	"github.com/redis/rueidis"
	"os"
	"strconv"
	"strings"
)

func (database *Database) CreateChatSchemaInCache(userID string) error {
	dims, err := strconv.Atoi(os.Getenv("EMBEDDING_DIMENSION"))
	if err != nil {
		return fmt.Errorf("failed to parse dimensions: %w", err)
	}

	database.Vector.SetIndexName(fmt.Sprintf("%s:%s", os.Getenv("INDEX_NAME"), userID))

	sc := redisearch.NewSchema(redisearch.DefaultOptions).
		AddField(redisearch.NewTextFieldOptions("session_id", redisearch.TextFieldOptions{NoStem: true, NoIndex: false})).
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

	//_ = database.Vector.Drop()

	if err := database.Vector.CreateIndex(sc); err != nil {
		if strings.Contains(err.Error(), "Index already exists") {
			return nil
		}
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
}

func (database *Database) AddToVectorCache(userID, sessionID string, timestamp int64, chat string, vector []float32) error {
	id := uuid.New()
	key := fmt.Sprintf("%s:%s:%s", os.Getenv("INDEX_NAME"), userID, id.String())

	doc := redisearch.NewDocument(key, 1.0)
	doc.Set("timestamp", timestamp).
		Set("session_id", sessionID).
		Set("chat", chat).
		Set("chat_embeddings", rueidis.VectorString32(vector))

	if err := database.Vector.Index([]redisearch.Document{doc}...); err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}

	return nil
}

func (database *Database) SearchInVectorCache(userID, sessionID string, userQueryEmbedding []float32) (docs []redisearch.Document, total int, err error) {
	maxLimit, _ := strconv.Atoi(os.Getenv("MAX_SEARCH_RESULTS_LIMIT"))
	database.Vector.SetIndexName(fmt.Sprintf("%s:%s", os.Getenv("INDEX_NAME"), userID))
	userQueryParsed := rueidis.VectorString32(userQueryEmbedding)

	fmt.Println("MAX LIMIT:", maxLimit)

	r := redisearch.Query{
		Raw: fmt.Sprintf("@session_id:(%s)=>[KNN 10 @chat_embeddings $query_vector AS vector_dist]", sessionID),
		Params: map[string]interface{}{
			"query_vector": userQueryParsed,
		},
		Dialect: 2,
		SortBy: &redisearch.SortingKey{
			Field: "vector_dist",
		},
		ReturnFields: []string{"chat", "vector_dist"},
	}
	query := r.Limit(0, maxLimit)
	docs, total, err = database.Vector.Search(query)
	if err != nil {
		return nil, -1, fmt.Errorf("failed to perform search: %w", err)
	}
	return
}
