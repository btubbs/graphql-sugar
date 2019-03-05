package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/graphql-go/graphql"
	handler "github.com/graphql-go/handler"
)

func main() {
	schema, _ := graphql.NewSchema(
		graphql.SchemaConfig{
			Query: graphql.NewObject(
				graphql.ObjectConfig{
					Name: "Query",
					Fields: graphql.Fields{
						"user": &graphql.Field{
							Type: userType, // this user graphql type is defined below
							Args: graphql.FieldConfigArgument{
								"id": &graphql.ArgumentConfig{
									Type: graphql.String,
								},
							},
							Resolve: resolveUser, // this resolver is defined below
						},
					},
				}),
			Mutation: graphql.NewObject(
				graphql.ObjectConfig{
					Name: "Mutation",
					Fields: graphql.Fields{
						"saveUser": &graphql.Field{
							Type: userType, // same user type as in the "user" resolver
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
									Type: graphql.NewList(graphql.String),
								},
							},
							Resolve: resolveSaveUser, // this mutation resolver is defined below
						},
					},
				}),
		},
	)

	h := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})

	http.Handle("/graphql", h)
	http.ListenAndServe(":8080", nil)
}

// user is the actual struct type that will be returned by our "user" and "saveUser" resolvers.
type user struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	JoinedAt         time.Time `json:"joinedAt"`
	NumberOfChildren int       `json:"numberOfChildren"`
	FavoriteMovies   []string  `json:"favoriteMovies"`
}

// userType is a graphql-go representation of a GraphQL type for our users.  This is where we
// declare what will be available in the schema for clients.
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

// resolveUser looks for a user in our data store and returns it if found
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

// resolveSaveUser upserts a user into our data store.
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

	if arg, ok := p.Args["favoriteMovies"].([]interface{}); ok {
		for _, movie := range arg {
			newUser.FavoriteMovies = append(newUser.FavoriteMovies, movie.(string))
		}
	}
	// save user
	newUser.JoinedAt = time.Now()
	data[id] = newUser

	// return user
	return newUser, nil
}

// because this is an example and not a real app, just maintain state in memory.  Start with one
// user.
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
