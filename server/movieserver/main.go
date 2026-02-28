package main

import (
	"context"
	"fmt"
	"log"

	database "github.com/Bluesyspyder/Movie-Streaming-Application/database"
	routes "github.com/Bluesyspyder/Movie-Streaming-Application/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning : Unable to find .env file")
	}

	router := gin.Default()
	var client *mongo.Client = database.Connect()
	if client == nil {
		log.Fatal("database connection failed")
	}

	if err := client.Ping(context.Background(), nil); err != nil {
		log.Fatalf("Failed to reach MongoDB server: %v", err)
	}

	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("Failed to disconnect from MongoDB: %v", err)
		}
	}()

	routes.Unprotected_routes(router, client)
	routes.Protected_routes(router, client)

	if err := router.Run(":8080"); err != nil {
		fmt.Println("Failed to start server : ", err)
	}
}
