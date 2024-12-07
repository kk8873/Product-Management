package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCreateProduct(t *testing.T) {
	// Mock the database and RabbitMQ setup here...

	router := setupRouter() // Initialize the router with routes

	// Prepare test data
	product := Product{
		UserID:             1,
		ProductName:        "Test Product",
		ProductDescription: "This is a test product",
		ProductImages:      []string{"http://example.com/image1.jpg", "http://example.com/image2.jpg"},
		ProductPrice:       99.99,
	}
	body, _ := json.Marshal(product)

	// Perform POST request
	req, _ := http.NewRequest("POST", "/products", bytes.NewReader(body))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusCreated, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotNil(t, response["product_id"])
}
