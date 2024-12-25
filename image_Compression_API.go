/*
go get github.com/streadway/amqp
go get github.com/nfnt/resize
go get github.com/sirupsen/logrus
go get github.com/disintegration/imaging
go get github.com/lib/pq
*/

package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	_ "github.com/lib/pq"
)

// Configuration
const (
	rabbitMQURL    = "amqp://guest:guest@localhost:5672/"
	imageQueueName = "image_queue"
	dbHost         = "localhost"
	dbPort         = 5432
	dbUser         = "your_user"
	dbPassword     = "your_password"
	dbName         = "your_db"
)

// PostgreSQL client
var db *sql.DB

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Initialize PostgreSQL
	var err error
	connStr := "host=" + dbHost + " port=" + string(dbPort) + " user=" + dbUser + " password=" + dbPassword + " dbname=" + dbName + " sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		logger.Fatal("Failed to connect to PostgreSQL:", err)
	}
	defer db.Close()

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

	// Convert image to bytes
	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, compressedImage, &jpeg.Options{Quality: 80}); err != nil {
		return err
	}
	imageBytes := buf.Bytes()

	// Save image to PostgreSQL
	imageName := getImageName(imageURL)
	return saveImageToDB(imageURL, imageName, imageBytes, "processed")
}

func saveImageToDB(imageURL, imageName string, imageBytes []byte, status string) error {
	query := `
		INSERT INTO image_data (image_url, image_name, image_content, status, processed_at)
		VALUES ($1, $2, $3, $4, NOW())
	`
	_, err := db.Exec(query, imageURL, imageName, imageBytes, status)
	return err
}

func getImageName(url string) string {
	parts := strings.Split(url, "/")
	return parts[len(parts)-1]
}
