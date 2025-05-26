package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
)

var ginLambda *ginadapter.GinLambda

func init() {
	// stdout and stderr are sent to AWS CloudWatch Logs
	log.Printf("Gin cold start")
	r := gin.Default()

	// Define your routes
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/hello/:name", func(c *gin.Context) {
		name := c.Param("name")
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello " + name + "!",
		})
	})

	// If you need to serve static files or use other Gin features,
	// configure them here.

	ginLambda = ginadapter.New(r)
}

// Handler is the Lambda function handler
func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// If no name is provided in the HTTP request, use the AWS_LAMBDA_FUNCTION_NAME environment variable
	log.Printf("Received rawPath: %s, path: %s", req.RequestContext.Path, req.Path) // Add this
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	// Check if running in Lambda environment
	if _, ok := os.LookupEnv("AWS_LAMBDA_FUNCTION_NAME"); ok {
		// Start Lambda handler
		lambda.Start(Handler)
	} else {
		// Start local Gin server (for local development)
		log.Println("Starting local Gin server on :8080")
		localRouter := gin.Default()
		localRouter.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "pong (local)"})
		})
		localRouter.GET("/hello/:name", func(c *gin.Context) {
			name := c.Param("name")
			c.JSON(http.StatusOK, gin.H{"message": "Hello " + name + "! (local)"})
		})
		if err := localRouter.Run(":8080"); err != nil {
			log.Fatalf("Failed to run local server: %v", err)
		}
	}
}
