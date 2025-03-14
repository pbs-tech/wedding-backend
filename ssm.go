package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ssm"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createSSMParameter(ctx *pulumi.Context, key string, description string, value pulumi.StringOutput) (*ssm.Parameter, error) {
	parameter, err := ssm.NewParameter(ctx, key, &ssm.ParameterArgs{
		Name:        pulumi.String(key),
		Description: pulumi.String(description),
		Type:        pulumi.String(ssm.ParameterTypeSecureString),
		Value:       value,
	})
	if err != nil {
		return nil, err
	}
	return parameter, err
}
