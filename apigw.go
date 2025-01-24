package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/apigatewayv2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lambda"
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
		RouteKey: pulumi.String(routeKey),
		Target:   pulumi.Sprintf("integrations/%s", integrationId),
	})
}

// Helper function to create an OPTIONS route for CORS preflight in the API Gateway
func createOptionsRoute(ctx *pulumi.Context, apiGateway *apigatewayv2.Api, routeKey string) (*apigatewayv2.Route, error) {
	// Create and return the OPTIONS route for CORS preflight handling
	return apigatewayv2.NewRoute(ctx, routeKey, &apigatewayv2.RouteArgs{
		ApiId:    apiGateway.ID(),
		RouteKey: pulumi.String(routeKey),
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
func createApiGatewayComponents(ctx *pulumi.Context, lambdas []*lambda.Function, zoneId pulumi.StringOutput) (*apigatewayv2.Api, *apigatewayv2.DomainName, error) {
	apiGateway, err := apigatewayv2.NewApi(ctx, "wedding-api", &apigatewayv2.ApiArgs{
		Name:                      pulumi.String("wedding-api"),
		ProtocolType:              pulumi.String("HTTP"),
		DisableExecuteApiEndpoint: pulumi.Bool(true),
		CorsConfiguration: &apigatewayv2.ApiCorsConfigurationArgs{
			AllowMethods: pulumi.StringArray{
				pulumi.String("GET"),
				pulumi.String("POST"),
				pulumi.String("OPTIONS"),
			},
			AllowOrigins: pulumi.StringArray{
				pulumi.String("*"),
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
			MaxAge: pulumi.Int(3600), // Optional: Time to cache preflight responses (in seconds)
		},
	})
	if err != nil {
		return nil, nil, err
	}

	authLambda := lambdas[0]
	refreshTokenLambda := lambdas[1]

	// Create Lambda integrations
	authLambdaIntegration, err := createLambdaIntegration(ctx, apiGateway, authLambda, "auth-lambda-integration")
	if err != nil {
		return nil, nil, err
	}
	refreshTokenLambdaIntegration, err := createLambdaIntegration(ctx, apiGateway, refreshTokenLambda, "refresh-lambda-integration")
	if err != nil {
		return nil, nil, err
	}

	// Create POST Routes
	_, err = createPostRoute(ctx, apiGateway, "POST /auth", authLambdaIntegration.ID())
	if err != nil {
		return nil, nil, err
	}
	_, err = createPostRoute(ctx, apiGateway, "POST /refresh", refreshTokenLambdaIntegration.ID())
	if err != nil {
		return nil, nil, err
	}
	// Create OPTIONS routes for CORS preflight handling
	_, err = createOptionsRoute(ctx, apiGateway, "OPTIONS /auth")
	if err != nil {
		return nil, nil, err
	}
	_, err = createOptionsRoute(ctx, apiGateway, "OPTIONS /refresh")
	if err != nil {
		return nil, nil, err
	}

	// Create Lambda permissions for API Gateway
	_, err = createLambdaPermission(ctx, authLambda, apiGateway, "auth-lambda-api-gateway-permission")
	if err != nil {
		return nil, nil, err
	}
	_, err = createLambdaPermission(ctx, refreshTokenLambda, apiGateway, "refresh-lambda-api-gateway-permission")
	if err != nil {
		return nil, nil, err
	}

	// Create deployment
	_, err = apigatewayv2.NewDeployment(ctx, "deployment", &apigatewayv2.DeploymentArgs{
		ApiId: apiGateway.ID(),
	}, pulumi.DependsOn([]pulumi.Resource{authLambdaIntegration, refreshTokenLambdaIntegration, apiGateway}))
	if err != nil {
		return nil, nil, err
	}

	// Create API stage
	apiStage, err := apigatewayv2.NewStage(ctx, "api-stage", &apigatewayv2.StageArgs{
		ApiId:      apiGateway.ID(),
		Name:       pulumi.String("v1"),
		AutoDeploy: pulumi.Bool(true),
	})
	if err != nil {
		return nil, nil, err
	}

	// DNS configuration (optional)
	conf := config.New(ctx, "")
	apiDomainStr := conf.Get("api-domain")
	if apiDomainStr != "" {
		// Load DNS zone
		apiDomainName, err := configureDnsForApiGateway(ctx, apiDomainStr, zoneId)
		if err != nil {
			return nil, nil, err
		}
		err = mapDnsToApiGateway(ctx, apiDomainStr, apiDomainName, apiStage.ID(), apiGateway.ID(), zoneId)
		if err != nil {
			return nil, nil, err
		}
		customUrl := pulumi.Sprintf("https://%s/", apiDomainStr)
		ctx.Export("custom-url", customUrl)
		return apiGateway, apiDomainName, nil
	}

	return apiGateway, nil, nil
}
