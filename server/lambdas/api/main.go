package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"goattic-api/internal/api"
	"goattic-api/internal/storage"

	"github.com/aws/aws-lambda-go/lambda"
	chiadapter "github.com/awslabs/aws-lambda-go-api-proxy/chi"
)

func main() {
	ctx := context.Background()

	db, err := storage.NewDynamoDBClient(ctx)
	if err != nil {
		log.Fatal("Failed to create DynamoDB client:", err)
	}

	s3, err := storage.NewS3Client(ctx)
	if err != nil {
		log.Fatal("Failed to create S3 client:", err)
	}

	mediaDomain := os.Getenv("MEDIA_DOMAIN")
	apiHandler := api.NewAPI(db, s3, mediaDomain)
	router := apiHandler.Router()

	// needed for localstack
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		adapter := chiadapter.New(router)
		lambda.Start(adapter.ProxyWithContext)
	} else {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		log.Printf("Starting server on :%s", port)
		if err := http.ListenAndServe(":"+port, router); err != nil {
			log.Fatal(err)
		}
	}
}
