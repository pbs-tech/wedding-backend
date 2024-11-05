package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Guest struct {
	guestName    string                 `dynamodbav:"guestName"`
	hasPlusOne   bool                   `dynamodbav:"hasPlusOne"`
	isDayGuest   bool                   `dynamodbav:"isDayGuest"`
	rsvpResponse map[string]interface{} `dynamodbav:"rsvpResponse"`
}

var dynamoDbObj *dynamodb.Client
var tableName string

func updateGuestRsvpResponse(ctx context.Context, guest Guest) (map[string]map[string]interface{}, error) {
	var err error
	var response *dynamodb.UpdateItemOutput
	var attributeMap map[string]map[string]interface{}
	update := expression.Set(expression.Name("rsvpResponse.attending"), expression.Value(guest.rsvpResponse["attending"]))
	update.Set(expression.Name("rsvpResponse.songRequest"), expression.Value(guest.rsvpResponse["songRequest"]))
	update.Set(expression.Name("rsvpResponse.dietaryReqs"), expression.Value(guest.rsvpResponse["dietaryReqs"]))
	if guest.hasPlusOne {
		update.Set(expression.Name("rsvpResponse.plusOneName"), expression.Value(guest.rsvpResponse["plusOneName"]))
	}
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		log.Printf("Couldn't build expression for update. Here's why: %v\n", err)
	} else {
		response, err = dynamoDbObj.UpdateItem(ctx, &dynamodb.UpdateItemInput{
			TableName:                 aws.String(tableName),
			Key:                       guest.GetKey(),
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
			UpdateExpression:          expr.Update(),
			ReturnValues:              types.ReturnValueUpdatedNew,
		})
		if err != nil {
			log.Printf("Could not update guest %v RSVP response. Here's why: %v", guest.guestName, err)
		} else {
			err = attributevalue.UnmarshalMap(response.Attributes, &attributeMap)
			if err != nil {
				log.Printf("Cloud not unmarshall update response. Here's why %v\n", err)
			}
		}
	}
	return attributeMap, err
}

func getGuestDetails(ctx context.Context, guestName string) (Guest, error) {
	guest := Guest{guestName: guestName}
	response, err := dynamoDbObj.GetItem(ctx, &dynamodb.GetItemInput{
		Key: guest.GetKey(), TableName: aws.String(tableName),
	})
	if err != nil {
		log.Printf("Could not get info about %v. Here's why: %v\n", guestName, err)
	} else {
		err = attributevalue.UnmarshalMap(response.Item, &guest)
		if err != nil {
			log.Printf("Could not unmarshal response. Here's why: %v\n", err)
		}
	}
	return guest, err

}

func (guest Guest) GetKey() map[string]types.AttributeValue {
	guestName, err := attributevalue.Marshal(guest.guestName)
	if err != nil {
		panic(err)
	}
	return map[string]types.AttributeValue{"guestName": guestName}
}

func handleRequest(ctx context.Context) error {
	guest, err := getGuestDetails(ctx, guestName)
	if err != nil {
		return err
	}
	_, err = updateGuestRsvpResponse(ctx, guest)
	if err != nil {
		return err
	}
}

func main() {
	tableName = os.Getenv("DYNAMODB_TABLE_NAME")
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	dynamoDbObj = dynamodb.NewFromConfig(cfg)
	lambda.Start(handleRequest)
}
