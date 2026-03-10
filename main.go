package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

func handler(ctx context.Context, s3Event events.S3Event) error {
	// Cargar configuración de AWS v2
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("error cargando config: %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)
	sesClient := sesv2.NewFromConfig(cfg)

	for _, record := range s3Event.Records {
		bucket := record.S3.Bucket.Name
		key := record.S3.Object.Key

		// 1. Obtener Metadata
		headOut, err := s3Client.HeadObject(ctx, &s3.HeadObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			log.Printf("Error S3 HeadObject: %v", err)
			continue
		}

		// En v2, la metadata es map[string]string (ya no son punteros)
		customerEmail := headOut.Metadata["client-email"]
		if customerEmail == "" {
			log.Printf("Metadata 'client-email' no encontrada en %s", key)
			continue
		}

		// 2. Enviar Email con SES v2
		_, err = sesClient.SendEmail(ctx, &sesv2.SendEmailInput{
			FromEmailAddress: aws.String("castillofranciscodaniel@gmail.com"),
			Destination: &types.Destination{
				ToAddresses: []string{customerEmail},
			},
			Content: &types.EmailContent{
				Simple: &types.Message{
					Subject: &types.Content{Data: aws.String("Contrato Procesado")},
					Body: &types.Body{
						Text: &types.Content{Data: aws.String("Tu archivo " + key + " ha sido procesado.")},
					},
				},
			},
		})

		if err != nil {
			log.Printf("Error SES: %v", err)
		} else {
			log.Printf("Éxito enviando mail a: %s", customerEmail)
		}
	}
	return nil
}

func main() {
	lambda.Start(handler)
}
