/*
AWS Configuration/Initialization.

Instantiates a session with the AWS SDK for use and opens/exposes a connection to a DynamoDB instance.
*/
package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type AwsConfig interface {
	Init()
	DynamoDbSvc() *dynamodb.DynamoDB
}

type awsConf struct {
	dynamodbSvc *dynamodb.DynamoDB
}

/*
Initialize the AWS Service.

	Uses the AWS_ACCESS_KEY & AWS_SECRET_KEY values stored in the environment to connect to the AWS Account.

	Once the credentials are loaded, instantiate a new DynamoDB service instance
*/
func (c *awsConf) Init() {
	// establish the aws config with the env access key and secret
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		panic(err)
	}
	cfg.Region = endpoints.UsEast1RegionID
	// use config to build dynamodb svc
	c.dynamodbSvc = dynamodb.New(cfg)
	fmt.Println("AWS Service Initiated")
}

// Expose the DynamoDb service instance
func (c *awsConf) DynamoDbSvc() *dynamodb.DynamoDB {
	return c.dynamodbSvc
}
