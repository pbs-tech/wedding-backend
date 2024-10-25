package main

import (
	"github.com/aws/aws-lambda-go/lambda"
)

func handleRequest() (string, error) {
	return "Hello there my fried friend", nil
}

func main() {
	lambda.Start(handleRequest)
}
