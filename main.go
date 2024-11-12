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

func createApiGateway(ctx *pulumi.Context, rsvpLambda *lambda.Function) (*apigatewayv2.Api, error) {
	apiGateway, err := createApiGatewayComponents(ctx, rsvpLambda)
	if err != nil {
		return nil, err
	}
	return apiGateway, err
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
		rsvpLambda, err := createRSVPResources(ctx, guestTableName)
		if err != nil {
			return err
		}
		apiGateway, err := createApiGateway(ctx, rsvpLambda)
		if err != nil {
			return err
		}
		ctx.Export("api-url", apiGateway.ApiEndpoint)
		return nil
	})
}
