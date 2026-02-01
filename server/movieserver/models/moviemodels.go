package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Genre struct {
	GenreID   int    `bson:"genre_id" json:"genre_id" validate:"required"`
	GenreName string `bson:"genre_name" json:"genre_name" validate:"required,min=2,max=50"`
}

type Ranking struct {
	Rankingvalue int `bson:"r_value" json:"r_value" validate:"required"`
	Rankingname  string `bson:"r_name" json:"r_name" validate:"required"`
}

type Movie struct {
	ID          bson.ObjectID `bson:"id" json:"id"`
	Imdb        string        `bson:"imdb" json:"imdb" validate:"required"`
	Title       string        `bson:"title" json:"title" validate:"required,min=2,max=50"`
	PosterPath  string        `bson:"poster" json:"poster" validate:"required,url"`
	YoutubeID   string        `bson:"youtube" json:"youtube" validate:"required"`
	Genre       []Genre       `bson:"genre" json:"genre" validate:"required,dive"`
	AdminReview string        `bson:"admin_review" json:"admin_review" validate:"required"`
	Ranking     Ranking       `bson:"ranking" json:"ranking" validate:"required"`
}