package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/dynamodb"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createGuestResources(ctx *pulumi.Context, guestTableName string) (*dynamodb.Table, *lambda.Function, error) {
	guestTable, err := createDynamodbTable(ctx, guestTableName)
	if err != nil {
		return nil, nil, err
	}
	authLambda, err := createAuthLambda(ctx, guestTableName)
	if err != nil {
		return nil, nil, err
	}
	return guestTable, authLambda, err
}

func createRSVPResources(ctx *pulumi.Context, guestTableName string) (*dynamodb.Table, *lambda.Function, error) {
	guestTable, err := createDynamodbTable(ctx, guestTableName)
	if err != nil {
		return nil, nil, err
	}
	rsvpLambda, err := createRSVPLambda(ctx, guestTableName)
	if err != nil {
		return nil, nil, err
	}
	return guestTable, rsvpLambda, err
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// _, err := aws.GetCallerIdentity(ctx)
		// if err != nil {
		// 	return err
		// }
		// _, err := aws.GetRegion(ctx, &aws.GetRegionArgs{})
		// if err != nil {
		// 	return err
		// }

		guestTableName := "wedding-guests"
		_, _, err := createGuestResources(ctx, guestTableName)
		if err != nil {
			return err
		}
		rsvpTableName := "rsvp-responses"
		_, _, err = createRSVPResources(ctx, rsvpTableName)
		if err != nil {
			return err
		}
		return nil
	})
}
