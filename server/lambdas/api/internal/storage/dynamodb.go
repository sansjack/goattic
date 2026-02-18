package storage

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoDBClient struct {
	client    *dynamodb.Client
	tableName string
}

// keep in sync with the CDK construct definition and seed data
type APIKey struct {
	APIKey    string `dynamodbav:"apiKey"`
	Owner     string `dynamodbav:"owner"`
	CreatedAt string `dynamodbav:"createdAt"`
	RateLimit int    `dynamodbav:"rateLimit"`
	Enabled   bool   `dynamodbav:"enabled"`
}

func NewDynamoDBClient(ctx context.Context) (*DynamoDBClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		if endpoint := os.Getenv("AWS_ENDPOINT_URL"); endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
	})

	tableName := os.Getenv("TABLE_NAME")
	if tableName == "" {
		tableName = "goattic-apikeys"
	}

	return &DynamoDBClient{
		client:    client,
		tableName: tableName,
	}, nil
}

func (d *DynamoDBClient) ValidateAPIKey(ctx context.Context, apiKey string) (*APIKey, error) {
	result, err := d.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]types.AttributeValue{
			"apiKey": &types.AttributeValueMemberS{Value: apiKey},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("API key not found")
	}

	var key APIKey
	if err := attributevalue.UnmarshalMap(result.Item, &key); err != nil {
		return nil, fmt.Errorf("failed to unmarshal item: %w", err)
	}

	return &key, nil
}

func (d *DynamoDBClient) ListAPIKeysByOwner(ctx context.Context, owner string) ([]APIKey, error) {
	fmt.Printf("DEBUG: Listing API keys for owner: %s\n", owner)
	result, err := d.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(d.tableName),
		IndexName:              aws.String("OwnerIndex"),
		KeyConditionExpression: aws.String("#owner = :owner"),
		ExpressionAttributeNames: map[string]string{
			"#owner": "owner",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":owner": &types.AttributeValueMemberS{Value: owner},
		},
	})
	if err != nil {
		fmt.Printf("DEBUG: Query error: %v\n", err)
		return nil, fmt.Errorf("failed to query items: %w", err)
	}

	fmt.Printf("DEBUG: Query returned %d items\n", len(result.Items))

	var keys []APIKey
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &keys); err != nil {
		fmt.Printf("DEBUG: Unmarshal error: %v\n", err)
		return nil, fmt.Errorf("failed to unmarshal items: %w", err)
	}

	fmt.Printf("DEBUG: Successfully unmarshaled %d keys\n", len(keys))
	return keys, nil
}
