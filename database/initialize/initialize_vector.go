package initialize

import (
	"github.com/RediSearch/redisearch-go/v2/redisearch"
	"github.com/gomodule/redigo/redis"
	"os"
	"strconv"
)

func InitRedisChatVector() *redisearch.Client {
	dbName, _ := strconv.Atoi(os.Getenv("REDIS_DB"))

	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", os.Getenv("REDIS_ADDRESS"),
				redis.DialPassword(os.Getenv("REDIS_PASSWORD")),
				redis.DialDatabase(dbName),
			)
		},
	}

	client := redisearch.NewClientFromPool(pool, "chat")
	return client
}
