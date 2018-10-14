# Boldly Go

GoLang with `go dep` dependency management and GraphQL.

This service utilizes the [graphql-go](https://github.com/graphql-go/graphql) library to build and expose a GraphQL schema
instance. The service is exposed at:

- `/graphql`

## Data Storage

This service uses DynamoDB to persist data. Check out the [docs](https://aws.amazon.com/dynamodb/) for more information.

### AWS Access

To access your AWS DynamoDB tables, you will need an AWS account with an IAM user that has access to Read, Write DynamoDB tables.
Once the IAM user is created, get the access key and secret and store them in environment variables:

- `AWS_ACCESS_KEY_ID`: The IAM user access key
- `AWS_SECRET_KEY`: The IAM user secret

## Dependency Management

This service uses [go dep](https://github.com/golang/dep) for the dependency management tool. After pulling the code down,
run `dep ensure`; this will install necessary dependencies to the project and get it ready for running.

## GraphQL

Checkout the [graphql docs](https://graphql.org/) to get an understanding of the Spec and how it is used and implemented.

### Queries

List of the queries exposed by the service:
    - `bankAccounts`: Get a list of the users BankAccount records by the Bank primary key
    - `bankAccount`: Get a unique user BankAccount record by the BankId Primary Key and Account Id
    - `accountCards`: A list of cards associated to the BankAccount
    - `accountCard`: A BankAccount Card record
    - `accountTransaction`: A BankAccount Transaction record
    
### Mutations

List of the mutations exposed by the service:
    - `saveBankAccount`: Save a new BankAccount record
    - `updateBankAccount`: Update a BankAccount record
    - `saveAccountCard`: Save a new BankAccount Card record
    - `inactivateAccountCard`: Inactivate a Bank Account Card record
    - `saveTransaction`: Save a Transaction record
    
## Running Queries/Mutations

You can use `curl` to run a GraphQL Query/Mutation by submitting a `POST` request to the graphql endpoing (`/graphql`) 
with the query you are trying to run as the data. This is not the easiest way. Here are a couple other ways:

- `Graphiql`: This service comes bundled with a graphiql instance. Graphiql is essentially a web-based IDE for introspecting
and running GraphQL queries/mutations. To access the graphiql instance, once the service is running: open a web browser,
navigate to `http://localhost:5000/graphql`.
- `Insomnia`: very similar to Postman, Insomnia is a REST client. In addition to traditional REST commands, Insomnia
also allows you to specify that you want to run a GraphQL query/mutation and provides intellisense and schema inspection
against your GraphQL service. Check out Insomnia [here](https://insomnia.rest/).

