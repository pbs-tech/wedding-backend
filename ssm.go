package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ssm"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createSSMParameter(ctx *pulumi.Context, authPassword string) (*ssm.Parameter, error) {
	parameter, err := ssm.NewParameter(ctx, "auth-password", &ssm.ParameterArgs{
		Name:        pulumi.String("authPassword"),
		Description: pulumi.String("Password for users to use to access the site"),
		Type:        pulumi.String(ssm.ParameterTypeSecureString),
		Value:       pulumi.String(authPassword),
	})
	if err != nil {
		return nil, err
	}
	return parameter, err
}
