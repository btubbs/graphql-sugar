package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/graphql-go/graphql"
	handler "github.com/graphql-go/handler"
)

func main() {

	h := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})

	http.Handle("/graphql", h)
	http.ListenAndServe(":8080", nil)
}

func executeQuery(query string, schema graphql.Schema) *graphql.Result {
	result := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: query,
	})
	if len(result.Errors) > 0 {
		fmt.Printf("wrong result, unexpected errors: %v", result.Errors)
	}
	return result
}

type user struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	JoinedAt         time.Time `json:"joinedAt"`
	NumberOfChildren int       `json:"numberOfChildren"`
	FavoriteMovies   []string  `json:"favoriteMovies"`
}

var userType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "User",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},
			"name": &graphql.Field{
				Type: graphql.String,
			},
			"joinedAt": &graphql.Field{
				Type: graphql.String,
			},
			"numberOfChildren": &graphql.Field{
				Type: graphql.Int,
			},
			"favoriteMovies": &graphql.Field{
				Type: graphql.NewList(graphql.String),
			},
		},
	},
)

var data = map[string]user{
	"bob": {
		ID:   "bob",
		Name: "Bob Loblaw",
		FavoriteMovies: []string{
			"The Shawshank Redemption",
			"Weekend at Bernie's 2",
		},
		NumberOfChildren: 7,
		JoinedAt:         time.Date(2012, time.February, 3, 9, 19, 38, 4213, time.UTC),
	},
}

var schema, _ = graphql.NewSchema(
	graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
	},
)

var queryType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"user": &graphql.Field{
				Type: userType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: resolveUser,
			},
		},
	})

func resolveUser(p graphql.ResolveParams) (interface{}, error) {
	argID, ok := p.Args["id"].(string)
	if !ok {
		return nil, errors.New("id is not a string")
	}
	if user, ok := data[argID]; !ok {
		return nil, fmt.Errorf("user %s not found", argID)
	} else {
		return user, nil
	}
}

var mutationType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"saveUser": &graphql.Field{
				Type: userType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"name": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"numberOfChildren": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
					"favoriteMovies": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: resolveSaveUser,
			},
		},
	})

func resolveSaveUser(p graphql.ResolveParams) (interface{}, error) {
	var id string
	newUser := user{}

	// parse args
	if arg, ok := p.Args["id"].(string); ok {
		id = arg
		newUser.ID = id
	} else {
		return nil, errors.New("id is not a string")
	}

	if arg, ok := p.Args["name"].(string); ok {
		newUser.Name = arg
	} else {
		return nil, errors.New("name is not a string")
	}

	if arg, ok := p.Args["numberOfChildren"].(int); ok {
		newUser.NumberOfChildren = arg
	} else {
		return nil, errors.New("numberOfChildren is not an int")
	}

	if arg, ok := p.Args["favoriteMovies"].(string); ok {
		// favoriteMovies should be a JSON formatted list of strings
		var movies []string
		err := json.Unmarshal([]byte(arg), &movies)
		if err != nil {
			return nil, errors.New("could not parse favoriteMovies; not a valid JSON array")
		}
		newUser.FavoriteMovies = movies
	} else {
		return nil, errors.New("favoriteMovies is not a string")
	}

	// save user
	data[id] = newUser

	// return user
	return newUser, nil
}
