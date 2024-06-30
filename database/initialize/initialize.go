package initialize

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"os"
	"strconv"
)

func InitRedis() *redis.Client {
	dbName, _ := strconv.Atoi(os.Getenv("REDIS_DB"))

	redisDb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDRESS"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       dbName,
	})
	return redisDb
}

func InitPostgres() *sqlx.DB {
	databaseString := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PASSWORD"),
	)

	fmt.Println(databaseString)

	db, err := sqlx.Connect("postgres", databaseString)
	if err != nil {
		panic(err)
	}

	return db
}
