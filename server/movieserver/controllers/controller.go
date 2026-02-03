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
	"github.com/go-playground/validator/v10"

)

var moviescollection *mongo.Collection = database.OpenCollection("movies")

var validate = validator.New()

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

func GetMovie() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(),100*time.Second)
		defer cancel()

		movieID := c.Param("imdb_id")

		if movieID == ""{
			c.JSON(http.StatusBadRequest,gin.H{"error":"Movie ID is not found"})
			return
		}

		var movie models.Movie
		err := moviescollection.FindOne(ctx, bson.M{"imdb_id":movieID}).Decode(&movie)


		if err != nil{
			c.JSON(http.StatusNotFound, gin.H{"error":"movie not found"})
			return
		}

		c.JSON(http.StatusOK, movie)
	}
}


func Addmovies() gin.HandlerFunc {
	return func(c *gin.Context){
		ctx, cancel := context.WithTimeout(context.Background(),100*time.Second)
		defer cancel()

		var movie models.Movie 

		if err := c.ShouldBindJSON(&movie); err != nil {
			c.JSON(http.StatusBadRequest,gin.H{"error":"Invalid input"})
			return
		}

		if err := validate.Struct(movie);err!=nil{
			c.JSON(http.StatusBadRequest,gin.H{"error":"Validation failed","details":err.Error()})
		}

		result, err := moviescollection.InsertOne(ctx, movie)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error":"Failed to add movie"})
			return
		}

		c.JSON(http.StatusCreated, result)

	}
}