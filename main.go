package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/amplify"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/apigatewayv2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func createAuthResources(ctx *pulumi.Context, authPassword string) (*lambda.Function, error) {
	authPasswordParam, err := createSSMParameter(ctx, authPassword)
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

func createAmplifyResources(ctx *pulumi.Context, frontEndDomain string) (*amplify.App, error) {
	app, err := createAmplifyApp(ctx)
	if err != nil {
		return nil, err
	}
	_, err = createAmplifyDomain(ctx, app, frontEndDomain)
	return app, err
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		conf := config.New(ctx, "")
		frontendDomain := conf.Require("frontend-domain")
		authPassword := conf.Require("authPassword")
		authLambda, err := createAuthResources(ctx, authPassword)
		if err != nil {
			return err
		}
		apiGateway, err := createApiGateway(ctx, authLambda)
		if err != nil {
			return err
		}

		frontEnd, err := createAmplifyResources(ctx, frontendDomain)
		if err != nil {

		}
		ctx.Export("api-url", apiGateway.ApiEndpoint)
		ctx.Export("frontend-url", frontEnd.DefaultDomain)
		return nil
	})
}
