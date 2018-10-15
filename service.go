package main

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"net/http"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/expression"
	"github.com/satori/go.uuid"
)

const (
	bankUrl = "http://localhost:5002/api/v1/user/4b7b2def-e76e-48bf-993b-8ec2b193b855/bank/{bankId}"
)

/*
Register a new User.
Hash the password before storing.
Return the created User record.
*/
func (u *User) Register() (*User, error) {
	hashedPwd, err := boldlygo.AuthService().HashPwd(u.Pwd) // use the AuthSvc to hash the users password
	if err != nil {
		return nil, err
	}
	u.Pwd = *hashedPwd                              // set new hashed password on user
	userMap, err := dynamodbattribute.MarshalMap(u) // marshal User to dynamodbattribute map
	if err != nil {
		return nil, err
	}
	// build item input request
	input := &dynamodb.PutItemInput{
		Item:      userMap,
		TableName: aws.String("Users"),
	}
	req := boldlygo.DynamoDbSvc().PutItemRequest(input) // save item to db
	_, err = req.Send()
	if err != nil {
		return nil, err
	}
	return u, nil
}

/*
Authenticate a user by their email and password.

	Attempt to find the user by the email.
		- If the user is found; get their hashed password, use the AuthSvc to compare it to the passed in password:
			- if the passwords match, generate a JWT and return
			- If the passwords do not match, return an error
		- If the user cannot be found, return an error
*/
func Authenticate(email, pwd string) Auth {
	req := boldlygo.DynamoDbSvc().GetItemRequest(&dynamodb.GetItemInput{
		TableName: aws.String("Users"),
		Key: map[string]dynamodb.AttributeValue{
			"email": {
				S: aws.String(email),
			},
		},
	}) // build the request to send to DynamoDB to find a unique user record by the email primary key
	output, err := req.Send() // send the request to the DynamoDB service; get the output result
	if err != nil {
		return Auth{
			Success: false,
			Message: "Unable to find user by that email. Please check your email and try again",
		}
	}
	if len(output.Item) == 0 {
		return Auth{
			Success: false,
			Message: "Unable to find user by that email. Please check your email and try again",
		}
	}
	// unmarshal returned map from DynamoDB into a User
	var user = new(User)
	err = dynamodbattribute.UnmarshalMap(output.Item, &user)
	if err != nil {
		return Auth{
			Success: false,
			Message: err.Error(),
		}
	}
	// verify that the passed in password matches the saved password for the user
	if verify := boldlygo.AuthService().VerifyPwd(user.Pwd, pwd); !verify {
		return Auth{
			Success: false,
			Message: "The password submitted does not match this users password. Please check the email and password and try again",
		}
	}
	token, expiry, err := boldlygo.AuthService().BuildToken(*user) // generate token from user
	if err != nil {
		return Auth{
			Success: false,
			Message: err.Error(),
		}
	}
	return Auth{
		Success:   true,
		Message:   "Success",
		Token:     *token,
		ExpiresAt: *expiry,
	}
}

/*
Utilize the HTTP client to make a REST call to get the Bank info by its PK id
*/
func GetBank(bankId uuid.UUID) (*Bank, error) {
	url := strings.Replace(bankUrl, "{bankId}", bankId.String(), -1) // build url
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// get the response body and parse into Bank
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var bank = new(Bank)
	json.Unmarshal(body, &bank) // unmarshal the response body into a bank
	return bank, nil
}

/*
Get a list of all of the users bank accounts by the bank id
*/
func GetUserBankAccounts(bankId uuid.UUID) ([]*BankAccount, error) {
	keyCond := expression.Key("bankId").Equal(expression.Value(bankId.String())) // build find BankAccount by BankId filter expression
	expr, err := expression.NewBuilder().
		WithKeyCondition(keyCond).
		Build()
	if err != nil {
		return nil, err
	}
	params := &dynamodb.QueryInput{
		TableName:                 aws.String("BankAccounts"),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeValues: expr.Values(),
		ExpressionAttributeNames:  expr.Names(),
	}
	req := boldlygo.DynamoDbSvc().QueryRequest(params) // build dynamodb query with key condition
	output, err := req.Send()                          // submit the dynamodb query request
	if err != nil {
		return nil, err
	}
	// unmarshal the return into the object
	if output.Items == nil {
		return nil, err
	}
	var accounts = make([]*BankAccount, 0)
	err = dynamodbattribute.UnmarshalListOfMaps(output.Items, &accounts) // unmarshal the found items into a list of accounts
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

/*
Get a unique BankAccount record by the Primary Key and Sort Key conditions
*/
func GetUserBankAccount(bankId, accountId uuid.UUID) (*BankAccount, error) {
	req := boldlygo.DynamoDbSvc().GetItemRequest(&dynamodb.GetItemInput{
		TableName: aws.String("BankAccounts"),
		Key: map[string]dynamodb.AttributeValue{
			"bankId": {
				S: aws.String(bankId.String()),
			},
			"accountId": {
				S: aws.String(accountId.String()),
			},
		},
	})
	output, err := req.Send()
	if err != nil {
		return nil, err
	}
	// unmarshal returned map into BankAccount
	var account = new(BankAccount)
	err = dynamodbattribute.UnmarshalMap(output.Item, &account)
	if err != nil {
		return nil, err
	}
	return account, nil
}

/*
Save a new BankAccount record to DynamoDB
*/
func (a *BankAccount) Save() (*BankAccount, error) {
	a.AccountId = uuid.NewV4().String()             // set unique account id
	acctMap, err := dynamodbattribute.MarshalMap(a) // marshal BankAccount to dynamodbattribute map
	if err != nil {
		return nil, err
	}
	// build item input request
	input := &dynamodb.PutItemInput{
		Item:      acctMap,
		TableName: aws.String("BankAccounts"),
	}
	// save item to db
	req := boldlygo.DynamoDbSvc().PutItemRequest(input)
	_, err = req.Send()
	if err != nil {
		return nil, err
	}
	return a, nil
}

/*
Update a BankAccount record in DynamoDB
*/
func (a *BankAccount) Update() (*BankAccount, error) {
	// Build Update expression to set which fields should be updated
	update := expression.
		Set(expression.Name("accountName"), expression.Value(a.AccountName)).
		Set(expression.Name("accountType"), expression.Value(a.AccountType)).
		Set(expression.Name("last4"), expression.Value(a.Last4)).
		Set(expression.Name("currentBalance"), expression.Value(a.CurrentBalance))
	// build update expression with update fields set
	expr, err := expression.NewBuilder().
		WithUpdate(update).
		Build()
	if err != nil {
		return nil, err
	}
	// build update BankAccount item input
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String("BankAccounts"),
		Key: map[string]dynamodb.AttributeValue{
			"bankId": {
				S: aws.String(a.BankId),
			},
			"accountId": {
				S: aws.String(a.AccountId),
			},
		},
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ReturnValues:              dynamodb.ReturnValueNone,
		UpdateExpression:          expr.Update(),
	}
	req := boldlygo.DynamoDbSvc().UpdateItemRequest(input) // build update item request
	_, err = req.Send()                                    // send update item request; expect nothing back
	if err != nil {
		return nil, err
	}
	return a, nil // return BankAccount
}

/*
Update a BankAccount record in DynamoDB.
Update the CurrentBalance as the result of a Transaction occurring on the BankAccount
*/
func (a *BankAccount) UpdateCurrentBalance(txnAmount float64, txnType string) error {
	// calculate the new Current Balance
	currBalance := a.CurrentBalance
	if txnType == "CREDIT" {
		txnAmount = math.Abs(txnAmount) * -1 // if transaction type is CREDIT, it needs to subtracted from the current balance of the BankAccount
	}
	currBalance += txnAmount
	// Build Update expression to set which fields should be updated
	update := expression.Set(expression.Name("currentBalance"), expression.Value(currBalance))
	// build update expression with update fields set
	expr, err := expression.NewBuilder().
		WithUpdate(update).
		Build()
	if err != nil {
		return err
	}
	// build update BankAccount item input
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String("BankAccounts"),
		Key: map[string]dynamodb.AttributeValue{
			"bankId": {
				S: aws.String(a.BankId),
			},
			"accountId": {
				S: aws.String(a.AccountId),
			},
		},
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ReturnValues:              dynamodb.ReturnValueNone,
		UpdateExpression:          expr.Update(),
	}
	req := boldlygo.DynamoDbSvc().UpdateItemRequest(input) // build update item request
	_, err = req.Send()                                    // send update item request; expect nothing back
	if err != nil {
		return err
	}
	return nil
}

/*
Get a list of Cards associated to the BankAccount
*/
func GetAccountCards(accountId uuid.UUID) ([]*Card, error) {
	keyCond := expression.Key("accountId").Equal(expression.Value(accountId.String())) // build find Card records by AccountId filter expression
	expr, err := expression.NewBuilder().
		WithKeyCondition(keyCond).
		Build()
	if err != nil {
		return nil, err
	}
	params := &dynamodb.QueryInput{
		TableName:                 aws.String("Cards"),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeValues: expr.Values(),
		ExpressionAttributeNames:  expr.Names(),
	}
	req := boldlygo.DynamoDbSvc().QueryRequest(params) // build dynamodb query with key condition
	output, err := req.Send()                          // submit the dynamodb query request
	if err != nil {
		return nil, err
	}
	// unmarshal the return into the object
	if output.Items == nil {
		return nil, err
	}
	var cards = make([]*Card, 0)
	err = dynamodbattribute.UnmarshalListOfMaps(output.Items, &cards) // unmarshal the found items into a list of cards
	if err != nil {
		return nil, err
	}
	return cards, nil
}

/*
Find a unique Card record by the accountId, cardId composite key
*/
func GetAccountCard(accountId, cardId uuid.UUID) (*Card, error) {
	req := boldlygo.DynamoDbSvc().GetItemRequest(&dynamodb.GetItemInput{
		TableName: aws.String("Cards"),
		Key: map[string]dynamodb.AttributeValue{
			"accountId": {
				S: aws.String(accountId.String()),
			},
			"cardId": {
				S: aws.String(cardId.String()),
			},
		},
	})
	output, err := req.Send()
	if err != nil {
		return nil, err
	}
	// unmarshal returned map into Card
	var card = new(Card)
	err = dynamodbattribute.UnmarshalMap(output.Item, &card)
	if err != nil {
		return nil, err
	}
	return card, nil
}

/*
Find the Card record associated to the BankAccount that is marked as Active
*/
func GetActiveAccountCard(accountId uuid.UUID) (*Card, error) {
	filter := expression.
		Name("accountId").Equal(expression.Value(accountId.String())).
		And(expression.Name("active").Equal(expression.Value(true))) // build filter for account id and active true
	expr, err := expression.NewBuilder().
		WithFilter(filter).
		Build()
	if err != nil {
		return nil, err
	}
	params := &dynamodb.ScanInput{
		TableName:                 aws.String("Cards"),
		FilterExpression:          expr.Filter(),
		ExpressionAttributeValues: expr.Values(),
		ExpressionAttributeNames:  expr.Names(),
	}
	req := boldlygo.DynamoDbSvc().ScanRequest(params) // build dynamodb query with key condition
	output, err := req.Send()                         // submit the dynamodb query request
	if err != nil {
		return nil, err
	}
	// unmarshal the return into the object
	if output.Items == nil || len(output.Items) == 0 {
		return nil, err
	}
	var cards = make([]*Card, 0)
	err = dynamodbattribute.UnmarshalListOfMaps(output.Items, &cards) // unmarshal the found items into a list of cards
	if err != nil {
		return nil, err
	}
	return cards[0], nil
}

/*
Save a Card record
*/
func (c *Card) Save() (*Card, error) {
	c.CardId = uuid.NewV4().String()                // set unique card id
	cardMap, err := dynamodbattribute.MarshalMap(c) // marshal Card to dynamodbattribute map
	if err != nil {
		return nil, err
	}
	// build item input request
	input := &dynamodb.PutItemInput{
		Item:      cardMap,
		TableName: aws.String("Cards"),
	}
	// save item to db
	req := boldlygo.DynamoDbSvc().PutItemRequest(input)
	_, err = req.Send()
	if err != nil {
		return nil, err
	}
	return c, nil
}

/*
Update an existing Card record
*/
func (c *Card) Inactivate() (*Card, error) {
	// Set the active field on the card to false
	update := expression.Set(expression.Name("active"), expression.Value(false))
	// build update expression with update fields set
	expr, err := expression.NewBuilder().
		WithUpdate(update).
		Build()
	if err != nil {
		return nil, err
	}
	// build update BankAccount item input
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String("Cards"),
		Key: map[string]dynamodb.AttributeValue{
			"accountId": {
				S: aws.String(c.AccountId),
			},
			"cardId": {
				S: aws.String(c.CardId),
			},
		},
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ReturnValues:              dynamodb.ReturnValueNone,
		UpdateExpression:          expr.Update(),
	}
	req := boldlygo.DynamoDbSvc().UpdateItemRequest(input) // build update item request
	_, err = req.Send()                                    // send update item request; expect nothing back
	if err != nil {
		return nil, err
	}
	return c, nil // return Card
}

/*
Get a list of all Transactions associated to the BankAccount
*/
func GetAccountTransactions(accountId uuid.UUID) ([]*Transaction, error) {
	keyCond := expression.Key("accountId").Equal(expression.Value(accountId.String())) // build find Transaction records by AccountId filter expression
	expr, err := expression.NewBuilder().
		WithKeyCondition(keyCond).
		Build()
	if err != nil {
		return nil, err
	}
	params := &dynamodb.QueryInput{
		TableName:                 aws.String("Transactions"),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeValues: expr.Values(),
		ExpressionAttributeNames:  expr.Names(),
	}
	req := boldlygo.DynamoDbSvc().QueryRequest(params) // build dynamodb query with key condition
	output, err := req.Send()                          // submit the dynamodb query request
	if err != nil {
		return nil, err
	}
	// unmarshal the return into the object
	if output.Items == nil {
		return nil, err
	}
	var transactions = make([]*Transaction, 0)
	err = dynamodbattribute.UnmarshalListOfMaps(output.Items, &transactions) // unmarshal the found items into a list of transactions
	if err != nil {
		return nil, err
	}
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].TransactionDate.Before(transactions[j].TransactionDate)
	})
	return transactions, nil
}

/*
Find a unique BankAccount Transaction record by the accountId and transactionId composite key
*/
func GetAccountTransaction(accountId, transactionId uuid.UUID) (*Transaction, error) {
	req := boldlygo.DynamoDbSvc().GetItemRequest(&dynamodb.GetItemInput{
		TableName: aws.String("Transactions"),
		Key: map[string]dynamodb.AttributeValue{
			"accountId": {
				S: aws.String(accountId.String()),
			},
			"transactionId": {
				S: aws.String(transactionId.String()),
			},
		},
	})
	output, err := req.Send()
	if err != nil {
		return nil, err
	}
	// unmarshal returned map into Transaction
	var txn = new(Transaction)
	err = dynamodbattribute.UnmarshalMap(output.Item, &txn)
	if err != nil {
		return nil, err
	}
	return txn, nil
}

/*
Save a Transaction to the BankAccount.
Update the CurrentBalance on the BankAccount as a result of the Transaction
*/
func (t *Transaction) Save(bankId uuid.UUID) (*Transaction, error) {
	t.TransactionId = uuid.NewV4().String()        // set unique transaction id
	txnMap, err := dynamodbattribute.MarshalMap(t) // marshal Transaction to dynamodbattribute map
	if err != nil {
		return nil, err
	}
	// build item input request
	input := &dynamodb.PutItemInput{
		Item:      txnMap,
		TableName: aws.String("Transactions"),
	}
	// save item to db
	req := boldlygo.DynamoDbSvc().PutItemRequest(input)
	_, err = req.Send()
	if err != nil {
		return nil, err
	}
	// update the current balance of the BankAccount
	acctId, err := uuid.FromString(t.AccountId)
	if err != nil {
		return nil, err
	}
	// get the BankAccount record
	bankAccount, err := GetUserBankAccount(bankId, acctId)
	if err != nil {
		return nil, err
	}
	err = bankAccount.UpdateCurrentBalance(t.Amount, t.TransactionType)
	if err != nil {
		return nil, err
	}
	// return the Transaction
	return t, nil
}
