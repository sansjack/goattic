package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"goattic-file-validator/validate"
	"goattic-shared/discord"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatal("Failed to load AWS config:", err)
	}
	s3Client := s3.NewFromConfig(cfg)

	webhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
	dispatcher := discord.NewDispatcher(webhookURL)

	lambda.Start(func(ctx context.Context, event events.S3Event) error {
		for _, record := range event.Records {
			bucket := record.S3.Bucket.Name
			key := record.S3.Object.Key
			size := record.S3.Object.Size

			header, err := validate.ReadHeader(ctx, s3Client, bucket, key)
			if err != nil {
				log.Printf("ERROR: failed to read header for s3://%s/%s: %v", bucket, key, err)
				continue
			}

			fileType, ok := validate.DetectFileType(header)
			if !ok {
				log.Printf("WARNING: unrecognized file type for s3://%s/%s (header: %x) — deleting", bucket, key, header)

				if err := validate.DeleteObject(ctx, s3Client, bucket, key); err != nil {
					log.Printf("ERROR: failed to delete s3://%s/%s: %v", bucket, key, err)
				}

				dispatcher.Send(discord.WebhookPayload{
					Embeds: []discord.Embed{{
						Title:       "Suspicious file deleted",
						Description: fmt.Sprintf("Unrecognized file type — deleted\n`s3://%s/%s` (%d bytes)\nHeader: `%x`", bucket, key, size, header),
						Color:       0xFF0000,
					}},
				})
				continue
			}

			log.Printf("File validated: s3://%s/%s type=%s size=%d", bucket, key, fileType, size)
			dispatcher.Send(discord.WebhookPayload{
				Embeds: []discord.Embed{{
					Title:       "File validated",
					Description: fmt.Sprintf("`s3://%s/%s`\nType: **%s** | Size: %d bytes", bucket, key, fileType, size),
					Color:       0x00FF00,
				}},
			})
		}

		dispatcher.Flush()
		return nil
	})
}
