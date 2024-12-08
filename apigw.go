package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/apigatewayv2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lambda"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/route53"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

// Helper function to create a Lambda integration for the API Gateway
func createLambdaIntegration(ctx *pulumi.Context, apiGateway *apigatewayv2.Api, lambdaFunction *lambda.Function, integrationName string) (*apigatewayv2.Integration, error) {
	return apigatewayv2.NewIntegration(ctx, integrationName, &apigatewayv2.IntegrationArgs{
		ApiId:                apiGateway.ID(),
		IntegrationType:      pulumi.String("AWS_PROXY"),
		IntegrationUri:       lambdaFunction.Arn,
		IntegrationMethod:    pulumi.String("POST"),
		PayloadFormatVersion: pulumi.String("2.0"),
	})
}

// Helper function to create a POST route for the API Gateway with Lambda integration
func createPostRoute(ctx *pulumi.Context, apiGateway *apigatewayv2.Api, routeKey string, integrationId pulumi.IDInput) (*apigatewayv2.Route, error) {
	// Create and return the POST route with Lambda integration
	return apigatewayv2.NewRoute(ctx, routeKey, &apigatewayv2.RouteArgs{
		ApiId:    apiGateway.ID(),
		RouteKey: pulumi.Sprintf("POST %s", routeKey),
		Target:   pulumi.Sprintf("integrations/%s", integrationId),
	})
}

// Helper function to create an OPTIONS route for CORS preflight in the API Gateway
func createOptionsRoute(ctx *pulumi.Context, apiGateway *apigatewayv2.Api, routeKey string) (*apigatewayv2.Route, error) {
	// Create and return the OPTIONS route for CORS preflight handling
	return apigatewayv2.NewRoute(ctx, routeKey, &apigatewayv2.RouteArgs{
		ApiId:    apiGateway.ID(),
		RouteKey: pulumi.Sprintf("OPTIONS %s", routeKey),
		Target:   pulumi.String("aws:proxy"), // CORS preflight route target
	})
}

// Helper function to create Lambda permission for API Gateway
func createLambdaPermission(ctx *pulumi.Context, lambdaFunction *lambda.Function, apiGateway *apigatewayv2.Api, permissionName string) (*lambda.Permission, error) {
	return lambda.NewPermission(ctx, permissionName, &lambda.PermissionArgs{
		Action:    pulumi.String("lambda:InvokeFunction"),
		Function:  lambdaFunction.Name,
		Principal: pulumi.String("apigateway.amazonaws.com"),
		SourceArn: pulumi.Sprintf("%s/*/*", apiGateway.ExecutionArn),
	})
}

// Main function to create API Gateway components
func createApiGatewayComponents(ctx *pulumi.Context, lambdas []*lambda.Function) (*apigatewayv2.Api, error) {
	apiGateway, err := apigatewayv2.NewApi(ctx, "wedding-api", &apigatewayv2.ApiArgs{
		Name:         pulumi.String("wedding-api"),
		ProtocolType: pulumi.String("HTTP"),
		CorsConfiguration: &apigatewayv2.ApiCorsConfigurationArgs{
			AllowMethods: pulumi.StringArray{
				pulumi.String("POST"),
				pulumi.String("OPTIONS"),
			},
			AllowOrigins: pulumi.StringArray{
				pulumi.String("https://peebles.lol"),
			},
			AllowHeaders: pulumi.StringArray{
				pulumi.String("Content-Type"),
				pulumi.String("Authorization"),
				pulumi.String("Origin"),
			},
			ExposeHeaders: pulumi.StringArray{
				pulumi.String("Content-Type"),
				pulumi.String("Authorization"),
			},
			AllowCredentials: pulumi.Bool(true),
			MaxAge:           pulumi.Int(3600), // Optional: Time to cache preflight responses (in seconds)
		},
	})
	if err != nil {
		return nil, err
	}

	authLambda := lambdas[0]
	refreshTokenLambda := lambdas[1]

	// Create Lambda integrations
	authLambdaIntegration, err := createLambdaIntegration(ctx, apiGateway, authLambda, "auth-lambda-integration")
	if err != nil {
		return nil, err
	}
	refreshTokenLambdaIntegration, err := createLambdaIntegration(ctx, apiGateway, refreshTokenLambda, "refresh-lambda-integration")
	if err != nil {
		return nil, err
	}

	// Create POST Routes
	_, err = createPostRoute(ctx, apiGateway, "/auth", authLambdaIntegration.ID())
	if err != nil {
		return nil, err
	}
	_, err = createPostRoute(ctx, apiGateway, "/refresh", refreshTokenLambdaIntegration.ID())
	if err != nil {
		return nil, err
	}
	// Create OPTIONS routes for CORS preflight handling
	_, err = createOptionsRoute(ctx, apiGateway, "/auth")
	if err != nil {
		return nil, err
	}
	_, err = createOptionsRoute(ctx, apiGateway, "/refresh")
	if err != nil {
		return nil, err
	}

	// Create Lambda permissions for API Gateway
	_, err = createLambdaPermission(ctx, authLambda, apiGateway, "auth-lambda-api-gateway-permission")
	if err != nil {
		return nil, err
	}
	_, err = createLambdaPermission(ctx, refreshTokenLambda, apiGateway, "refresh-lambda-api-gateway-permission")
	if err != nil {
		return nil, err
	}

	// Create deployment
	_, err = apigatewayv2.NewDeployment(ctx, "deployment", &apigatewayv2.DeploymentArgs{
		ApiId: apiGateway.ID(),
	}, pulumi.DependsOn([]pulumi.Resource{authLambdaIntegration, apiGateway}))
	if err != nil {
		return nil, err
	}

	// Create API stage
	apiStage, err := apigatewayv2.NewStage(ctx, "api-stage", &apigatewayv2.StageArgs{
		ApiId:      apiGateway.ID(),
		Name:       pulumi.String("v1"),
		AutoDeploy: pulumi.Bool(true),
	})
	if err != nil {
		return nil, err
	}

	// DNS configuration (optional)
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

		// Configure domain mapping
		_, err = apigatewayv2.NewApiMapping(ctx,
			"api-domain-mapping",
			&apigatewayv2.ApiMappingArgs{
				ApiId:      apiGateway.ID(),
				DomainName: apiDomainName.DomainName,
				Stage:      apiStage.ID(),
			})
		if err != nil {
			return nil, err
		}

		customUrl := pulumi.Sprintf("https://%s/", apiSubdomain)
		ctx.Export("custom-url", customUrl)
	}

	return apiGateway, nil
}
