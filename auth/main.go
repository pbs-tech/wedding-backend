package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ParameterStore is responsible for fetching parameters from AWS SSM
type ParameterStore struct {
	client *ssm.Client
}

// RequestBody struct to handle incoming request body
type RequestBody struct {
	UserPassword string `json:"userPassword"`
}

// Claims struct defines the JWT claims (only exp and iat)
type Claims struct {
	jwt.RegisteredClaims
}

// NewParameterStoreClient initializes and returns an SSM client
func NewParameterStoreClient() *ParameterStore {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load AWS config: %v", err)
	}
	client := ssm.NewFromConfig(cfg)
	return &ParameterStore{
		client: client,
	}
}

// Get retrieves a parameter from SSM and returns it as a string
func (ps *ParameterStore) Get(name string, withDecryption bool) string {
	input := &ssm.GetParameterInput{
		Name:           &name,
		WithDecryption: &withDecryption,
	}
	results, err := ps.client.GetParameter(context.TODO(), input)
	if err != nil {
		log.Fatalf("failed to retrieve parameter %s: %v", name, err)
	}
	if results.Parameter.Value == nil {
		log.Fatalf("failed to find param %s", name)
	}
	return *results.Parameter.Value
}

// CreateJWT generates a new JWT token with expiration and issue time
func CreateJWT(secretKey string) (string, error) {
	// Set the expiration time (e.g., 1 hour)
	expirationTime := time.Now().Add(1 * time.Hour) // 1 hour expiration time

	// Create the claims with expiration and issued at time
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()), // Issued at the current time
		},
	}

	// Create the JWT token and sign it
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("could not sign the token: %v", err)
	}
	return tokenString, nil
}

// HandleRequest is the Lambda handler that processes the request and generates a JWT token
func HandleRequest(ctx context.Context, apiGatewayRequest events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// Get parameters from SSM
	authPasswordParam := os.Getenv("AUTH_PASSWORD_PARAM")
	jwtSigningParam := os.Getenv("JWT_SIGNING_SECRET_PARAM")
	paramStore := NewParameterStoreClient()
	authPassword := paramStore.Get(authPasswordParam, true)
	jwtSecret := paramStore.Get(jwtSigningParam, true)

	// Parse incoming request body
	var body RequestBody
	err := json.Unmarshal([]byte(apiGatewayRequest.Body), &body)
	if err != nil {
		log.Printf("Failed to parse request body: %v", &apiGatewayRequest.Body)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusBadRequest,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"message": "Invalid request body"}`,
		}, nil
	}

	// Compare the provided password with the stored hash in SSM
	err = bcrypt.CompareHashAndPassword([]byte(authPassword), []byte(body.UserPassword))
	if err != nil {
		log.Printf("Password comparison failed: %v", err)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusUnauthorized,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"message": "Unauthorized"}`,
		}, nil
	}

	// Generate a new JWT token if password is correct
	token, err := CreateJWT(jwtSecret)
	if err != nil {
		log.Printf("Failed to generate JWT token: %v", err)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"message": "Failed to generate JWT token"}`,
		}, nil
	}

	// Return the JWT token in the response body
	return events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: fmt.Sprintf(`{"token": "%s"}`, token),
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
