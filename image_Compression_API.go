
/*
go get github.com/streadway/amqp
go get github.com/aws/aws-sdk-go
go get github.com/nfnt/resize
go get github.com/sirupsen/logrus
go get github.com/disintegration/imaging
*/

package main

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/disintegration/imaging"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

// Configuration
const (
	rabbitMQURL    = "amqp://guest:guest@localhost:5672/"
	imageQueueName = "image_queue"
	s3Bucket       = "your-s3-bucket"
	region         = "us-east-1"
)

// AWS S3 client
var s3Client *s3.S3

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Initialize RabbitMQ
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		logger.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		logger.Fatal("Failed to open a channel in RabbitMQ:", err)
	}
	defer ch.Close()

	msgs, err := ch.Consume(
		imageQueueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logger.Fatal("Failed to register a consumer:", err)
	}

	// Initialize AWS S3
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	s3Client = s3.New(sess)

	// Process messages from the queue
	forever := make(chan bool)

	go func() {
		for msg := range msgs {
			logger.Info("Received message:", string(msg.Body))
			var imageURLs []string
			err := json.Unmarshal(msg.Body, &imageURLs)
			if err != nil {
				logger.Error("Failed to parse message:", err)
				continue
			}

			for _, url := range imageURLs {
				logger.Info("Processing image:", url)
				if err := processImage(url); err != nil {
					logger.Error("Failed to process image:", err)
				} else {
					logger.Info("Successfully processed image:", url)
				}
			}
		}
	}()

	logger.Info("Waiting for messages...")
	<-forever
}

func processImage(imageURL string) error {
	// Download image
	resp, err := http.Get(imageURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return err
	}

	// Compress image
	compressedImage := imaging.Resize(img, 800, 0, imaging.Lanczos)

	// Save compressed image to S3
	imageName := getImageName(imageURL)
	return uploadToS3(compressedImage, imageName)
}

func uploadToS3(img image.Image, imageName string) error {
	// Convert image to JPEG format
	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 80}); err != nil {
		return err
	}

	// Upload to S3
	_, err := s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(s3Bucket),
		Key:         aws.String(imageName),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String("image/jpeg"),
	})
	return err
}

func getImageName(url string) string {
	parts := strings.Split(url, "/")
	return parts[len(parts)-1]
}
