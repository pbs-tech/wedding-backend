package auth

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type ParameterStore struct {
	client *ssm.Client
}

func NewParameterStoreClient() *ParameterStore {
	cfg, err := config.LoadDefaultConfig((context.TODO()))
	if err != nil {
		panic(err)
	}
	client := ssm.NewFromConfig(cfg)
	return &ParameterStore{
		client: client,
	}
}

func (ps *ParameterStore) Auth(name string, withDecryption bool) string {
	input := &ssm.GetParameterInput{
		Name:           &name,
		WithDecryption: &withDecryption,
	}
	results, err := ps.client.GetParameter(context.TODO(), input)
	if err != nil {
		panic(err)
	}
	if results.Parameter.Value == nil {
		panic(fmt.Errorf("failed to find param %s", name))
	}
	return *results.Parameter.Value
}
