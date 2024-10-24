package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/dynamodb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createDynamodbTable(ctx *pulumi.Context) (*dynamodb.Table, error) {
	return dynamodb.NewTable(ctx, "wedding-guests", &dynamodb.TableArgs{
		Name:        pulumi.String("wedding-guests"),
		BillingMode: pulumi.String("PAY_PER_REQUEST"),
		Attributes: dynamodb.TableAttributeArray{
			&dynamodb.TableAttributeArgs{
				Name: pulumi.String("guestId"),
				Type: pulumi.String("S"),
			},
			&dynamodb.TableAttributeArgs{
				Name: pulumi.String("guestName"),
				Type: pulumi.String("S"),
			},
		},
		HashKey:  pulumi.String("guestId"),
		RangeKey: pulumi.String("guestName"),
	})
}
