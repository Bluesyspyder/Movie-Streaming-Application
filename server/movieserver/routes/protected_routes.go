package routes

import (
	controllers "github.com/Bluesyspyder/Movie-Streaming-Application/controllers"
	"github.com/Bluesyspyder/Movie-Streaming-Application/middleware"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func Protected_routes(router *gin.Engine, client *mongo.Client) {
	router.Use(Authentication.AuthMiddleWare())

	router.GET("/movies", controllers.Getmovies(client))
	router.GET("/movie/:imdb_id", controllers.GetMovie(client))
	router.POST("/addmovie", controllers.Addmovies(client))
	router.GET("/recommendedmovies", controllers.GetRecommendation(client))
	router.PATCH("/updatereview/:imdb_id", controllers.AdminReview(client))
}
