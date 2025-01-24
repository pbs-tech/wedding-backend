package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/amplify"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createAmplifyApp(ctx *pulumi.Context, buildSpecStr string, apiGatewayEndpoint pulumi.StringOutput, githubUrl string) (*amplify.App, error) {
	policyDocument, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
		Statements: []iam.GetPolicyDocumentStatement{
			// Statement for CloudWatch Logs
			{
				Sid:     pulumi.StringRef("PushLogs"),
				Actions: []string{"logs:CreateLogStream", "logs:PutLogEvents"},
				Resources: []string{
					"arn:aws:logs:eu-west-2:087085463074:log-group:/aws/amplify/*:log-stream:*",
				},
			},
			// Statement for creating Log Groups
			{
				Sid:     pulumi.StringRef("CreateLogGroup"),
				Actions: []string{"logs:CreateLogGroup"},
				Resources: []string{
					"arn:aws:logs:eu-west-2:087085463074:log-group:/aws/amplify/*",
				},
			},
			// Statement for describing Log Groups
			{
				Sid:     pulumi.StringRef("DescribeLogGroups"),
				Actions: []string{"logs:DescribeLogGroups"},
				Resources: []string{
					"arn:aws:logs:eu-west-2:087085463074:log-group:*",
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	amplifyPolicy, err := iam.NewPolicy(ctx, "amplify-logging-policy", &iam.PolicyArgs{
		Name:   pulumi.String("AmplifySSRLoggingPolicy-3c2c9294-ef89-42f0-bb6b-af818b0a742b"),
		Path:   pulumi.String("/service-role/"),
		Policy: pulumi.String(policyDocument.Json),
	})
	if err != nil {
		return nil, err
	}
	assumeRolePolicyDoc, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
		Statements: []iam.GetPolicyDocumentStatement{
			{
				Actions: []string{"sts:AssumeRole"},
				Principals: []iam.GetPolicyDocumentStatementPrincipal{
					{
						Type: "Service",
						Identifiers: []string{
							"amplify.amazonaws.com",
						},
					},
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	amplifyRole, err := iam.NewRole(ctx, "wedding-frontend-amplify-role", &iam.RoleArgs{
		AssumeRolePolicy: pulumi.String(assumeRolePolicyDoc.Json),
		Description:      pulumi.String("The service role for AWS Amplify to handle web compute app logging."),
		ManagedPolicyArns: pulumi.StringArray{
			amplifyPolicy.Arn,
			pulumi.String("arn:aws:iam::aws:policy/service-role/AmplifyBackendDeployFullAccess"),
		},
		Name: pulumi.String("AmplifySSRLoggingRole-3c2c9294-ef89-42f0-bb6b-af818b0a742b"),
		Path: pulumi.String("/service-role/"),
	})
	if err != nil {
		return nil, err
	}
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
		IamServiceRoleArn: amplifyRole.Arn,
		Name:              pulumi.String("wedding-frontend"),
		Platform:          pulumi.String("WEB"),
		Repository:        pulumi.String(githubUrl),
	})
	if err != nil {
		return nil, err
	}
	return frontEnd, nil
}
