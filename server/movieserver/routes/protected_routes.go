package routes

import (
	controllers "github.com/Bluesyspyder/Movie-Streaming-Application/controllers"
	"github.com/Bluesyspyder/Movie-Streaming-Application/middleware"
	"github.com/gin-gonic/gin"

)

func Protected_routes(router *gin.Engine){
	router.Use(Authentication.AuthMiddleWare())

	router.GET("/movies",controllers.Getmovies())
	router.GET("/movie/:imdb_id",controllers.GetMovie())
	router.POST("/addmovie",controllers.Addmovies())
	router.GET("/recommendedmovies",controllers.GetRecommendation())
}