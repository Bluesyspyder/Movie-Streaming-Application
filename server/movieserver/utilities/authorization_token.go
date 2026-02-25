package utilities

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Bluesyspyder/Movie-Streaming-Application/database"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

/* =====================================================
   STRUCT
===================================================== */

type SignedDetails struct {
	Email     string
	Firstname string
	Lastname  string
	Role      string
	UserID    string
	jwt.RegisteredClaims
}

var userCollection *mongo.Collection =
	database.OpenCollection("users")

/* =====================================================
   ENV HELPERS (IMPORTANT FIX)
===================================================== */

func getSecretKey() string {
	key := os.Getenv("SECRET_KEY")
	if key == "" {
		panic("SECRET_KEY not set")
	}
	return key
}

func getRefreshSecretKey() string {
	key := os.Getenv("SECRET_REFRESH_KEY")
	if key == "" {
		panic("SECRET_REFRESH_KEY not set")
	}
	return key
}

/* =====================================================
   GENERATE TOKEN
===================================================== */

func GenerateToken(
	email, firstName, lastName, role, userID string,
) (string, string, error) {

	claims := &SignedDetails{
		Email:     email,
		Firstname: firstName,
		Lastname:  lastName,
		Role:      role,
		UserID:    userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "BluesyMovies",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err :=
		token.SignedString([]byte(getSecretKey()))
	if err != nil {
		return "", "", err
	}

	refreshClaims := &SignedDetails{
		Email:     email,
		Firstname: firstName,
		Lastname:  lastName,
		Role:      role,
		UserID:    userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "BluesyMovies",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		},
	}

	refreshToken :=
		jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)

	refreshSignedToken, err :=
		refreshToken.SignedString([]byte(getRefreshSecretKey()))
	if err != nil {
		return "", "", err
	}

	return signedToken, refreshSignedToken, nil
}

/* =====================================================
   UPDATE TOKENS
===================================================== */

func UpdateTokens(UserId, token, refresh_token string) error {

	ctx, cancel :=
		context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	updateAt, _ :=
		time.Parse(time.RFC3339,
			time.Now().Format(time.RFC3339))

	updateData := bson.M{
		"$set": bson.M{
			"token":         token,
			"refresh_token": refresh_token,
			"update_at":     updateAt,
		},
	}

	x, err := userCollection.UpdateOne(
		ctx,
		bson.M{"user_id": UserId},
		updateData,
	)

	if err != nil {
		return err
	}

	if x.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

/* =====================================================
   GET ACCESS TOKEN FROM HEADER
===================================================== */

func GetAccessTokens(c *gin.Context) (string, error) {

	header := c.GetHeader("Authorization")

	if header == "" {
		return "", errors.New("authorization header required")
	}

	const bearerPrefix = "Bearer "

	if !strings.HasPrefix(header, bearerPrefix) {
		return "", errors.New("authorization must start with Bearer")
	}

	tokenString :=
		strings.TrimSpace(strings.TrimPrefix(header, bearerPrefix))

	if tokenString == "" {
		return "", errors.New("token missing")
	}

	return tokenString, nil
}

/* =====================================================
   VALIDATE TOKEN
===================================================== */

func ValidateTokens(tokenString string) (*SignedDetails, error) {

	claims := &SignedDetails{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(t *jwt.Token) (interface{}, error) {

			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil,
					fmt.Errorf("unexpected signing method")
			}

			return []byte(getSecretKey()), nil
		},
	)

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token expired")
	}

	return claims, nil
}


func GetuserID(c *gin.Context)(string,error){
	UserId, exists := c.Get("userId")

	if !exists{
		return "",errors.New("User ID does not exist in this context")
	}

	id, ok := UserId.(string)

	if !ok{
		return "",errors.New("User ID does not exist in this context")
	}

	return id, nil
}