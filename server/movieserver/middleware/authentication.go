package Authentication

import (
	"net/http"

	"github.com/Bluesyspyder/Movie-Streaming-Application/utilities"
	"github.com/gin-gonic/gin"
)

func AuthMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context){
		token, err := utilities.GetAccessTokens(c)
		
		if err!=nil{
			c.JSON(http.StatusUnauthorized,gin.H{"error":err.Error()})
			c.Abort() 
			return
		}

		if token == ""{
			c.JSON(http.StatusUnauthorized,gin.H{"error":"Unable to find the token provided"})
			c.Abort()	
			return
		}

		claims, err := utilities.ValidateTokens(token)
		if err!=nil{
			c.JSON(http.StatusUnauthorized,gin.H{"error":"Invalid Token"})
			c.Abort()
			return
		}

		c.Set("userId",claims.UserID)
		c.Set("role",claims.Role)

		c.Next()
	}
} 






