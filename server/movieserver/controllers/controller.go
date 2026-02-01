package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Bluesyspyder/Movie-Streaming-Application/database"
	"github.com/Bluesyspyder/Movie-Streaming-Application/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var moviescollection *mongo.Collection = database.OpenCollection("movies")

func Getmovies() gin.HandlerFunc {
	return func(c *gin.Context){
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

	movies := make([]models.Movie, 0)

		cursor, err := moviescollection.Find(ctx,bson.M{})

		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error":"Failed to fetch movies"})
		}

		defer cursor.Close(ctx)

		if err:= cursor.All(ctx, &movies); err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to decode movies"})
		}

		c.JSON(http.StatusOK, movies)
	}


	
}