package config

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

type Config struct {
	DatabaseURL string
	APIKey      string
}

func NewConfig() *Config {
	// Get AWS region from environment variable or use default
	region := os.Getenv("AWS_REGION")
	fmt.Println("AWS_REGION", region)
	if region == "" {
		region = "ap-southeast-2"
	}

	// Create AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		log.Fatalf("Failed to create AWS session: %v", err)
	}

	// Create SSM client
	ssmClient := ssm.New(sess)

	// Get parameters from Systems Manager Parameter Store
	databaseURL, err := getParameter(ssmClient, "/app/databaseURL")
	if err != nil {
		log.Fatalf("Failed to get DATABASE_URL parameter: %v", err)
	}

	apiKey, err := getParameter(ssmClient, "/app/submitPatientApiKey")
	if err != nil {
		log.Fatalf("Failed to get API_KEY parameter: %v", err)
	}

	return &Config{
		DatabaseURL: databaseURL,
		APIKey:      apiKey,
	}
}

func getParameter(ssmClient *ssm.SSM, parameterName string) (string, error) {
	input := &ssm.GetParameterInput{
		Name:           aws.String(parameterName),
		WithDecryption: aws.Bool(true),
	}

	result, err := ssmClient.GetParameter(input)
	if err != nil {
		return "", err
	}

	return *result.Parameter.Value, nil
}
