package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
)

var ginLambda *ginadapter.GinLambdaV2

func main() {
	lambda.Start(Handler)
}

func GetShortURL(c *gin.Context) {
	shortcode := c.Param("shortcode")
	c.JSON(http.StatusOK, gin.H{"shortcode": shortcode})
}

func init() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello, World!"})
	})

	// access url
	r.GET("/app/:shortcode", GetShortURL)

	ginLambda = ginadapter.NewV2(r)
}

func Handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// parse json request
	reqJson, err := json.Marshal(req)
	if err != nil {
		log.Println("Error marshalling request", err)
	}

	log.Println("Request received", string(reqJson))
	return ginLambda.ProxyWithContext(ctx, req)
}
