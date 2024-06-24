package services

import (
	"fmt"
	"github.com/RediSearch/redisearch-go/v2/redisearch"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/redis/rueidis"
	"log"
	"os"
	"strconv"
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

	_ = database.Vector.Drop()

	if err := database.Vector.CreateIndex(sc); err != nil {
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
	//  FT.SEARCH user-chat:123123 "(@session_id:assdasd)=>[KNN 5 @chat_embeddings $vector AS vector_score]" PARAMS 2 vector "[0.2,0.3,0.4,...]" RETURN 3 session_id chat vector_score SORTBY vector_score DIALECT 2

	maxLimit, _ := strconv.Atoi(os.Getenv("MAX_SEARCH_RESULTS_LIMIT"))
	database.Vector.SetIndexName(fmt.Sprintf("%s:%s", os.Getenv("INDEX_NAME"), userID))
	userQueryParsed := rueidis.VectorString32(userQueryEmbedding)

	res, err := redis.Values(database.Vector.GetConnection().Do(
		"FT.SEARCH", fmt.Sprintf("%s:%s", os.Getenv("INDEX_NAME"), userID),
		fmt.Sprintf("@session_id:%s @chat_embeddings:[VECTOR_RANGE 0.2 $query_vector]=>{$YIELD_DISTANCE_AS: vector_dist}", sessionID),
		"PARAMS", 2,
		"query_vector", userQueryParsed,
		"SORTBY", "vector_dist",
		"LIMIT", "0", maxLimit,
		"DIALECT", 2))

	//res, err := redis.Values(database.Vector.GetConnection().Do(
	//	"FT.SEARCH", fmt.Sprintf("%s:%s", os.Getenv("INDEX_NAME"), userID),
	//	fmt.Sprintf("(@session_id:%s)=>[KNN %d @chat_embeddings $blob AS x]", sessionID, maxLimit),
	//	"PARAMS", 2,
	//	"blob", userQueryParsed,
	//	"SORTBY", "x",
	//	"LIMIT", "0", maxLimit,
	//	"DIALECT", 2))
	if err != nil {
		return nil, -1, err
	}

	if total, err = redis.Int(res[0], nil); err != nil {
		return nil, -1, err
	}

	docs = make([]redisearch.Document, 0, len(res)-1)

	skip := 2
	scoreIdx := -1
	fieldsIdx := 1
	payloadIdx := 1
	if len(res) > skip {
		for i := 1; i < len(res); i += skip {
			if d, e := redisearch.LoadDocument(res, i, scoreIdx, payloadIdx, fieldsIdx); e == nil {
				docs = append(docs, d)
			} else {
				log.Print("Error parsing doc: ", e)
			}
		}
	}

	return
}
