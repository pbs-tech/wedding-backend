package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lambda"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ssm"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createAuthLambda(ctx *pulumi.Context, authPasswordParam *ssm.Parameter) (*lambda.Function, error) {
	lambdaName := "auth"
	role, lambdaPolicy, err := createLambdaIamRolePolicy(ctx, lambdaName, authPasswordParam.Arn)
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

func createLambdaIamRolePolicy(ctx *pulumi.Context, lambdaName string, authPasswordParamArn pulumi.StringOutput) (*iam.Role, *iam.Policy, error) {
	assumeRolePolicyStatement := pulumi.String(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": "sts:AssumeRole",
				"Principal": {
					"Service": "lambda.amazonaws.com"
				}
			}
		]
	}`)

	role, err := iam.NewRole(ctx, lambdaName+"-exec-role", &iam.RoleArgs{
		AssumeRolePolicy: assumeRolePolicyStatement,
	})
	if err != nil {
		return nil, nil, err
	}

	lambdaPolicyStatement := pulumi.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Sid": "LambdaLogging",
				"Effect": "Allow",
				"Action": [
					"logs:CreateLogGroup",
					"logs:CreateLogStream",
					"logs:PutLogEvents"
				],
				"Resource": [
					"arn:aws:logs:*:*:*"
				]
			},
			{
				"Sid": "ssmParameterAccess",
				"Effect": "Allow",
				"Action": [
					"ssm:GetParameter",
					"kms:Decrypt"
				],
				"Resource": [
					"%s"
				]
			}
		]
	}`, authPasswordParamArn)
	lambdaPolicy, err := iam.NewPolicy(ctx, lambdaName+"-lambda-policy", &iam.PolicyArgs{
		Policy: lambdaPolicyStatement,
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
