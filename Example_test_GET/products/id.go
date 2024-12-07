func TestGetProductByID(t *testing.T) {
	router := setupRouter() // Initialize the router

	// Mock a product in the database
	productID := 1

	// Perform GET request
	req, _ := http.NewRequest("GET", "/products/"+strconv.Itoa(productID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	var product Product
	json.Unmarshal(w.Body.Bytes(), &product)
	assert.Equal(t, productID, product.ID)
}
