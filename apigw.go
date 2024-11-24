package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/apigatewayv2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lambda"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/route53"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func createApiGatewayComponents(ctx *pulumi.Context, authLambda *lambda.Function) (*apigatewayv2.Api, error) {
	apiGateway, err := apigatewayv2.NewApi(ctx, "wedding-api", &apigatewayv2.ApiArgs{
		Name:         pulumi.String("wedding-api"),
		ProtocolType: pulumi.String("HTTP"),
		CorsConfiguration: &apigatewayv2.ApiCorsConfigurationArgs{
			AllowMethods: pulumi.StringArray{
				pulumi.String("POST"),
			},
			AllowOrigins: pulumi.StringArray{
				pulumi.String("*"),
			},
		},
	})
	if err != nil {
		return nil, err
	}
	authLambdaIntegration, err := apigatewayv2.NewIntegration(ctx, "auth-lambda-integration", &apigatewayv2.IntegrationArgs{
		ApiId:                apiGateway.ID(),
		IntegrationType:      pulumi.String("AWS_PROXY"),
		IntegrationUri:       authLambda.Arn, // lambda arn,
		IntegrationMethod:    pulumi.String("POST"),
		PayloadFormatVersion: pulumi.String("2.0"),
	})
	if err != nil {
		return nil, err
	}

	_, err = apigatewayv2.NewRoute(ctx, "default-route", &apigatewayv2.RouteArgs{
		ApiId:    apiGateway.ID(),
		RouteKey: pulumi.String("POST /auth"),
		Target:   pulumi.Sprintf("integrations/%s", authLambdaIntegration.ID()),
	})
	if err != nil {
		return nil, err
	}

	_, err = lambda.NewPermission(ctx, "api-gateway-permission", &lambda.PermissionArgs{
		Action:    pulumi.String("lambda:InvokeFunction"),
		Function:  authLambda.Name,
		Principal: pulumi.String("apigateway.amazonaws.com"),
		SourceArn: pulumi.Sprintf("%s/*/*", apiGateway.ExecutionArn),
	})
	if err != nil {
		return nil, err
	}

	_, err = apigatewayv2.NewDeployment(ctx, "deployment", &apigatewayv2.DeploymentArgs{
		ApiId: apiGateway.ID(),
	}, pulumi.DependsOn([]pulumi.Resource{authLambdaIntegration, apiGateway}))
	if err != nil {
		return nil, err
	}

	apiStage, err := apigatewayv2.NewStage(ctx, "api-stage", &apigatewayv2.StageArgs{
		ApiId:      apiGateway.ID(),
		Name:       pulumi.String("v1"),
		AutoDeploy: pulumi.Bool(true),
	})
	if err != nil {
		return nil, err
	}
	conf := config.New(ctx, "")
	apiSubdomain := conf.Get("api-subdomain")
	if apiSubdomain != "" {
		// Load DNS zone
		dnsZone := conf.Require("dns-zone")
		zone, err := route53.LookupZone(ctx, &route53.LookupZoneArgs{Name: pulumi.StringRef(dnsZone)})
		if err != nil {
			return nil, err
		}
		apiDomainName, err := configureDns(ctx, apiSubdomain, zone.ZoneId)
		if err != nil {
			return nil, err
		}

		apiMapping, err := apigatewayv2.NewApiMapping(ctx,
			"api-domain-mapping",
			&apigatewayv2.ApiMappingArgs{
				ApiId:      apiGateway.ID(),
				DomainName: apiDomainName.DomainName,
				Stage:      apiStage.ID(),
			})
		if err != nil {
			return nil, err
		}
		customUrl := pulumi.Sprintf("https://%s/", apiMapping.DomainName)
		ctx.Export("custom-url", customUrl)

	}
	return apiGateway, err
}
