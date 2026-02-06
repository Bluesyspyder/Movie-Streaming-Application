package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/Bluesyspyder/Movie-Streaming-Application/models"
	"github.com/Bluesyspyder/Movie-Streaming-Application/database"
	"github.com/Bluesyspyder/Movie-Streaming-Application/utilities"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"
)

var usercollection *mongo.Collection = database.OpenCollection("users")


func HashPass(password string)(string,error){
	HashPassword, err := bcrypt.GenerateFromPassword([]byte(password),bcrypt.DefaultCost)

	if err!=nil {
		return "",err
	}

	return string(HashPassword),nil
}

func AddUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User

		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest,gin.H{"err":"Invalid Input Data"})
			return
		}

		validate := validator.New()

		if err := validate.Struct(user);err!=nil{
			c.JSON(http.StatusBadRequest,gin.H{"err":"Validation failed","details":err.Error()})
			return
		}

		HashPassword, err := HashPass(user.Password)
			
		if err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{"err":"Unable to Hash the password"})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(),100*time.Second)
		defer cancel()

		count, err := usercollection.CountDocuments(ctx, bson.M{"email":user.Email})

		if err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{"err":"Failed to check the existing user"})
			return
		}

		if count > 0 {
			c.JSON(http.StatusConflict,gin.H{"err":"User already exist"})
			return
		}

		user.UserID = bson.NewObjectID().Hex()
		user.CreatedAt = time.Now()
		user.UpdateAt = time.Now()
		user.Password = HashPassword

		result, err := usercollection.InsertOne(ctx, user)

		if err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{"err":"Failed to create user"})
			return 
		}

		c.JSON(http.StatusCreated,result)


	}
}

func LoginUser() gin.HandlerFunc {
	return  func(c *gin.Context) {
		var userLogin models.Login

		if err := c.ShouldBindBodyWithJSON(&userLogin);err!=nil{
			c.JSON(http.StatusBadRequest,gin.H{"err":"Invalid Input Data"})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(),100*time.Second)
		defer cancel()


		var founduser models.User

		if err := usercollection.FindOne(ctx, bson.M{"email": userLogin.Email}).Decode(&founduser);err != nil{
			c.JSON(http.StatusUnauthorized,gin.H{"err":"Invalid email or password"})
			return
		}

		err := bcrypt.CompareHashAndPassword([]byte(founduser.Password),[]byte(userLogin.Password))
		if err != nil {
			c.JSON(http.StatusUnauthorized,gin.H{"err":"Invalid email or password"})
			return
		}

		tokens, refreshTokens, err := utilities.GenerateToken(founduser.Email,founduser.Firstname,founduser.Lastname,founduser.Role,founduser.UserID)

		if err != nil {
			c.JSON(http.StatusInternalServerError,gin.H{"err":"Failed to generate TOkens"})
			return
		}

		err = utilities.UpdateTokens(founduser.UserID, tokens,refreshTokens)

		if err != nil {
			c.JSON(http.StatusInternalServerError,gin.H{"err":"Failed to Update TOkens"})
			return
		}

		c.JSON(http.StatusOK,models.Response{
			UserID: founduser.UserID,
			Firstname: founduser.Firstname,
			Lastname: founduser.Lastname,
			Role: founduser.Role,
			Email:founduser.Email,
			FavouriteGenre: founduser.FavouriteGenre,
			Token: tokens,
			RefreshToken: refreshTokens,
		})
	}
}

