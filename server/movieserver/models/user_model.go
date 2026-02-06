package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	ID 							bson.ObjectID `bson:"_id,omitempty" json:_id,omitempty"`
	UserID 					string 				`bson:"user_id" json:"user_id"`
	Firstname 			string 				`bson:"first_name" json:"first_name" validate:"required,min=2,max=50"`
	Lastname 				string 				`bson:"last_name" json:"last_name" validate:"required,min=2,max=50"`
	Password 				string 				`bson:"password" json:"password" validate:"required,min=2,max=50"`
	Email 					string 				`bson:"email" json:"email" validate:"required,email"`
	Role 						string 				`bson:"role" json:"role" validate:"oneof=ADMIN USER"`
	CreatedAt 			time.Time 		`bson:"created_at" json:"created_at" validate:""`
	UpdateAt 				time.Time 		`bson:"update_at" json:"update_at"`
	Token 					string 				`bson:"token" json:"token"`
	RefreshToken 		string 				`bson:"refresh_token" json:"refresh_token"`
	FavouriteGenre []Genre 				`bson:"favourite_genres" json:"favourite_genres"`
}


type Login struct {
	Email string `json:"email" bson:"email" validate:"required,min=2,email"`
	Password string `json:"password" bson:"password" validate:"required,min=8,max=20"`
}

type Response struct {
	UserID string `json:"user_id"`
	Firstname string `json:"first_name" `
	Lastname string `json:"last_name"`
	Email string `json:"email"`
	Role string `json:"role"`
	Token string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	FavouriteGenre []Genre `json:"favourite_genre"`
}