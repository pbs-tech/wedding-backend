package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createLambda(ctx *pulumi.Context, lambdaName string, path string, lambdaEnvVars pulumi.StringMap) (*lambda.Function, error) {
	role, lambdaPolicy, err := createLambdaIamRolePolicy(ctx, lambdaName)
	if err != nil {
		return nil, err
	}
	authLambda, err := lambda.NewFunction(ctx, lambdaName, &lambda.FunctionArgs{
		Runtime: lambda.RuntimeCustomAL2023,
		Code: pulumi.NewAssetArchive(map[string]interface{}{
			".": pulumi.NewFileArchive(path),
		}),
		Handler:       pulumi.String(lambdaName),
		Role:          role.Arn,
		Architectures: pulumi.StringArray{pulumi.String("arm64")},
		Timeout:       pulumi.Int(15),
		Environment: &lambda.FunctionEnvironmentArgs{
			Variables: lambdaEnvVars,
		},
	}, pulumi.DependsOn([]pulumi.Resource{lambdaPolicy}))
	if err != nil {
		return nil, err
	}
	return authLambda, nil
}

func createLambdaIamRolePolicy(ctx *pulumi.Context, lambdaName string) (*iam.Role, *iam.Policy, error) {

	// Define the Assume Role Policy for Lambda using getPolicyDocument
	assumeRolePolicyDoc, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
		Statements: []iam.GetPolicyDocumentStatement{
			{
				Actions: []string{"sts:AssumeRole"},
				Principals: []iam.GetPolicyDocumentStatementPrincipal{
					{
						Type: "Service",
						Identifiers: []string{
							"lambda.amazonaws.com",
						},
					},
				},
			},
		},
	})
	if err != nil {
		return nil, nil, err
	}

	// Create the IAM role for the Lambda function with the generated Assume Role Policy
	role, err := iam.NewRole(ctx, lambdaName+"-exec-role", &iam.RoleArgs{
		AssumeRolePolicy: pulumi.String(assumeRolePolicyDoc.Json), // Pass the JSON of the assume role policy
	})
	if err != nil {
		return nil, nil, err
	}

	// Create the IAM policy document using getPolicyDocument for permissions
	policyDocument, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
		Statements: []iam.GetPolicyDocumentStatement{
			// Statement for CloudWatch Logs
			{
				Sid:     pulumi.StringRef("LambdaLogging"),
				Actions: []string{"logs:CreateLogGroup", "logs:CreateLogStream", "logs:PutLogEvents"},
				Resources: []string{
					"arn:aws:logs:*:*:*",
				},
			},
			// Statement for SSM Parameter Access
			{
				Sid:     pulumi.StringRef("ssmParameterAccess"),
				Actions: []string{"ssm:GetParameter", "kms:Decrypt"},
				Resources: []string{
					"arn:aws:ssm:*:*:*",
				},
			},
		},
	})
	if err != nil {
		return nil, nil, err
	}
	lambdaPolicy, err := iam.NewPolicy(ctx, lambdaName+"-lambda-policy", &iam.PolicyArgs{
		Policy: pulumi.String(policyDocument.Json),
	})
	if err != nil {
		return nil, nil, err
	}
	_, err = iam.NewRolePolicyAttachment(ctx, lambdaName+"-policy-attachment", &iam.RolePolicyAttachmentArgs{
		Role:      role.Name,
		PolicyArn: lambdaPolicy.Arn,
	})

	if err != nil {
		return nil, nil, err
	}
	return role, lambdaPolicy, err
}
