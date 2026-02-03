package main

import (
	"fmt"

	
	"github.com/gin-gonic/gin"

	controllers "github.com/Bluesyspyder/Movie-Streaming-Application/controllers"
)

func main() {
	router := gin.Default()


	router.GET("/hello",func(c *gin.Context){
		c.String(200, "Hello, Bluesyspydermovies")
	})

	router.GET("/movies",controllers.Getmovies())
	router.GET("/movie/:imdb_id",controllers.GetMovie())
	router.POST("/addmovie",controllers.Addmovies())

	if err:=router.Run(":8080");err!=nil{
		fmt.Println("Failed to start server : ",err)
	}
}