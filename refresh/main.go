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
)

type ParameterStore struct {
	client *ssm.Client
}

type Claims struct {
	jwt.RegisteredClaims
}

type RequestBody struct {
	JWTToken string `json:"jwtToken"`
}

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

// GenerateRefreshToken generates a JWT refresh token with a 30-day expiration
func GenerateRefreshToken(jwtSecret string) (string, error) {
	expirationTime := time.Now().Add(30 * 24 * time.Hour) // 30 days

	// Create the JWT claims (no username in this case)
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", fmt.Errorf("could not sign the token: %v", err)
	}
	return signedToken, nil
}

// VerifyToken verifies the JWT token and returns the claims or an error
func VerifyToken(tokenString, secretKey string) (*Claims, error) {
	// Parse and verify the token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Check that the signing method matches
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("could not parse token: %v", err)
	}

	// Check if the token is valid and return the claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, fmt.Errorf("invalid token")
	}
}

// handleRequest handles incoming API Gateway requests
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

	// If JWTToken is provided, verify it
	if body.JWTToken != "" {
		// Verify the provided token
		_, err := VerifyToken(body.JWTToken, jwtSecret)
		if err != nil {
			// If verification fails, return Unauthorized response
			return events.APIGatewayV2HTTPResponse{
				StatusCode: http.StatusUnauthorized,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: "Unauthorized",
			}, nil
		}
	}

	// Generate a new refresh token
	refreshToken, err := GenerateRefreshToken(jwtSecret)
	if err != nil {
		log.Printf("Failed to generate refresh token: %v", err)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: "Failed to generate refresh token",
		}, nil
	}

	// Return the refresh token
	return events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: fmt.Sprintf(`{"jwtToken": "%s"}`, refreshToken),
	}, nil
}

func main() {
	lambda.Start(handleRequest)
}
