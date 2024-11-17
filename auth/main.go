package auth

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
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

func handleRequest(ctx context.Context, apiGatewayRequest events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	authPasswordParam := os.Getenv("AUTH_PASSWORD_PARAM")
	paramStore := NewParameterStoreClient()
	authPassword := paramStore.Auth(authPasswordParam, true)
	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Body:       authPassword,
	}, nil
}

func main() {
	lambda.Start(handleRequest)
}
