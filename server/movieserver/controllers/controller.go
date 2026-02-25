package controllers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/Bluesyspyder/Movie-Streaming-Application/database"
	"github.com/Bluesyspyder/Movie-Streaming-Application/models"
	"github.com/Bluesyspyder/Movie-Streaming-Application/utilities"
	"github.com/go-playground/validator/v10"
	"github.com/tmc/langchaingo/llms/openai"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var moviescollection *mongo.Collection = database.OpenCollection("movies")
var rankingsCollection *mongo.Collection = database.OpenCollection("rankings")

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


func AdminReview() gin.HandlerFunc {
	return func(c *gin.Context){
		movieId := c.Param("imdb_id")
		if movieId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Movie Id required"})
			return
		}
		var req struct {
			AdminReview string `json:"admin_review"`
		}
		var response struct {
			RankingName string `json:"ranking_name"`
			AdminReview string `json:"admin_review"`
		}
		
		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest,gin.H{"error":"Invalid request body"})
			return
		}

		sentiment, rankval, err := GetReviewRanking(req.AdminReview)

		if err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"Error getting review ranking"})
			return
		}

		filter := bson.M{"imdb_id":movieId}
		
		update := bson.M{
			"$set":bson.M{
				"admin_review":req.AdminReview,
				"ranking":bson.M{
					"ranking_value":rankval,
					"ranking_name":sentiment,
				},
			},
		}

		var ctx,cancel = context.WithTimeout(context.Background(),100*time.Second)
		defer cancel()

		result, err := moviescollection.UpdateOne(ctx,filter,update)

		if err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"Error updating the movie"})
			return
		}

		if result.MatchedCount == 0{
			c.JSON(http.StatusNotFound,gin.H{"error":"Movie not found"})
			return
		}

		response.RankingName = sentiment
		response.AdminReview = req.AdminReview

		c.JSON(http.StatusOK,response)

	}
}


func GetReviewRanking(admin_review string) (string,int,error){
	rankings, err := getRankings()

	if err != nil{
		return "", 0, err
	}

	sentimentDelimited := ""

	for _, rankings := range rankings {
		if rankings.Rankingvalue != 999{
			sentimentDelimited = sentimentDelimited + rankings.Rankingname
		}
	}

	sentimentDelimited = strings.Trim(sentimentDelimited,",")

	err = godotenv.Load(".env")
	if err!=nil{
		log.Println("Warning: .env file not found")
	}

	OpenAIApi := os.Getenv("OPEN_AI_KEY")
	if OpenAIApi == ""{
		return "",0,errors.New("could not read OPENAI_AI_KEY")
	}

	llm, err := openai.New(openai.WithToken(OpenAIApi))

	if err!=nil{
		return "",0, err
	}

	base_template := os.Getenv("BASE_PROMPT_TEMPLATE")
	base_prompt := strings.Replace(base_template, "{rankings}",sentimentDelimited,1)

	response, err:= llm.Call(context.Background(),base_prompt + admin_review)
	if err!=nil{
		return "",0,err
	}

	rankVal := 0

	for _,ranking := range rankings{
		if ranking.Rankingname == response{
			rankVal = ranking.Rankingvalue
		}
	}
	return response,rankVal,nil
}

func getRankings()([]models.Ranking,error){
	var rankings []models.Ranking

	var ctx, cancel = context.WithTimeout(context.Background(),100*time.Second)
	defer cancel()

	cursor, err := rankingsCollection.Find(ctx, bson.M{})

	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &rankings);err!=nil{
		return  nil, err
	}
	return rankings, nil
}


func GetRecommendation() gin.HandlerFunc{
	return func(c *gin.Context){
		  userId, err := utilities.GetuserID(c)

			if err != nil{
				c.JSON(http.StatusBadRequest,gin.H{"error":err.Error()})
				return
			}

			favGenres, err := GetuserFavRec(userId)

			if err!=nil{
				c.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
				return
			}

			err = godotenv.Load(".env")
			if err!=nil{
				log.Println("Warning : .env file not found")
			}


			var recmovielimit int64 = 5
			recommendationlimitstr := os.Getenv("RECOMMENDED_MOVIE_LIMIT")
			
			if recommendationlimitstr!=""{
				recmovielimit, _ = strconv.ParseInt(recommendationlimitstr,10,64)
			}

			findOptions := options.Find()

			findOptions.SetSort(bson.D{{Key:"ranking.ranking_value",Value:1}})

			findOptions.SetLimit(recmovielimit)

			filter := bson.M{"genre.genre_name":bson.M{"$in":favGenres}}

			ctx, cancel := context.WithTimeout(context.Background(),100*time.Second)
			defer cancel()

			cursor, err := moviescollection.Find(ctx,filter,findOptions)

			if err!=nil{
				c.JSON(http.StatusInternalServerError,gin.H{"error":"Error fetching recommended movies"})
				return
			}

			defer cursor.Close(ctx)

			var recommendedmovies []models.Movie

			if err!=cursor.All(ctx, &recommendedmovies){
				c.JSON(http.StatusInternalServerError,gin.H{"error": err.Error()})
				return
			}
			fmt.Println("USER ID:", userId)
			fmt.Println("FAV GENRES:", favGenres)
			c.JSON(http.StatusOK,recommendedmovies)
	}
}

func GetuserFavRec(userID string)([]string,error){
	var ctx, cancel = context.WithTimeout(context.Background(),100*time.Second)
	defer cancel()

	filter := bson.M{
		"userId":userID,
	}

	projection := bson.M{
		"favourite_genres.genre_name":1,
		"id":0,
	}

	options := options.FindOne().SetProjection(projection)
	var result bson.M

	err := usercollection.FindOne(ctx, filter, options).Decode(&result)

	if err!=nil{
		if err == mongo.ErrNoDocuments{
			return []string{}, nil
		}
	}

favGenresRaw, exists := result["favourite_genres"]
if !exists {
	return []string{}, nil
}

favGenres, ok := favGenresRaw.([]interface{})
if !ok {
	return []string{}, errors.New("invalid favourite_genres format")
}

var genreName []string

for _, item := range favGenres {
	if genreMap, ok := item.(bson.M); ok {
		if name, ok := genreMap["genre_name"].(string); ok {
			genreName = append(genreName, name)
		}
	}
}
fmt.Println("Fav Genres:", genreName)

return genreName, nil
}