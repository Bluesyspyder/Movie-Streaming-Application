package routes

import (
	controllers "github.com/Bluesyspyder/Movie-Streaming-Application/controllers"
	"github.com/gin-gonic/gin"
)

func Unprotected_routes(router *gin.Engine){
	router.POST("/register",controllers.AddUsers())
	router.POST("/login",controllers.LoginUser())
}