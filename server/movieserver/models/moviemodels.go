package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Genre struct{
	GenreID int `bson:"genre_id" json:"genre_id"`
	GenreName string `bson:"genre_name" json:"genre_name"`
}

type ranking struct{
	rankingvalue int 
	rankingname string 
}

type Movie struct {
	ID bson.ObjectID `bson:"id" json:"id"`
	Imdb string `bson:"imdb" json:"imdb"`
	Title string `bson:"title" json:"title"`
	PosterPath string `bson:"poster" json:"poster"`
	YoutubeID string `bson:"youtube" json:"youtube"`
	Genre []Genre `bson:"genre" json:"genre"`
	AdminReview string `bson:"admin_review" json:"admin_review"`
	Ranking ranking `bson:"ranking" json:"ranking"`
}

func main(){

}