package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/apigatewayv2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createAuthResources(ctx *pulumi.Context) (*lambda.Function, error) {
	authPassword, err := createSSMParameter(ctx)
	if err != nil {
		return nil, err
	}
	authLambda, err := createAuthLambda(ctx, authPassword)
	if err != nil {
		return nil, err
	}
	return authLambda, err
}

func createApiGateway(ctx *pulumi.Context, authLambda *lambda.Function) (*apigatewayv2.Api, error) {
	apiGateway, err := createApiGatewayComponents(ctx, authLambda)
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
		authLambda, err := createAuthResources(ctx)
		if err != nil {
			return err
		}
		apiGateway, err := createApiGateway(ctx, authLambda)
		if err != nil {
			return err
		}
		ctx.Export("api-url", apiGateway.ApiEndpoint)
		return nil
	})
}
