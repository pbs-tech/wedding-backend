package main

import (
	"os"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/amplify"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/apigatewayv2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lambda"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ssm"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func createLambdaResources(ctx *pulumi.Context, authPasswordValue string, jwtSigingSecret string) ([]*lambda.Function, error) {
	authPasswordParam, err := createSSMParameter(ctx,
		"authPassword",
		"Password for users to use to access the site",
		authPasswordValue,
	)
	if err != nil {
		return nil, err
	}
	jwtSecretParam, err := createSSMParameter(ctx,
		"jwtSecret",
		"Secret used to sign JWT Tokens",
		jwtSigingSecret,
	)
	if err != nil {
		return nil, err
	}
	params := []*ssm.Parameter{authPasswordParam, jwtSecretParam}
	if err != nil {
		return nil, err
	}
	authLambda, err := createLambda(ctx,
		"auth",
		"./bin/auth.zip",
		pulumi.StringMap{
			"AUTH_PASSWORD_PARAM":      params[0].Name,
			"JWT_SIGNING_SECRET_PARAM": params[1].Name,
		},
	)
	if err != nil {
		return nil, err
	}
	refreshTokenLambda, err := createLambda(ctx,
		"refresh",
		"./bin/refresh.zip",
		pulumi.StringMap{
			"JWT_SIGNING_SECRET_PARAM": params[1].Name,
		})
	if err != nil {
		return nil, err
	}
	return []*lambda.Function{authLambda, refreshTokenLambda}, err
}

func createApiGateway(ctx *pulumi.Context, lambdas []*lambda.Function) (*apigatewayv2.Api, error) {
	apiGateway, err := createApiGatewayComponents(ctx, lambdas)
	if err != nil {
		return nil, err
	}
	return apiGateway, err
}

func createAmplifyResources(ctx *pulumi.Context, frontEndDomain string, frontEndBuildSpecStr string, apiEndpoint pulumi.StringOutput) (*amplify.App, error) {
	app, err := createAmplifyApp(ctx, frontEndBuildSpecStr, apiEndpoint)
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
		jwtSecret := conf.Require("jwtSecret")

		lambdas, err := createLambdaResources(ctx, authPassword, jwtSecret)
		if err != nil {
			return err
		}
		apiGateway, err := createApiGateway(ctx, lambdas)
		if err != nil {
			return err
		}

		frontendBuildSpec, err := os.ReadFile("npm.yaml")
		if err != nil {
			return err
		}
		frontendBuildSpecStr := string(frontendBuildSpec)
		frontEnd, err := createAmplifyResources(ctx, frontendDomain, frontendBuildSpecStr, apiGateway.ApiEndpoint)
		if err != nil {
			return err
		}
		ctx.Export("api-url", apiGateway.ApiEndpoint)
		ctx.Export("frontend-url", frontEnd.DefaultDomain)
		return nil
	})
}
