package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/dynamodb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createDynamodbTable(ctx *pulumi.Context, dynamoDbTableName string) (*dynamodb.Table, error) {
	return dynamodb.NewTable(ctx, dynamoDbTableName, &dynamodb.TableArgs{
		Name:        pulumi.String(dynamoDbTableName),
		BillingMode: pulumi.String("PAY_PER_REQUEST"),
		Attributes: dynamodb.TableAttributeArray{
			&dynamodb.TableAttributeArgs{
				Name: pulumi.String("guestName"),
				Type: pulumi.String("S"),
			},
		},
		HashKey:  pulumi.String("guestName"),
		RangeKey: pulumi.String("guestName"),
	})
}
