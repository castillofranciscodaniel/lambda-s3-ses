package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

const (
	senderEmail     = "castillofranciscodaniel@gmail.com"
	presignDuration = 24 * time.Hour
)

type awsClients struct {
	s3        *s3.Client
	s3Presign *s3.PresignClient
	ses       *sesv2.Client
}

func handler(ctx context.Context, s3Event events.S3Event) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)
	clients := &awsClients{
		s3:        s3Client,
		s3Presign: s3.NewPresignClient(s3Client),
		ses:       sesv2.NewFromConfig(cfg),
	}

	for _, record := range s3Event.Records {
		if err := processRecord(ctx, clients, record); err != nil {
			log.Printf("Error processing record: %v", err)
			// Continue with next record even if one fails
			continue
		}
	}

	return nil
}

func processRecord(ctx context.Context, clients *awsClients, record events.S3EventRecord) error {
	bucket := record.S3.Bucket.Name
	key := record.S3.Object.Key

	// 1. Get Metadata to find the client's email
	headOut, err := clients.s3.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("error getting S3 metadata: %w", err)
	}

	customerEmail := headOut.Metadata["client-email"]
	if customerEmail == "" {
		return fmt.Errorf("metadata 'client-email' not found for key: %s", key)
	}

	// 2. Generate a Presigned URL for the downloader
	presignedURL, err := getPresignedURL(ctx, clients.s3Presign, bucket, key, presignDuration)
	if err != nil {
		return fmt.Errorf("error generating presigned URL: %w", err)
	}

	// 3. Send notification email
	if err := sendNotificationEmail(ctx, clients.ses, customerEmail, key, presignedURL); err != nil {
		return fmt.Errorf("error sending email: %w", err)
	}

	log.Printf("Successfully processed and notified for: %s", key)
	return nil
}

func getPresignedURL(ctx context.Context, presignClient *s3.PresignClient, bucket, key string, duration time.Duration) (string, error) {
	params := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	result, err := presignClient.PresignGetObject(ctx, params, func(opts *s3.PresignOptions) {
		opts.Expires = duration
	})
	if err != nil {
		return "", err
	}

	return result.URL, nil
}

func sendNotificationEmail(ctx context.Context, sesClient *sesv2.Client, recipient, filename, downloadURL string) error {
	body := fmt.Sprintf(
		"Hello! Your file has been processed successfully.\n\n"+
			"You can download it at the following link (valid for 24 hours):\n%s\n\n"+
			"Thank you for using our platform.",
		downloadURL,
	)

	_, err := sesClient.SendEmail(ctx, &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(senderEmail),
		Destination: &types.Destination{
			ToAddresses: []string{recipient},
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{Data: aws.String("Contract Processed")},
				Body: &types.Body{
					Text: &types.Content{Data: aws.String(body)},
				},
			},
		},
	})

	return err
}

func main() {
	lambda.Start(handler)
}
