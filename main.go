/*
Main entry point for Boldly Go GraphQL Application.

	Simple GraphQL API for:
		- A list of all of the users bank accounts
		- A list of all of the cards associated to the bank account
		- All of the transactions for that account
		- Mutation to store a transaction for a user
		- Mutation to create a new bank account
		- GraphQL Relay implementation to get user bank account with transactions

	GraphQL Endpoint:
		- /graphql
*/
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
)

const appPortKey = ":5000"

type BoldlyGo interface {
	Initialize()
	GraphQLSchema() *graphql.Schema
	DynamoDbSvc() *dynamodb.DynamoDB
	AuthService() AuthSvc
}

type boldlyGo struct {
	schema      *graphql.Schema
	dynamodbSvc *dynamodb.DynamoDB
	authsvc     AuthSvc
}

/*
Initialize the Boldly Go API Service

	Init required dependencies and services:
		- AWS Service Instance
		- GraphQL Schema
*/
func (b *boldlyGo) Initialize() {
	var (
		boldlyGoGraphQL BoldlyGoGraphQL = &boldlyGoGraphQL{}
		awsSvc          AwsConfig       = &awsConf{}
		auth            AuthSvc         = &authSvc{}
	)
	// init services
	schema := boldlyGoGraphQL.BuildSchema() // build Boldly Go GraphQL Schema
	b.schema = &schema
	awsSvc.Init() // build and initialize AWS Services
	b.dynamodbSvc = awsSvc.DynamoDbSvc()
	auth.Initialize() // build and initialize Auth Service
	b.authsvc = auth
}

func (b *boldlyGo) GraphQLSchema() *graphql.Schema {
	return b.schema
}

func (b *boldlyGo) DynamoDbSvc() *dynamodb.DynamoDB {
	return b.dynamodbSvc
}

func (b *boldlyGo) AuthService() AuthSvc {
	return b.authsvc
}

var boldlygo BoldlyGo = &boldlyGo{}

func main() {
	// instantiate Boldly Go Service
	boldlygo.Initialize()
	// instantiate mux router
	router := mux.NewRouter().StrictSlash(true)
	router.Methods("GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS").Schemes("http")
	// graphql handler
	h := handler.New(&handler.Config{
		Schema:   boldlygo.GraphQLSchema(),
		Pretty:   true,
		GraphiQL: true,
	})
	router.Handle("/graphql", authHeaderMiddleware(h))
	// add CORS acceptance to all requests
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "X-Requested-With", "Accept", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization"}),
	)(router)
	// start app
	fmt.Println(fmt.Sprintf("App Running on Port %s", appPortKey))
	log.Fatal(http.ListenAndServe(appPortKey, handlers.LoggingHandler(os.Stdout, corsHandler)))
}

// Add the Authorization header to the context passed to the GraphQL Handler
func authHeaderMiddleware(next *handler.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "Authorization", r.Header.Get("Authorization"))

		next.ContextHandler(ctx, w, r)
	})
}
