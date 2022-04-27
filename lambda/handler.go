package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func HandleRequest(ctx context.Context, event events.S3Event) (string, error) {
	log.Printf("Typed EVENT: %#v", event)

	eventJson, _ := json.Marshal(event)
	log.Printf("Unformatted EVENT: %s", eventJson)

	eventJson, _ = json.MarshalIndent(event, "", "  ")
	log.Printf("EVENT: %s", eventJson)

	// environment variables
	log.Printf("REGION: %s", os.Getenv("AWS_REGION"))
	log.Println("ALL ENV VARS:")
	for _, element := range os.Environ() {
		log.Println(element)
	}

	return fmt.Sprintf("Hello %#v!", event), nil
}

func main() {
	lambda.Start(HandleRequest)
}
