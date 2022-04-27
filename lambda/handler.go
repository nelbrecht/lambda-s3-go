package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
)

func HandleRequest(ctx context.Context, name interface{}) (string, error) {
	return fmt.Sprintf("Hello %#v!", name), nil
}

func main() {
	lambda.Start(HandleRequest)
}
