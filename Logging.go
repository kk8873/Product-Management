func createProduct(c *gin.Context) {
	var product Product
	if err := c.ShouldBindJSON(&product); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"body":  c.Request.Body,
		}).Error("Failed to parse request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Insert product into the database
	query := `INSERT INTO products (user_id, product_name, product_description, product_images, product_price) 
	          VALUES ($1, $2, $3, $4, $5) RETURNING id`
	var productID int
	err := db.QueryRow(query, product.UserID, product.ProductName, product.ProductDescription, pq.Array(product.ProductImages), product.ProductPrice).Scan(&productID)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"query": query,
		}).Error("Failed to insert product into database")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save product"})
		return
	}
	product.ID = productID

	logrus.WithFields(logrus.Fields{
		"product_id": productID,
	}).Info("Product successfully created")

	// Send image URLs to RabbitMQ for processing
	imageData, _ := json.Marshal(product.ProductImages)
	err = mqChannel.Publish("", "image_queue", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        imageData,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"images": product.ProductImages,
		}).Error("Failed to publish images to RabbitMQ")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue images"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"product_id": product.ID})
}
