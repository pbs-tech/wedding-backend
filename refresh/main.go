package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type ParameterStore struct {
	client *ssm.Client
}

type RequestBody struct {
	JWTToken string `json:"jwtToken"`
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

func (ps *ParameterStore) Get(name string, withDecryption bool) string {
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

func VerifyToken(jwtToken string, jwtSecret string) error {
	token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return fmt.Errorf(("invalid token"))
	}
	return nil
}

func handleRequest(ctx context.Context, apiGatewayRequest events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	jwtSigningParam := os.Getenv("JWT_SIGNING_SECRET_PARAM")
	paramStore := NewParameterStoreClient()
	jwtSecret := paramStore.Get(jwtSigningParam, true)

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
	result := VerifyToken(body.JWTToken, jwtSecret)
	if result == nil {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusAccepted,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: "Authorised",
		}, nil
	} else {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusUnauthorized,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: "Unauthorised",
		}, nil
	}
}

func main() {
	lambda.Start(handleRequest)
}
