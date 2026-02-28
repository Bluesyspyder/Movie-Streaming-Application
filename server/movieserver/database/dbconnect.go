package database

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func Connect() *mongo.Client {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning : Unable to find .env file")
	}

	mongodb := os.Getenv("MONGODB_URI")
	if mongodb == "" {
		log.Fatal("MONGO_URI not set in environment")
	}

	fmt.Println("MongoDB URI : ", mongodb)

	clientoptions := options.Client().ApplyURI(mongodb)

	client, err := mongo.Connect(clientoptions) //context.Background(), clientoptions) for v1 not for v2

	if err != nil {
		return nil
	}

	return client
}

// func OpenCollection(collectionname string) *mongo.Collection{
// 	err := godotenv.Load(".env");if err!=nil{
// 		log.Println("Warning : Unable to find the .env file")
// 	}

// 	databasename := os.Getenv("DATABASE_NAME")

// 	fmt.Println("DATABASE_NAME : ",databasename)

// 	collection:= client.Database(databasename).Collection(collectionname)

// 	return collection
// }

func OpenCollection(collectionname string, client *mongo.Client) *mongo.Collection {
	databasename := os.Getenv("DATABASE_NAME")

	fmt.Println("USING DATABASE:", databasename)
	fmt.Println("USING COLLECTION:", collectionname)

	collection := client.Database(databasename).Collection(collectionname)

	return collection
}
