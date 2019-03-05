package main

import (
	"time"

	gql "github.com/btubbs/garphunql"
	"github.com/davecgh/go-spew/spew"
)

func main() {
	client := gql.NewClient("http://localhost:8080/graphql")

	var bob user
	bobField := gql.Field("user",
		gql.Arg("id", "bob"),
		gql.Field("id"),
		gql.Field("name"),
		gql.Field("joinedAt"),
		gql.Field("numberOfChildren"),
		gql.Field("favoriteMovies"),
		gql.Dest(&bob),
	)
	err := client.Query(bobField)
	spew.Dump(bob, err)
}

type user struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	JoinedAt         time.Time `json:"joinedAt"`
	NumberOfChildren int       `json:"numberOfChildren"`
	FavoriteMovies   []string  `json:"favoriteMovies"`
}
