/*
go get github.com/gin-gonic/gin
go get github.com/jmoiron/sqlx
go get github.com/redis/go-redis/v9
go get github.com/streadway/amqp
go get github.com/sirupsen/logrus

*/

router.Use(errorHandlingMiddleware)

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

// Product represents the product structure
type Product struct {
	ID                  int      `json:"id" db:"id"`
	UserID              int      `json:"user_id" db:"user_id"`
	ProductName         string   `json:"product_name" db:"product_name"`
	ProductDescription  string   `json:"product_description" db:"product_description"`
	ProductImages       []string `json:"product_images" db:"product_images"`
	ProductPrice        float64  `json:"product_price" db:"product_price"`
	CompressedImagesURL []string `json:"compressed_product_images" db:"compressed_product_images"`
}

// Database connection
var db *sqlx.DB

// Redis client
var redisClient *redis.Client

// RabbitMQ connection
var mqChannel *amqp.Channel

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Initialize database
	var err error
	db, err = sqlx.Connect("postgres", "host=localhost port=5432 user=youruser password=yourpassword dbname=yourdb sslmode=disable")
	if err != nil {
		logger.Fatal("Failed to connect to the database:", err)
	}

	// Initialize Redis
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	_, err = redisClient.Ping(context.Background()).Result()
	if err != nil {
		logger.Fatal("Failed to connect to Redis:", err)
	}

	// Initialize RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		logger.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer conn.Close()

	mqChannel, err = conn.Channel()
	if err != nil {
		logger.Fatal("Failed to open a channel in RabbitMQ:", err)
	}
	defer mqChannel.Close()

	// Declare a queue
	_, err = mqChannel.QueueDeclare("image_queue", false, false, false, false, nil)
	if err != nil {
		logger.Fatal("Failed to declare a queue:", err)
	}

	// Initialize router
	router := gin.Default()

	// Define routes
	router.POST("/products", createProduct)
	router.GET("/products/:id", getProductByID)
	router.GET("/products", getProducts)

	// Start server
	logger.Info("Starting server on port 8080...")
	router.Run(":8080")
}

func createProduct(c *gin.Context) {
	var product Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Insert product into the database
	query := `INSERT INTO products (user_id, product_name, product_description, product_images, product_price) 
	          VALUES ($1, $2, $3, $4, $5) RETURNING id`
	var productID int
	err := db.QueryRow(query, product.UserID, product.ProductName, product.ProductDescription, pq.Array(product.ProductImages), product.ProductPrice).Scan(&productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save product"})
		return
	}
	product.ID = productID

	// Send image URLs to RabbitMQ for processing
	imageData, _ := json.Marshal(product.ProductImages)
	err = mqChannel.Publish("", "image_queue", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        imageData,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue images"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"product_id": product.ID})
}

func getProductByID(c *gin.Context) {
	id := c.Param("id")

	// Check Redis cache
	cacheKey := "product:" + id
	cachedProduct, err := redisClient.Get(context.Background(), cacheKey).Result()
	if err == nil {
		var product Product
		json.Unmarshal([]byte(cachedProduct), &product)
		c.JSON(http.StatusOK, product)
		return
	}

	// Fetch from database
	var product Product
	query := `SELECT * FROM products WHERE id = $1`
	err = db.Get(&product, query, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Cache the product in Redis
	productJSON, _ := json.Marshal(product)
	redisClient.Set(context.Background(), cacheKey, productJSON, 0)

	c.JSON(http.StatusOK, product)
}

func getProducts(c *gin.Context) {
	userID := c.Query("user_id")
	minPrice := c.Query("min_price")
	maxPrice := c.Query("max_price")

	query := `SELECT * FROM products WHERE user_id = $1`
	args := []interface{}{userID}

	if minPrice != "" {
		query += " AND product_price >= $2"
		args = append(args, minPrice)
	}
	if maxPrice != "" {
		query += " AND product_price <= $3"
		args = append(args, maxPrice)
	}

	var products []Product
	err := db.Select(&products, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve products"})
		return
	}

	c.JSON(http.StatusOK, products)
}
