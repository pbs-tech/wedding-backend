package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type userModelAuth struct {
	guestName string `dynamodbav:"guestName"`
	authToken string `dynamodbav:"authToken"`

	hasPlusOne bool `dynamodbav:"hasPlusOne"`
	isDayGuest bool `dynamodbav:"isDayGuest"`
}

var dynamoDbObj *dynamodb.Client
var tableName string

func handleRequest() (string, error) {
	return "Hello there my fried friend", nil

}

func main() {
	tableName = os.Getenv("DYNAMODB_TABLE_NAME")
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	dynamoDbObj = dynamodb.NewFromConfig(cfg)
	lambda.Start(handleRequest)
}
