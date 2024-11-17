package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/apigatewayv2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createAuthResources(ctx *pulumi.Context) (*lambda.Function, error) {
	authPasswordParam, err := createSSMParameter(ctx)
	if err != nil {
		return nil, err
	}
	authLambda, err := createAuthLambda(ctx, authPasswordParam)
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
