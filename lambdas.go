package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createAuthLambda(ctx *pulumi.Context, dynamoDbTableName string) (*lambda.Function, error) {
	role, logPolicy, err := createLambdaIamRolePolicy(ctx)
	authLambda, err := lambda.NewFunction(ctx, "auth-lambda", &lambda.FunctionArgs{
		Runtime: lambda.RuntimeCustomAL2023,
		Code: pulumi.NewAssetArchive(map[string]interface{}{
			".": pulumi.NewFileArchive("./bin/auth.zip"),
		}),
		Handler:       pulumi.String("auth"),
		Role:          role.Arn,
		Architectures: pulumi.StringArray{pulumi.String("arm64")},
		Environment: &lambda.FunctionEnvironmentArgs{
			Variables: pulumi.StringMap{
				"DYNAMODB_TABLE_NAME": pulumi.String(dynamoDbTableName),
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{logPolicy}))
	if err != nil {
		return nil, err
	}
	return authLambda, nil
}

func createRSVPLambda(ctx *pulumi.Context, dynamoDbTableName string) (*lambda.Function, error) {
	role, logPolicy, err := createLambdaIamRolePolicy(ctx)
	rsvpLambda, err := lambda.NewFunction(ctx, "rsvp-lambda", &lambda.FunctionArgs{
		Runtime: lambda.RuntimeCustomAL2023,
		Code: pulumi.NewAssetArchive(map[string]interface{}{
			".": pulumi.NewFileArchive("./bin/rsvp.zip"),
		}),
		Handler:       pulumi.String("rsvp"),
		Role:          role.Arn,
		Architectures: pulumi.StringArray{pulumi.String("arm64")},
		Environment: &lambda.FunctionEnvironmentArgs{
			Variables: pulumi.StringMap{
				"DYNAMODB_TABLE_NAME": pulumi.String(dynamoDbTableName),
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{logPolicy}))
	if err != nil {
		return nil, err
	}
	return rsvpLambda, nil
}

func createLambdaIamRolePolicy(ctx *pulumi.Context) (*iam.Role, *iam.RolePolicy, error) {
	role, err := iam.NewRole(ctx, "auth-exec-role", &iam.RoleArgs{
		AssumeRolePolicy: pulumi.String(`{
			"Version": "2012-10-17",
			"Statement": [{
					"Sid": "",
					"Effect": "Allow",
					"Principal": {
						"Service": "lambda.amazonaws.com"
					},
					"Action": "sts:AssumeRole"
			}]
			}`),
	})
	if err != nil {
		return nil, nil, err
	}
	logPolicy, err := iam.NewRolePolicy(ctx, "lambda-log-policy", &iam.RolePolicyArgs{
		Role: role.Name,
		Policy: pulumi.String(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Sid": "",
				"Effect": "Allow",
				"Action": [
					"logs:CreateLogGroup",
					"logs:CreateLogStream",
					"logs:PutLogEvents"
				],
				"Resource": "arn:aws:logs:*:*:*"
			}]
		}`),
	})
	if err != nil {
		return nil, nil, err
	}
	return role, logPolicy, err
}
