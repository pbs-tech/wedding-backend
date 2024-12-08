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
func (ps *ParameterStore) Get(name string, withDecryption bool) (string, error) {
	input := &ssm.GetParameterInput{
		Name:           &name,
		WithDecryption: &withDecryption,
	}
	results, err := ps.client.GetParameter(context.TODO(), input)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve parameter %s: %v", name, err)
	}
	if results.Parameter.Value == nil {
		return "", fmt.Errorf("failed to find param %s", name)
	}
	return *results.Parameter.Value, nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
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
	responseHeaders := map[string]string{
		"Content-Type":                 "application/json",
		"Access-Control-Allow-Origin":  "https://peebles.lol",
		"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
		"Access-Control-Allow-Headers": "Content-Type, Authorization, Origin",
	}
	// Check if environment variables are set
	if authPasswordParam == "" || jwtSigningParam == "" {
		log.Println("Environment variables for SSM parameters are missing.")
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    responseHeaders,
			Body:       `{"message": "Required environment variables are not set"}`,
		}, nil
	}

	// Initialize the Parameter Store client
	paramStore := NewParameterStoreClient()

	// Fetch the parameters from SSM
	authPassword, err := paramStore.Get(authPasswordParam, true)
	if err != nil {
		log.Printf("Error fetching auth password: %v", err)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    responseHeaders,
			Body:       `{"message": "Error retrieving auth password"}`,
		}, nil
	}
	authPasswordHash, err := HashPassword(authPassword)
	if err != nil {
		log.Printf("Error hashing auth password: %v", err)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    responseHeaders,
			Body:       `{"message": "Error hashing auth password"}`,
		}, nil
	}

	jwtSecret, err := paramStore.Get(jwtSigningParam, true)
	if err != nil {
		log.Printf("Error fetching JWT secret: %v", err)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    responseHeaders,
			Body:       `{"message": "Error retrieving JWT secret"}`,
		}, nil
	}

	// Parse incoming request body
	var body RequestBody
	err = json.Unmarshal([]byte(apiGatewayRequest.Body), &body)
	if err != nil {
		log.Printf("Failed to parse request body: %v", apiGatewayRequest.Body)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusBadRequest,
			Headers:    responseHeaders,
			Body:       `{"message": "Invalid request body"}`,
		}, nil
	}

	// Compare the provided password with the stored hash in SSM
	err = bcrypt.CompareHashAndPassword([]byte(authPasswordHash), []byte(body.UserPassword))
	if err == nil {
		// Generate a new JWT token if password is correct
		token, err := CreateJWT(jwtSecret)
		if err != nil {
			log.Printf("Error creating JWT token: %v", err)
			return events.APIGatewayV2HTTPResponse{
				StatusCode: http.StatusInternalServerError,
				Headers:    responseHeaders,
				Body:       `{"message": "Error creating JWT token"}`,
			}, nil
		}

		// Return the JWT token in the response body
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusOK,
			Headers:    responseHeaders,
			Body:       fmt.Sprintf(`{"jwtToken": "%s"}`, token),
		}, nil
	} else {
		// Handle the case where the password comparison failed
		log.Printf("Password comparison failed: %v", err)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusUnauthorized,
			Headers:    responseHeaders,
			Body:       `{"message": "Unauthorized"}`,
		}, nil
	}
}

func main() {
	lambda.Start(HandleRequest)
}
