package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ssm"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func createSSMParameter(ctx *pulumi.Context) (*ssm.Parameter, error) {
	cfg := config.New(ctx, "")
	parameter, err := ssm.NewParameter(ctx, "auth-password", &ssm.ParameterArgs{
		Name:        pulumi.String("authPassword"),
		Description: pulumi.String("Password for users to use to access the site"),
		Type:        pulumi.String(ssm.ParameterTypeSecureString),
		Value:       cfg.RequireSecret("authPassword"),
	})
	if err != nil {
		return nil, err
	}
	return parameter, err
}
