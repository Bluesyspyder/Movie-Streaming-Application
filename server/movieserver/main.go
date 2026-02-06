package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	controllers "github.com/Bluesyspyder/Movie-Streaming-Application/controllers"
)

func main() {
	err:=godotenv.Load()
	if err != nil{
		log.Fatal("Error loading the enviroment variables")
	}
	log.Println("ACCESS:", os.Getenv("SECRET_KEY"))
	log.Println("ACCESS:", os.Getenv("SECRET_REFRESH_KEY"))

	router := gin.Default()

	router.GET("/hello",func(c *gin.Context){
		c.String(200, "Hello, Bluesyspydermovies")
	})

	router.GET("/movies",controllers.Getmovies())
	router.GET("/movie/:imdb_id",controllers.GetMovie())
	router.POST("/addmovie",controllers.Addmovies())
	router.POST("/register",controllers.AddUsers())

	router.POST("/login",controllers.LoginUser())

	if err:=router.Run(":8080");err!=nil{
		fmt.Println("Failed to start server : ",err)
	}
}