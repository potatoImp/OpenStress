package pool

// Error handling module
// This file is responsible for handling errors in the goroutine pool to ensure task execution stability.
// Main functions include:
// - Define error types
// - Capture and log error information
// - Provide a unified error handling interface
// - Support custom error types (to be implemented)
// 
// Technical implementation details:
// 1. Define custom error types to describe different error situations.
// 2. Provide a unified error handling interface for easy error capture and handling.
// 3. Support custom error types to allow users to extend error handling capabilities.
// 4. Provide error logging functionality to log error information.
// 5. Implement error classification and statistics for easy error analysis and handling.
// 
// Common interface error return contents:
// - 400 Bad Request: Request parameter error
// - 401 Unauthorized: User unauthorized
// - 403 Forbidden: Access forbidden
// - 404 Not Found: Requested resource not found
// - 500 Internal Server Error: Internal server error
// - 503 Service Unavailable: Service unavailable
// 
// Other common interface return errors:
// - 408 Request Timeout: Request timeout
// - 429 Too Many Requests: Too many requests
// - 501 Not Implemented: Server does not support this functionality
// - 502 Bad Gateway: Bad gateway
// - 504 Gateway Timeout: Gateway timeout

import (
	"fmt"
	"log"
)

// CustomError Custom error type
// Used to describe different error situations
type CustomError struct {
	Message string
	Code    int
}

// Error Implement error interface
// Return error information
func (e *CustomError) Error() string {
	return fmt.Sprintf("Error %d: %s", e.Code, e.Message)
}

// HandleError Handle error and log
// Log level is ERROR
func HandleError(err error) {
	if err != nil {
		logger, logErr := NewStressLogger("logs/", "error.log", "ErrorModule")
		if logErr != nil {
			log.Println("Failed to create logger:", logErr)
			return
		}
		logger.Log("ERROR", "Error occurred: " + err.Error()) // Log error information
	}
}

// ExampleFunction Example function
// Demonstrate how to use custom error types
func ExampleFunction() {
	// Simulate an error
	err := &CustomError{Message: "Something went wrong", Code: 500}
	HandleError(err)
}
