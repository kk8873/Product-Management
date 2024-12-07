// # Product Management System

// A high-performance backend system built in Golang for managing products. This application incorporates modern backend development practices such as asynchronous image processing, caching, and scalable architecture.

// ## Features
// - **RESTful API**: Endpoints to create, retrieve, and filter products.
// - **Asynchronous Image Processing**: Uses RabbitMQ to compress and store images (AWS S3 setup is pending).
// - **Caching with Redis**: Reduces database load for frequent queries.
// - **Structured Logging**: Tracks operations with Logrus for debugging and monitoring.
// - **Error Handling**: Handles errors with retry mechanisms and dead-letter queues.

// ## Technologies Used
// - **Programming Language**: Golang
// - **Database**: PostgreSQL
// - **Message Queue**: RabbitMQ
// - **Caching**: Redis
// - **Image Storage**: (Placeholder for AWS S3)
// - **Logging Library**: Logrus

// ## Setup Instructions

// ### Prerequisites
// - **Go**: Version 1.20 or higher
// - **PostgreSQL**: Running instance for data storage
// - **Redis**: For caching
// - **RabbitMQ**: Message broker
// - **AWS Account (Optional)**: For S3 bucket setup (not mandatory for this submission)

// ### Installation

// 1. Clone the repository:
//    ```bash
//    git clone https://github.com/karansingh/product-management-system.git
//    cd product-management-system
//    ```

// 2. Install dependencies:
//    ```bash
//    go mod tidy
//    ```

// 3. Create a .env file with the following:
//    ```env
//    DB_HOST=localhost
//    DB_PORT=5432
//    DB_USER=postgres
//    DB_PASSWORD=yourpassword
//    DB_NAME=productdb
//    REDIS_HOST=localhost
//    REDIS_PORT=6379
//    RABBITMQ_URL=amqp://guest:guest@localhost:5672/
//    S3_BUCKET=placeholder-bucket-name
//    AWS_REGION=us-east-1
//    AWS_ACCESS_KEY=placeholder-access-key
//    AWS_SECRET_KEY=placeholder-secret-key
//    ```

// 4. Run the application:
//    ```bash
//    go run main.go
//    ```

// ## API Endpoints
// 1. Create Product
//    **POST /products**

/*
   Request Body:
   {
     "user_id": 1,
     "product_name": "Sample Product",
     "product_description": "This is a sample product description",
     "product_images": ["http://example.com/image1.jpg"],
     "product_price": 99.99
   }
*/

/*
   ## Response
   {
     "product_id": 1
   }
*/

// 2. Get Product by ID
//    **GET /products/:id**

/*
   Response:
   {
     "id": 1,
     "user_id": 1,
     "product_name": "Sample Product",
     "product_description": "This is a sample product description",
     "product_images": ["http://example.com/image1.jpg"],
     "compressed_product_images": ["http://s3-bucket.com/compressed_image1.jpg"],
     "product_price": 99.99
   }
*/

// 3. Get Products for a User
//    **GET /products**

/*
   Query Parameters:
   - `user_id` (required): User ID for product filtering
   - `min_price` (optional): Minimum price range
   - `max_price` (optional): Maximum price range
*/

// Example Request:
//    GET /products?user_id=1&min_price=50&max_price=150

// ## Testing
// 1. Run Unit Tests:
//    ```bash
//    go test ./...
//    ```

// 2. Integration Testing:
//    Ensure RabbitMQ, Redis, and PostgreSQL are running before testing the full flow.

// ## Assumptions
// - Product images are publicly accessible URLs.
// - RabbitMQ, PostgreSQL, and Redis are correctly configured and available during runtime.
// - AWS S3 bucket setup is pending, and placeholder credentials are used for now.

// ## Architecture
// **Modular Design**:
// - Clear separation of concerns: API layer, asynchronous processing, caching, and logging.

// **Scalability**:
// - RabbitMQ enables distributed image processing.
// - Redis ensures efficient retrieval of frequently accessed data.

// **Transactional Consistency**:
// - Data consistency between database, cache, and message queues.

// ## Future Improvements
// - Implement API authentication (e.g., JWT).
// - Add support for multiple image formats (PNG, GIF).
// - Extend logging for real-time monitoring.
// - Set up distributed message queues (e.g., Kafka) for scalability.
