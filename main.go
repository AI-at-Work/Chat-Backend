package main

import (
	"ai-chat/database/services"
	"ai-chat/handlers"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

/*
We need one more service which will perodically checks is database and cache is consistent or not;
this is to make sure that if some bad happens then still all things remain consistent;
this could be achieved by monitoring the logs metrics very carefully.

*/

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Unable to load .env")
		return
	}

	database := services.GetDataBase()
	log.Println("Database connected")

	if err = services.LoadAllModels(database.Db); err != nil {
		log.Println("Unable to load model data in database", err)
		return
	}
	log.Println("AI Models Loaded successfully")

	if err := services.LoadAllUsers(database); err != nil {
		log.Println("Unable to load users data in cache", err)
		return
	}
	log.Println("Users Data Loaded successfully")

	if err := services.PopulateRedisCache(database); err != nil {
		log.Println("Unable to load session data in cache", err)
		return
	}
	log.Println("Session Loaded successfully")

	maxFileSize, _ := strconv.Atoi(os.Getenv("MAX_FILE_SIZE"))
	app := fiber.New(fiber.Config{
		BodyLimit:    maxFileSize * 1024 * 1024, // 50MB
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	})

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowHeaders: "Origin,Content-Type,Accept,Content-Length,Accept-Language,Accept-Encoding,Connection,Access-Control-Allow-Origin",
		AllowOrigins: "*",
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
	}))

	// Static files
	app.Static("/uploads", "./"+os.Getenv("PUBLIC_DIR"))

	// websockets
	handlers.WebsocketHandler("/ws", app, database)

	// file upload
	handlers.FileUploadHandler("/upload", app, database)

	log.Printf("Server is starting at %s\n", os.Getenv("SERVER_ADDRESS"))
	log.Fatal(app.Listen(fmt.Sprintf("%s:%s", os.Getenv("SERVER_HOST"), os.Getenv("SERVER_PORT"))))
}
