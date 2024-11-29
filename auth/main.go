package main

import (
	"bcrypt"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"golang.org/x/crypto/bcrypt"
)

type ParameterStore struct {
	client *ssm.Client
}

type RequestBody struct {
	UserPassword string `json:"userPassword"`
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
	var body RequestBody
	err := json.Unmarshal([]byte(apiGatewayRequest.Body), &body)
	if err != nil {
		log.Printf("Failed to parse request body: %v", &apiGatewayRequest.Body)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusBadRequest,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: "Invalid request body",
		}, nil
	}
	if bcrypt.CompareHashAndPassword(body.UserPassword, authPassword) {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusAccepted,
			Body:       "Authorised",
		}, nil
	} else {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusUnauthorized,
			Body:       "Unauthorised",
		}, nil
	}
}

func main() {
	lambda.Start(handleRequest)
}
