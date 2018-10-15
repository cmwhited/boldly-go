/*
GraphQL Instance for the Boldly Go Application.

	Instantiates the GraphQL Schema with all of the queries and mutations to be used for the app.
*/
package main

import (
	"errors"
	"fmt"

	"github.com/graphql-go/graphql"
	"github.com/mitchellh/mapstructure"
	"github.com/satori/go.uuid"
)

type BoldlyGoGraphQL interface {
	BuildSchema() graphql.Schema
	buildQuery()
	buildMutation()
}

type boldlyGoGraphQL struct {
	queries   graphql.ObjectConfig
	mutations graphql.ObjectConfig
	schema    graphql.Schema
}

// Build the Boldly Go RootQuery object which contains the queries being exposed by the service.
func (b *boldlyGoGraphQL) buildQuery() {
	b.queries = graphql.ObjectConfig{
		Name: "RootQuery",
		Fields: graphql.Fields{
			"bankAccounts": &graphql.Field{
				Type:        graphql.NewList(BankAccountType),
				Description: "Get a list of the users BankAccount records by the Bank primary key",
				Args: graphql.FieldConfigArgument{
					"bankId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					_, err := boldlygo.AuthService().ValidateToken(p.Context.Value("Authorization")) // validate auth token exists and is valid
					if err != nil {
						return nil, err
					}
					bankId := p.Args["bankId"]                       // get passed in bankId from arguments
					_bankId, err := uuid.FromString(bankId.(string)) // convert the bankId arg to a UUID
					if err != nil {
						return nil, err
					}
					return GetUserBankAccounts(_bankId) // get a list of the users BankAccounts by the bankId
				},
			},
			"bankAccount": &graphql.Field{
				Type:        BankAccountType,
				Description: "Get a unique user BankAccount record by the BankId Primary Key and Account Id",
				Args: graphql.FieldConfigArgument{
					"bankId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"accountId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					bankId := p.Args["bankId"]                       // get passed in bankId from args
					_bankId, err := uuid.FromString(bankId.(string)) // convert the bankId arg to a UUID
					if err != nil {
						return nil, err
					}
					acctId := p.Args["accountId"]                    // get passed in accountId from args
					_acctId, err := uuid.FromString(acctId.(string)) // convert the acctId arg to a UUID
					if err != nil {
						return nil, err
					}
					return GetUserBankAccount(_bankId, _acctId) // get a unique BankAccount by the BankId and AccountId
				},
			},
			"accountCards": &graphql.Field{
				Type:        graphql.NewList(CardType),
				Description: "A list of cards associated to the BankAccount",
				Args: graphql.FieldConfigArgument{
					"accountId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					acctId := p.Args["accountId"]                    // get passed in accountId from args
					_acctId, err := uuid.FromString(acctId.(string)) // convert the acctId arg to a UUID
					if err != nil {
						return nil, err
					}
					return GetAccountCards(_acctId)
				},
			},
			"accountCard": &graphql.Field{
				Type:        CardType,
				Description: "A BankAccount Card record",
				Args: graphql.FieldConfigArgument{
					"accountId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"cardId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					acctId := p.Args["accountId"]                    // get passed in accountId from args
					_acctId, err := uuid.FromString(acctId.(string)) // convert the acctId arg to a UUID
					if err != nil {
						return nil, err
					}
					cardId := p.Args["cardId"]                       // get passed in cardId from args
					_cardId, err := uuid.FromString(cardId.(string)) // convert the cardId arg to a UUID
					if err != nil {
						return nil, err
					}
					return GetAccountCard(_acctId, _cardId) // get a unique BankAccount Card by the AccountId and CardId
				},
			},
			"accountTransaction": &graphql.Field{
				Type:        TransactionType,
				Description: "A BankAccount Transaction record",
				Args: graphql.FieldConfigArgument{
					"accountId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"transactionId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					acctId := p.Args["accountId"]                    // get passed in accountId from args
					_acctId, err := uuid.FromString(acctId.(string)) // convert the acctId arg to a UUID
					if err != nil {
						return nil, err
					}
					transactionId := p.Args["transactionId"]                       // get passed in transactionId from args
					_transactionId, err := uuid.FromString(transactionId.(string)) // convert the transactionId arg to a UUID
					if err != nil {
						return nil, err
					}
					return GetAccountTransaction(_acctId, _transactionId) // get a unique BankAccount Transaction by the AccountId and TransactionId
				},
			},
		},
	}
}

// Build the Boldly Go RootMutation object which exposes the mutations available to this service
func (b *boldlyGoGraphQL) buildMutation() {
	b.mutations = graphql.ObjectConfig{
		Name: "RootMutation",
		Fields: graphql.Fields{
			"authenticate": &graphql.Field{
				Type:        graphql.NewNonNull(AuthType),
				Description: "Authenticate the user with the email and password. Returns an auth token",
				Args: graphql.FieldConfigArgument{
					"email": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"password": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					email, pwd := p.Args["email"].(string), p.Args["password"].(string)
					return Authenticate(email, pwd), nil
				},
			},
			"register": &graphql.Field{
				Type:        UserType,
				Description: "Register a new user record",
				Args: graphql.FieldConfigArgument{
					"user": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(UserInputType),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					user := p.Args["user"]                       // get the User input out of the arguments
					userMap, ok := user.(map[string]interface{}) // convert the input type to a User
					if !ok {
						return nil, errors.New("unable to convert input object to User record")
					}
					var u = new(User)                // instantiate user
					mapstructure.Decode(userMap, &u) // destructure userMap into User
					return u.Register()              // save user and return
				},
			},
			"saveBankAccount": &graphql.Field{
				Type:        BankAccountType,
				Description: "Save a new BankAccount record",
				Args: graphql.FieldConfigArgument{
					"acct": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(BankAccountInputType),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					acct := p.Args["acct"]                              // get the BankAccount input out of the arguments
					bankAccountMap, ok := acct.(map[string]interface{}) // convert the input type to a BankAccount
					if !ok {
						return nil, errors.New("unable to convert input object to BankAccount record")
					}
					var bankAccount = new(BankAccount)                // instantiate bank account
					mapstructure.Decode(bankAccountMap, &bankAccount) // destructure bankAccountMap into BankAccount
					return bankAccount.Save()                         // save bank account and return
				},
			},
			"updateBankAccount": &graphql.Field{
				Type:        BankAccountType,
				Description: "Update a BankAccount record",
				Args: graphql.FieldConfigArgument{
					"acct": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(BankAccountInputType),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					acct := p.Args["acct"]                              // get the BankAccount input out of the arguments
					bankAccountMap, ok := acct.(map[string]interface{}) // convert the input type to a BankAccount
					if !ok {
						return nil, errors.New("unable to convert input object to BankAccount record")
					}
					var bankAccount = new(BankAccount)                // instantiate bank account
					mapstructure.Decode(bankAccountMap, &bankAccount) // destructure bankAccountMap into BankAccount
					return bankAccount.Update()                       // save bank account and return
				},
			},
			"saveAccountCard": &graphql.Field{
				Type:        CardType,
				Description: "Save a new BankAccount Card record",
				Args: graphql.FieldConfigArgument{
					"card": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(CardInputType),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					c := p.Args["card"]                       // get the Card input out of the arguments
					cardMap, ok := c.(map[string]interface{}) // convert the input type to a Card Map
					if !ok {
						return nil, errors.New("unable to convert input object to Card record")
					}
					var card = new(Card)                // instantiate card
					mapstructure.Decode(cardMap, &card) // destructure cardMap into Card
					return card.Save()                  // save card and return
				},
			},
			"inactivateAccountCard": &graphql.Field{
				Type:        CardType,
				Description: "Inactivate a Bank Account Card record",
				Args: graphql.FieldConfigArgument{
					"card": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(CardInputType),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					c := p.Args["card"]                       // get the Card input out of the arguments
					cardMap, ok := c.(map[string]interface{}) // convert the input type to a Card Map
					if !ok {
						return nil, errors.New("unable to convert input object to Card record")
					}
					var card = new(Card)                // instantiate card
					mapstructure.Decode(cardMap, &card) // destructure cardMap into Card
					return card.Inactivate()            // inactivate card and return
				},
			},
			"saveTransaction": &graphql.Field{
				Type:        TransactionType,
				Description: "Save a Transaction record",
				Args: graphql.FieldConfigArgument{
					"bankId": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"txn": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(TransactionInputType),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					bankId := p.Args["bankId"]                       // get passed in bankId from args
					_bankId, err := uuid.FromString(bankId.(string)) // convert the bankId arg to a UUID
					if err != nil {
						return nil, err
					}
					t := p.Args["txn"]                       // get the Transaction input out of the arguments
					txnMap, ok := t.(map[string]interface{}) // convert the input type to a Transaction Map
					if !ok {
						return nil, errors.New("unable to convert input object to Transaction record")
					}
					var txn = new(Transaction)        // instantiate Transaction
					mapstructure.Decode(txnMap, &txn) // destructure txnMap into a Transaction
					return txn.Save(_bankId)          // return the saved transaction
				},
			},
		},
	}
}

/*
Initialize the Boldly Go GraphQL Schema Instance

	Build all of the queries being exposed
	Build all of the mutation being exposed

	Utilize the queries and mutations to build the GraphQL Schema instance
*/
func (b *boldlyGoGraphQL) BuildSchema() graphql.Schema {
	b.buildQuery()    // build all queries
	b.buildMutation() // build all mutations
	// use the built queries and mutations to build the graphql schema config
	schemaConfig := graphql.SchemaConfig{
		Query:    graphql.NewObject(b.queries),
		Mutation: graphql.NewObject(b.mutations),
	}
	// build the graphql schema instance
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		panic(err)
	}
	b.schema = schema
	fmt.Println("GraphQL Schema Instance initialized")
	return b.schema
}
