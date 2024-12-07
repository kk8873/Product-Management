func getProductByID(c *gin.Context) {
	id := c.Param("id")

	// Check Redis cache
	cacheKey := "product:" + id
	cachedProduct, err := redisClient.Get(context.Background(), cacheKey).Result()
	if err == nil {
		// If cache hit, return the cached data
		var product Product
		json.Unmarshal([]byte(cachedProduct), &product)
		logrus.Info("Cache hit for product ID:", id)
		c.JSON(http.StatusOK, product)
		return
	}

	// Fetch from database if cache miss
	var product Product
	query := `SELECT * FROM products WHERE id = $1`
	err = db.Get(&product, query, id)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product"})
		}
		return
	}

	// Cache the product in Redis
	productJSON, _ := json.Marshal(product)
	redisClient.Set(context.Background(), cacheKey, productJSON, 0)

	logrus.Info("Product cached in Redis with ID:", id)
	c.JSON(http.StatusOK, product)
}
