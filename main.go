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

func createLambdaResources(ctx *pulumi.Context, authPasswordValue string, jwtSigingSecret string, frontendDomain string) ([]*lambda.Function, error) {
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

	authLambda, err := createLambda(ctx,
		"auth",
		"./bin/auth.zip",
		pulumi.StringMap{
			"AUTH_PASSWORD_PARAM":      params[0].Name,
			"JWT_SIGNING_SECRET_PARAM": params[1].Name,
			"FRONTEND_DOMAIN":          pulumi.String(frontendDomain),
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
			"FRONTEND_DOMAIN":          pulumi.String(frontendDomain),
		})
	if err != nil {
		return nil, err
	}
	return []*lambda.Function{authLambda, refreshTokenLambda}, err
}

func createApiGateway(ctx *pulumi.Context, lambdas []*lambda.Function, frontendURL string) (*apigatewayv2.Api, *apigatewayv2.DomainName, error) {
	apiGateway, apiDomainName, err := createApiGatewayComponents(ctx, lambdas, frontendURL)
	if err != nil {
		return nil, nil, err
	}
	return apiGateway, apiDomainName, err
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
		frontendURL := "https://" + frontendDomain
		lambdas, err := createLambdaResources(ctx, authPassword, jwtSecret, frontendDomain)
		if err != nil {
			return err
		}
		apiGateway, apiDomainName, err := createApiGateway(ctx, lambdas, frontendURL)
		if err != nil {
			return err
		}
		apiUrl := apiGateway.ApiEndpoint
		if apiDomainName != nil {
			apiUrl = pulumi.Sprintf("https://%s", apiDomainName.DomainName)
		}
		frontendBuildSpec, err := os.ReadFile("npm.yaml")
		if err != nil {
			return err
		}
		frontendBuildSpecStr := string(frontendBuildSpec)
		frontEnd, err := createAmplifyResources(ctx, frontendDomain, frontendBuildSpecStr, apiUrl)
		if err != nil {
			return err
		}

		ctx.Export("api-url", apiUrl)
		ctx.Export("frontend-url", frontEnd.DefaultDomain)
		return nil
	})
}
