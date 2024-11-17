package main

import (
	"encoding/json"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lambda"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ssm"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createAuthLambda(ctx *pulumi.Context, authPasswordParam *ssm.Parameter) (*lambda.Function, error) {
	lambdaName := "auth-lambda"
	authPasswordParamArn := pulumi.StringOutput(authPasswordParam.Arn)
	role, lambdaPolicy, err := createLambdaIamRolePolicy(ctx, lambdaName, authPasswordParamArn)
	if err != nil {
		return nil, err
	}
	authLambda, err := lambda.NewFunction(ctx, lambdaName, &lambda.FunctionArgs{
		Runtime: lambda.RuntimeCustomAL2023,
		Code: pulumi.NewAssetArchive(map[string]interface{}{
			".": pulumi.NewFileArchive("./bin/auth.zip"),
		}),
		Handler:       pulumi.String("auth"),
		Role:          role.Arn,
		Architectures: pulumi.StringArray{pulumi.String("arm64")},
		Environment: &lambda.FunctionEnvironmentArgs{
			Variables: pulumi.StringMap{
				"AUTH_PASSWORD_PARAM": authPasswordParam.Name,
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{lambdaPolicy}))
	if err != nil {
		return nil, err
	}
	return authLambda, nil
}

func createLambdaIamRolePolicy(ctx *pulumi.Context, lambdaName string, authPasswordParamArn pulumi.StringOutput) (*iam.Role, *iam.RolePolicy, error) {
	assumeRolePolicyJSON, err := json.Marshal(map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": map[string]interface{}{
			"Sid":    "",
			"Effect": "Allow",
			"Action": "sts:AssumeRole",
			"Principal": map[string]interface{}{
				"Service": "lambda.amazonaws.com",
			},
		},
	})
	if err != nil {
		return nil, nil, err
	}
	assumeRolePolicyStr := string(assumeRolePolicyJSON)
	role, err := iam.NewRole(ctx, lambdaName+"-exec-role", &iam.RoleArgs{
		AssumeRolePolicy: pulumi.String(assumeRolePolicyStr),
	})
	if err != nil {
		return nil, nil, err
	}
	lambdaPolicyJSON, err := json.Marshal(map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Sid":    "LambdaLogging",
				"Effect": "Allow",
				"Action": []string{
					"logs:CreateLogGroup",
					"logs:CreateLogStream",
					"logs:PutLogEvents",
				},
				"Resource": "arn:aws:logs:*:*:*",
			},
			{
				"Sid":    "LambdaLogging",
				"Effect": "Allow",
				"Action": []string{
					"ssm:GetParameters",
					"kms:Decrypt",
				},
				"Resource": authPasswordParamArn,
			},
		},
	})
	if err != nil {
		return nil, nil, err
	}
	lambdaPolicyStr := string(lambdaPolicyJSON)
	lambdaPolicy, err := iam.NewRolePolicy(ctx, lambdaName+"-lambda-policy", &iam.RolePolicyArgs{
		Role:   role.Name,
		Policy: pulumi.String(lambdaPolicyStr),
	})
	if err != nil {
		return nil, nil, err
	}
	return role, lambdaPolicy, err
}
