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

var validate = validator.New()

func Getmovies(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c, 100*time.Second)
		defer cancel()

		var moviecollection *mongo.Collection = database.OpenCollection("movies", client)

		movies := make([]models.Movie, 0)

		cursor, err := moviecollection.Find(ctx, bson.M{})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch movies"})
			return
		}

		defer cursor.Close(ctx)

		if err := cursor.All(ctx, &movies); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode movies"})
			return
		}

		c.JSON(http.StatusOK, movies)
	}
}

func GetMovie(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c, 100*time.Second)
		defer cancel()

		movieID := c.Param("imdb_id")

		if movieID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Movie ID is not found"})
			return
		}

		var movie models.Movie
		var moviecollection *mongo.Collection = database.OpenCollection("movies", client)
		err := moviecollection.FindOne(ctx, bson.M{"imdb_id": movieID}).Decode(&movie)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
			return
		}

		c.JSON(http.StatusOK, movie)
	}
}

func Addmovies(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c, 100*time.Second)
		defer cancel()

		var movie models.Movie

		if err := c.ShouldBindJSON(&movie); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		if err := validate.Struct(movie); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
			return
		}

		var moviecollection *mongo.Collection = database.OpenCollection("movies", client)
		result, err := moviecollection.InsertOne(ctx, movie)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add movie"})
			return
		}

		c.JSON(http.StatusCreated, result)

	}
}

func AdminReview(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		role, err := utilities.GetuserRole(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Not able to fetch user role"})
			return
		}

		if role != "ADMIN" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authorized for the requested actions"})
			return
		}

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

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		sentiment, rankval, err := GetReviewRanking(req.AdminReview, client, c)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		filter := bson.M{"imdb_id": movieId}

		update := bson.M{
			"$set": bson.M{
				"admin_review": req.AdminReview,
				"ranking": bson.M{
					"ranking_value": rankval,
					"ranking_name":  sentiment,
				},
			},
		}

		var ctx, cancel = context.WithTimeout(c, 100*time.Second)
		defer cancel()

		var moviecollection *mongo.Collection = database.OpenCollection("movies", client)
		result, err := moviecollection.UpdateOne(ctx, filter, update)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating the movie"})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
			return
		}

		response.RankingName = sentiment
		response.AdminReview = req.AdminReview

		c.JSON(http.StatusOK, response)

	}
}

func GetReviewRanking(admin_review string, client *mongo.Client, c *gin.Context) (string, int, error) {
	rankings, err := getRankings(client, c)

	if err != nil {
		return "", 0, err
	}

	sentimentDelimited := ""

	for _, ranking := range rankings {
		if ranking.Rankingvalue != 999 {
			if sentimentDelimited != "" {
				sentimentDelimited += ","
			}
			sentimentDelimited += ranking.Rankingname
		}
	}

	err = godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: .env file not found")
	}

	OpenAIApi := os.Getenv("OPEN_AI_KEY")
	if OpenAIApi == "" {
		OpenAIApi = os.Getenv("OPENAI_KEY")
	}
	if OpenAIApi == "" {
		return "", 0, errors.New("could not read OPEN_AI_KEY or OPENAI_KEY")
	}

	llm, err := openai.New(openai.WithToken(OpenAIApi))

	if err != nil {
		return "", 0, err
	}

	base_template := os.Getenv("BASE_PROMPT_TEMPLATE")
	base_prompt := strings.Replace(base_template, "{rankings}", sentimentDelimited, 1)

	response, err := llm.Call(c, base_prompt+admin_review)
	if err != nil {
		return "", 0, err
	}
	response = strings.TrimSpace(response)

	rankVal := 0

	for _, ranking := range rankings {
		if ranking.Rankingname == response {
			rankVal = ranking.Rankingvalue
		}
	}
	return response, rankVal, nil
}

func getRankings(client *mongo.Client, c *gin.Context) ([]models.Ranking, error) {
	var rankings []models.Ranking

	var ctx, cancel = context.WithTimeout(c, 100*time.Second)
	defer cancel()

	var rankingcollection *mongo.Collection = database.OpenCollection("rankings", client)
	cursor, err := rankingcollection.Find(ctx, bson.M{})

	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &rankings); err != nil {
		return nil, err
	}
	return rankings, nil
}

func GetRecommendation(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId, err := utilities.GetuserID(c)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		favGenres, err := GetuserFavRec(userId, client, c)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		err = godotenv.Load(".env")
		if err != nil {
			log.Println("Warning : .env file not found")
		}

		var recmovielimit int64 = 5
		recommendationlimitstr := os.Getenv("RECOMMENDED_MOVIE_LIMIT")

		if recommendationlimitstr != "" {
			recmovielimit, _ = strconv.ParseInt(recommendationlimitstr, 10, 64)
		}

		findOptions := options.Find()

		findOptions.SetSort(bson.D{{Key: "ranking.ranking_value", Value: 1}})

		findOptions.SetLimit(recmovielimit)

		filter := bson.M{"genre.genre_name": bson.M{"$in": favGenres}}

		ctx, cancel := context.WithTimeout(c, 100*time.Second)
		defer cancel()

		var moviecollection *mongo.Collection = database.OpenCollection("movies", client)
		cursor, err := moviecollection.Find(ctx, filter, findOptions)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching recommended movies"})
			return
		}

		defer cursor.Close(ctx)

		recommendedmovies := make([]models.Movie, 0)

		if err := cursor.All(ctx, &recommendedmovies); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, recommendedmovies)
	}
}

func GetuserFavRec(userID string, client *mongo.Client, c *gin.Context) ([]string, error) {
	var ctx, cancel = context.WithTimeout(c, 100*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id": userID,
	}

	projection := bson.M{
		"favourite_genres.genre_name": 1,
		"_id":                         0,
	}

	options := options.FindOne().SetProjection(projection)
	var result bson.M

	var usercollection *mongo.Collection = database.OpenCollection("users", client)
	err := usercollection.FindOne(ctx, filter, options).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return []string{}, nil
		}
		return nil, err
	}

	favGenresRaw, exists := result["favourite_genres"]
	if !exists {
		return []string{}, nil
	}

	var favGenres []interface{}
	switch genres := favGenresRaw.(type) {
	case bson.A:
		favGenres = make([]interface{}, len(genres))
		for i, item := range genres {
			favGenres[i] = item
		}
	case []interface{}:
		favGenres = genres
	default:
		return []string{}, nil
	}

	genreName := make([]string, 0, len(favGenres))
	for _, item := range favGenres {
		switch genreMap := item.(type) {
		case bson.M:
			if name, ok := genreMap["genre_name"].(string); ok {
				genreName = append(genreName, name)
			}
		case map[string]interface{}:
			if name, ok := genreMap["genre_name"].(string); ok {
				genreName = append(genreName, name)
			}
		case bson.D:
			for _, entry := range genreMap {
				if entry.Key == "genre_name" {
					if name, ok := entry.Value.(string); ok {
						genreName = append(genreName, name)
					}
					break
				}
			}
		}
	}
	fmt.Println("Fav Genres:", genreName)

	return genreName, nil
}
