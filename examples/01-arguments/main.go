package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	sugar "github.com/btubbs/graphql-sugar"
	"github.com/graphql-go/graphql"
	handler "github.com/graphql-go/handler"
)

func main() {

	mutationType := graphql.NewObject(
		graphql.ObjectConfig{
			Name: "Mutation",
			Fields: graphql.Fields{
				"saveUser": &graphql.Field{
					Type:    userType,
					Args:    sugar.ArgsConfig(saveUserArgs{}),
					Resolve: resolveSaveUser,
				},
			},
		})

	schema, _ := graphql.NewSchema(
		graphql.SchemaConfig{
			Query:    queryType,
			Mutation: mutationType,
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
				Type: sugar.Timestamp,
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

type saveUserArgs struct {
	ID               string     `arg:"id,required" desc:"A short identifier for this user."`
	Name             string     `arg:"name,required" desc:"This user's name."`
	NumberOfChildren int        `arg:"numberOfChildren" desc:"The number of children that this user has."`
	FavoriteMovies   moviesList `arg:"favoriteMovies" desc:"A JSON-formatted list of this user's favorite movies."`
}

type moviesList []string

func init() {
	parseMoviesList := func(arg interface{}) (moviesList, error) {
		if ms, ok := arg.([]interface{}); ok {
			out := moviesList{}
			for _, m := range ms {
				out = append(out, m.(string))
			}
			return out, nil
		}
		return nil, errors.New("invalid movie list")
	}
	if err := sugar.RegisterArgParser(parseMoviesList, graphql.NewList(graphql.String)); err != nil {
		panic(err)
	}
}

func resolveSaveUser(p graphql.ResolveParams) (interface{}, error) {
	args := saveUserArgs{}
	if err := sugar.LoadArgs(p, &args); err != nil {
		return nil, err
	}

	newUser := user{
		ID:               args.ID,
		Name:             args.Name,
		NumberOfChildren: args.NumberOfChildren,
		FavoriteMovies:   args.FavoriteMovies,
	}
	newUser.JoinedAt = time.Now()

	// save user
	data[newUser.ID] = newUser

	// return user
	return newUser, nil
}
