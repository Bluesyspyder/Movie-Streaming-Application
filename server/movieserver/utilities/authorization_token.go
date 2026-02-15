package utilities

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/Bluesyspyder/Movie-Streaming-Application/database"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type SignedDetails struct {
	Email string
	Firstname string
	Lastname string
	Role string
	UserID string
	jwt.RegisteredClaims//Issuer(Iss),subject(Sub),Audience(Aud),expiresat,notbefore,ID(jti)
}
var SECRET_KEY string = os.Getenv("SECRET_KEY")
var SECRET_REFRESH_KEY string = os.Getenv("SECRET_REFRESH_KEY")
var userCollection *mongo.Collection = database.OpenCollection("users")

func GenerateToken(email,firstName,lastName,role,userID string)(string,string,error){
	claims := &SignedDetails{
		Email:email,
		Firstname: firstName,
		Lastname: lastName,
		Role: role,
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:"BluesyMovies",
			IssuedAt: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15*time.Minute)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(SECRET_KEY))

	if err != nil {
		return "","",err
	}
	refresh_claims := &SignedDetails{
		Email:email,
		Firstname: firstName,
		Lastname: lastName,
		Role: role,
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:"BluesyMovies",
			IssuedAt: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24*7*time.Hour)),
		},
	}

	refresh_token := jwt.NewWithClaims(jwt.SigningMethodHS256, refresh_claims)

	refresh_signedToken, err := refresh_token.SignedString([]byte(SECRET_REFRESH_KEY))

	if err != nil {
		return "","",err
	}

	return signedToken, refresh_signedToken ,nil 
}


func UpdateTokens(UserId, token, refresh_token string) (err error) {
	var ctx, cancel  = context.WithTimeout(context.Background(),100*time.Second)
	defer cancel()


	updateAt, _ := time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))

	updateData := bson.M{
		"$set": bson.M{
			"token" : token,
			"refresh_token" : refresh_token,
			"update_at" : updateAt,
		},
	}

	x, err := userCollection.UpdateOne(
		ctx,
		bson.M{"user_id":UserId},
		updateData,
		) 

	if err!= nil {
		return err
	}

	if x.MatchedCount == 0{
		return mongo.ErrNoDocuments
	}


	return nil

}




func GetAccessTokens(c *gin.Context)(string,error){
	header := c.Request.Header.Get("Authorization")

	if header == ""{
		return "",errors.New("Authorization header is required")
	}

	tokenstring := header[len("Bearer"):]

	if tokenstring == ""{
		return "",errors.New("Bearer token is required")
	}

	return tokenstring,nil
}


func ValidateTokens(tokenString string) (*SignedDetails,error){
	claims := &SignedDetails{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY),nil
	})

	if err!=nil {
		return nil,err
	}

	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok{
		return nil,err
	}


	if claims.ExpiresAt.Time.Before(time.Now()){
		return nil, errors.New("Token has expired")
	}

	return claims,nil
}