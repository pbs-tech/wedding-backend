package main

import (
	"os"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/amplify"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/apigatewayv2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/cloudfront"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lambda"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ssm"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func createLambdaResources(ctx *pulumi.Context, dayGuestPasswordValue string, eveningGuestPasswordValue string, jwtSigingSecret string, frontendDomain string) ([]*lambda.Function, error) {
	dayGuestPasswordParam, err := createSSMParameter(ctx,
		"dayGuestPassword",
		"Password for day guests to use to access the site",
		dayGuestPasswordValue,
	)
	if err != nil {
		return nil, err
	}
	eveningGuestPasswordParam, err := createSSMParameter(ctx,
		"eveningGuestPassword",
		"Password for evening guests to use to access the site",
		eveningGuestPasswordValue,
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
	params := []*ssm.Parameter{dayGuestPasswordParam, eveningGuestPasswordParam, jwtSecretParam}

	authLambda, err := createLambda(ctx,
		"auth",
		"./bin/auth.zip",
		pulumi.StringMap{
			"DAY_GUEST_PASSWORD_PARAM":     params[0].Name,
			"EVENING_GUEST_PASSWORD_PARAM": params[1].Name,
			"JWT_SIGNING_SECRET_PARAM":     params[2].Name,
			"FRONTEND_DOMAIN":              pulumi.String(frontendDomain),
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

func createApiGateway(ctx *pulumi.Context, lambdas []*lambda.Function, zoneId pulumi.StringOutput) (*apigatewayv2.Api, *apigatewayv2.DomainName, *cloudfront.Distribution, error) {
	apiGateway, apiDomainName, err := createApiGatewayComponents(ctx, lambdas, zoneId)
	if err != nil {
		return nil, nil, nil, err
	}
	distribution, err := createCloudfrontDistributionForApiGateway(ctx, "api-gateway", apiGateway)
	if err != nil {
		return apiGateway, apiDomainName, distribution, err
	}
	return apiGateway, apiDomainName, distribution, err
}

func createAmplifyResources(ctx *pulumi.Context, frontEndDomain string, frontEndBuildSpecStr string, apiEndpoint pulumi.StringOutput, githubUrl string) (*amplify.App, error) {
	app, err := createAmplifyApp(ctx, frontEndBuildSpecStr, apiEndpoint, githubUrl)
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
		apiDomain := conf.Require("api-domain")
		githubUrl := conf.Require("githubUrl")
		apiUrl := pulumi.Sprintf("https://%s", apiDomain)

		dayGuestPassword := conf.Require("dayGuestPassword")
		eveningGuestPassword := conf.Require("eveningGuestPassword")
		jwtSecret := conf.Require("jwtSecret")
		rootDnsZone, err := createDnsZone(ctx, frontendDomain)
		if err != nil {
			return err
		}
<<<<<<< HEAD
		lambdas, err := createLambdaResources(ctx, dayGuestPassword, eveningGuestPassword, jwtSecret, frontendDomain)
		if err != nil {
			return err
		}
		apiGateway, apiDomainName, distribution, err := createApiGateway(ctx, lambdas, rootDnsZone.ZoneId)
		if err != nil {
			return err
		}
		cloudfrontDomain := distribution.DomainName
		apiUrl := apiGateway.ApiEndpoint
		if apiDomainName != nil {
			apiUrl = pulumi.Sprintf("https://%s", apiDomainName.DomainName)
		}
=======
>>>>>>> 1e52aaf (fix: update github url)
		frontendBuildSpec, err := os.ReadFile("npm.yaml")
		if err != nil {
			return err
		}
		frontendBuildSpecStr := string(frontendBuildSpec)
		frontEnd, err := createAmplifyResources(ctx, frontendDomain, frontendBuildSpecStr, apiUrl, githubUrl)
		if err != nil {
			return err
		}
		lambdas, err := createLambdaResources(ctx, dayGuestPassword, eveningGuestPassword, jwtSecret, frontendDomain)
		if err != nil {
			return err
		}
		_, apiDomainName, err := createApiGateway(ctx, lambdas, rootDnsZone.ZoneId)
		if err != nil {
			return err
		}

		ctx.Export("api-url", apiDomainName.DomainName)
		ctx.Export("frontend-url", frontEnd.DefaultDomain)
		return nil
	})
}
