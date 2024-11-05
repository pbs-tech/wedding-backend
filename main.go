package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/apigatewayv2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createGuestResources(ctx *pulumi.Context, guestTableName string) (*lambda.Function, error) {
	authLambda, err := createAuthLambda(ctx, guestTableName)
	if err != nil {
		return nil, err
	}
	return authLambda, err
}

func createRSVPResources(ctx *pulumi.Context, guestTableName string) (*lambda.Function, error) {
	rsvpLambda, err := createRSVPLambda(ctx, guestTableName)
	if err != nil {
		return nil, err
	}
	return rsvpLambda, err
}

func createApiGateway(ctx *pulumi.Context) (*apigatewayv2.Api, error) {
	apiGway, err := createApiGatewayComponents(ctx)
	if err != nil {
		return nil, err
	}
	return apiGway, err
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
		_, err := createDynamodbTable(ctx, guestTableName)
		if err != nil {
			return err
		}
		_, err = createGuestResources(ctx, guestTableName)
		if err != nil {
			return err
		}
		_, err = createRSVPResources(ctx, guestTableName)
		if err != nil {
			return err
		}
		apiGateway, err = createApiGateway(ctx)
		if err != nil {
			return err
		}
		ctx.Export("api-url", apiGateway.Url)
		return nil
	})
}
