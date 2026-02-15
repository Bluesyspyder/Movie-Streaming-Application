package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	routes "github.com/Bluesyspyder/Movie-Streaming-Application/routes"
)

func main() {
	err:=godotenv.Load()
	if err != nil{
		log.Fatal("Error loading the enviroment variables")
	}
	log.Println("ACCESS:", os.Getenv("SECRET_KEY"))
	log.Println("ACCESS:", os.Getenv("SECRET_REFRESH_KEY"))

	router := gin.Default()
	

	routes.Unprotected_routes(router)
	routes.Protected_routes(router)

	if err:=router.Run(":8080");err!=nil{
		fmt.Println("Failed to start server : ",err)
	}
}