package routes

import (
	controllers "github.com/Bluesyspyder/Movie-Streaming-Application/controllers"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func Unprotected_routes(router *gin.Engine,client *mongo.Client){
	router.POST("/register",controllers.AddUsers(client))
	router.POST("/login",controllers.LoginUser(client))
	
}