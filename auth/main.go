package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type userModelAuth struct {
	guestName string `dynamodbav:"guestName"`
	authToken string `dynamodbav:"authToken"`

	hasPlusOne bool `dynamodbav:"hasPlusOne"`
	isDayGuest bool `dynamodbav:"isDayGuest"`
}

var dynamoDbObj *dynamodb.Client
var tableName string

func isAuthorised(u string, token string) (userModelAuth, bool) {
	var user userModelAuth
	resp, err := dynamoDbObj.GetItem(context.TODO(), &dynamodb.GetItemInput{Key: getDynamoKeys(u), TableName: aws.String(tableName)})
	if err != nil {
		fmt.Printf("Cannot get item from table %s: %v\n", tableName, err)
		return userModelAuth{}, false
	} else {
		err = attributevalue.UnmarshalMap(resp.Item, &user)
		if err != nil {
			fmt.Printf("Could not unmarshal response: %v", err)
		}
	}
	if user.authToken == token {
		return user, true
	}
	return userModelAuth{}, false
}

func getDynamoKeys(username string) map[string]types.AttributeValue {
	encoded, err := attributevalue.Marshal(username)
	if err != nil {
		panic(err)
	}
	return map[string]types.AttributeValue{"guestName": encoded}
}

func handleRequest(ctx context.Context, apiGatewayRequest events.APIGatewayProxyRequest) (events.APIGatewayV2CustomAuthorizerSimpleResponse, error) {
	if user, isAuth := isAuthorised(apiGatewayRequest.Headers["guestName"], apiGatewayRequest.Headers["authToken"]); isAuth {
		m := make(map[string]interface{})
		m["hasPlusOne"] = user.hasPlusOne
		m["isDayGuest"] = user.isDayGuest
		m["complete"] = user
		fmt.Println(user)
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{IsAuthorized: true,
			Context: m}, nil
	}

	return events.APIGatewayV2CustomAuthorizerSimpleResponse{IsAuthorized: false}, nil
}

func main() {
	tableName = os.Getenv("DYNAMODB_TABLE_NAME")
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	dynamoDbObj = dynamodb.NewFromConfig(cfg)
	lambda.Start(handleRequest)
}
