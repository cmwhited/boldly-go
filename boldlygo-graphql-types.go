package main

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"
	"github.com/satori/go.uuid"
	"golang.org/x/net/context"
)

var (
	// OUTPUT TYPES
	AuthType = graphql.NewObject(graphql.ObjectConfig{
		Name: "Auth",
		Fields: graphql.Fields{
			"success":   &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
			"message":   &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"token":     &graphql.Field{Type: graphql.String},
			"expiresAt": &graphql.Field{Type: graphql.Float},
		},
	})
	UserType = graphql.NewObject(graphql.ObjectConfig{
		Name: "User",
		Fields: graphql.Fields{
			"email": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"name":  &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		},
	})
	BankType = graphql.NewObject(graphql.ObjectConfig{
		Name: "Bank",
		Fields: graphql.Fields{
			"owningUserId":  &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"bankId":        &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"bankName":      &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"accountNumber": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		},
	})
	BankAccountType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "BankAccount",
		Description: "The users Bank Account information",
		Fields: graphql.Fields{
			"bankId":         &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"accountId":      &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"accountName":    &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"accountType":    &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"last4":          &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"currentBalance": &graphql.Field{Type: graphql.Float},
			"activeCard": &graphql.Field{
				Type:        CardType,
				Description: "The Active Card associated with the BankAccount",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if a, ok := p.Source.(*BankAccount); ok {
						acctId, err := uuid.FromString(a.AccountId)
						if err != nil {
							return nil, err
						}
						return GetActiveAccountCard(acctId)
					}
					return nil, nil
				},
			},
			"transactions": &graphql.Field{
				Type:        graphql.NewList(TransactionType),
				Description: "A list of Transactions associated to the Account",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if a, ok := p.Source.(*BankAccount); ok {
						acctId, err := uuid.FromString(a.AccountId)
						if err != nil {
							return nil, err
						}
						return GetAccountTransactions(acctId)
					}
					return nil, nil
				},
			},
			"txnsConn": &graphql.Field{
				Type: TransactionConnection.ConnectionType,
				Args: relay.ConnectionArgs,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					// convert args map[string]interface into ConnectionArguments
					args := relay.NewConnectionArguments(p.Args)
					// get transactions for current account
					var txns []interface{}
					if a, ok := p.Source.(*BankAccount); ok {
						acctId, err := uuid.FromString(a.AccountId)
						if err != nil {
							return nil, err
						}
						transactions, err := GetAccountTransactions(acctId)
						if err != nil {
							return nil, err
						}
						for _, txn := range transactions {
							txns = append(txns, txn)
						}
					}
					// let relay library figure out the result, given
					// - the list of transactions for this bank account
					// - and the filter arguments (i.e. first, last, after, before)
					return relay.ConnectionFromArray(txns, args), nil
				},
			},
			"bank": &graphql.Field{
				Type:        BankType,
				Description: "The Bank record the Account Belongs to",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					tokenEmail, err := boldlygo.AuthService().ValidateToken(p.Context.Value("Authorization"))
					if err != nil || tokenEmail == nil {
						return nil, nil
					}
					if a, ok := p.Source.(*BankAccount); ok {
						bankId, err := uuid.FromString(a.BankId)
						if err != nil {
							return nil, err
						}
						return GetBank(tokenEmail.(string), bankId)
					}
					return nil, nil
				},
			},
		},
	})
	TransactionConnection = relay.ConnectionDefinitions(relay.ConnectionConfig{
		Name:     "Txn",
		NodeType: TransactionType,
	})
	CardType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "Card",
		Description: "A Debit/Credit Card record associated to a Users Bank Account",
		Fields: graphql.Fields{
			"accountId":   &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"cardId":      &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"last4":       &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"expiryMonth": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"expiryYear":  &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"cvv":         &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"active":      &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
		},
	})
	TransactionType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "Transaction",
		Description: "A Transaction record associated with the BankAccount",
		Fields: graphql.Fields{
			"id": relay.GlobalIDField("TxnType", func(obj interface{}, info graphql.ResolveInfo, ctx context.Context) (string, error) {
				return "transactionId", nil
			}),
			"accountId":       &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"transactionId":   &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"transactionDate": &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime)},
			"amount":          &graphql.Field{Type: graphql.NewNonNull(graphql.Float)},
			"transactionType": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"description":     &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"cardId":          &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"card": &graphql.Field{
				Type:        CardType,
				Description: "The Card associated with the Transaction",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if t, ok := p.Source.(*Transaction); ok {
						if t.CardId == nil {
							return nil, nil
						}
						cardId, err := uuid.FromString(*t.CardId)
						if err != nil {
							return nil, err
						}
						acctId, err := uuid.FromString(t.AccountId)
						if err != nil {
							return nil, err
						}
						return GetAccountCard(acctId, cardId)
					}
					return nil, nil
				},
			},
		},
	})
	// MUTATION INPUT TYPES
	UserInputType = graphql.NewInputObject(graphql.InputObjectConfig{
		Name: "UserInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"email": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
			"pwd":   &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
			"name":  &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
		},
	})
	BankAccountInputType = graphql.NewInputObject(graphql.InputObjectConfig{
		Name:        "BankAccountInput",
		Description: "The BankAccount input object to use to create/update a BankAccount record",
		Fields: graphql.InputObjectConfigFieldMap{
			"bankId":         &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
			"accountId":      &graphql.InputObjectFieldConfig{Type: graphql.String},
			"accountName":    &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
			"accountType":    &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
			"last4":          &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
			"currentBalance": &graphql.InputObjectFieldConfig{Type: graphql.Float},
		},
	})
	CardInputType = graphql.NewInputObject(graphql.InputObjectConfig{
		Name:        "CardInput",
		Description: "The Card input object to use to create/update a Card record",
		Fields: graphql.InputObjectConfigFieldMap{
			"accountId":   &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
			"cardId":      &graphql.InputObjectFieldConfig{Type: graphql.String},
			"last4":       &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
			"expiryMonth": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
			"expiryYear":  &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
			"cvv":         &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
			"active":      &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.Boolean)},
		},
	})
	TransactionInputType = graphql.NewInputObject(graphql.InputObjectConfig{
		Name:        "TransactionInput",
		Description: "The Transaction input object to use to save a Transaction record",
		Fields: graphql.InputObjectConfigFieldMap{
			"accountId":       &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
			"transactionId":   &graphql.InputObjectFieldConfig{Type: graphql.String},
			"transactionDate": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.DateTime)},
			"amount":          &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.Float)},
			"transactionType": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
			"description":     &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
			"cardId":          &graphql.InputObjectFieldConfig{Type: graphql.String},
		},
	})
)
