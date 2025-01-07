package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/amplify"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createAmplifyApp(ctx *pulumi.Context, buildSpecStr string, apiGatewayEndpoint pulumi.StringOutput) (*amplify.App, error) {
	frontEnd, err := amplify.NewApp(ctx, "wedding-frontend", &amplify.AppArgs{
		BuildSpec: pulumi.String(buildSpecStr),
		CacheConfig: &amplify.AppCacheConfigArgs{
			Type: pulumi.String("AMPLIFY_MANAGED"),
		},
		EnvironmentVariables: pulumi.StringMap{
			"VITE_API_URL": apiGatewayEndpoint,
		},
		CustomRules: amplify.AppCustomRuleArray{
			&amplify.AppCustomRuleArgs{
				Source: pulumi.String("/<*>"),
				Status: pulumi.String("404-200"),
				Target: pulumi.String("/index.html"),
			},
			&amplify.AppCustomRuleArgs{
				Source: pulumi.String("</^[^.]+$|\\.(?!(css|gif|ico|jpg|js|png|txt|svg|woff|ttf|map|json)$)([^.]+$)/>"),
				Status: pulumi.String("200"),
				Target: pulumi.String("/index.html"),
			},
		},
		Name:       pulumi.String("wedding-frontend"),
		Platform:   pulumi.String("WEB"),
		Repository: pulumi.String("https://github.com/af-tec/wedding-frontend"),
	}, pulumi.Protect(true))
	if err != nil {
		return nil, err
	}
	return frontEnd, nil
}
