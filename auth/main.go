package auth

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

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
