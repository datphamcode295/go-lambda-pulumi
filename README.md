# Gin Lambda API

A serverless Go API built with Gin framework and deployed on AWS Lambda using Pulumi for infrastructure as code.

## Overview

This project provides a REST API for managing patient transactions, deployed as a serverless function on AWS Lambda. It uses Gin as the web framework and PostgreSQL as the database, with Pulumi handling the AWS infrastructure deployment.


## API Endpoints

### POST /app/patients/pay-transaction

Process a payment transaction for a patient.

**Example request:**
```
curl --location 'https://d90cvn773m.execute-api.ap-southeast-2.amazonaws.com/app/patients/pay-transaction' \
--header 'Content-Type: application/json' \
--data '{
	"patient_id": "9c7006ad-56e0-47cb-a166-f22426586cd2",
	"date_of_birth": "12-12-2000",
	"record_type": "NEW"
}'
```

**Example response:**
```
{
    "id": "b48e654b-e4dd-4614-b0b7-fba186f8d9bb",
    "patient_id": "9c7006ad-56e0-47cb-a166-f22426586cd2",
    "status": "success",
    "api_response": {
        "message": "Transaction success"
    },
    "record_type": "NEW",
    "date_of_birth": "12-12-2000",
    "created_at": "2025-05-27T17:36:13.774575422Z"
}
```

## Prerequisites

- Go 1.23.4+
- AWS CLI configured
- Pulumi CLI installed
- PostgreSQL database
- Make utility

## Local Development

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd gin-lambda-api
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Run tests:**
   ```bash
   make test
   # or for verbose output
   make test-verbose
   # or for coverage report
   make test-coverage
   ```

## Deployment

### Build the Lambda Function

```bash
make build
```

This command:
- Compiles the Go application for Linux ARM64 architecture
- Creates a `bootstrap` executable
- Packages it into `deployment.zip`

### Deploy Infrastructure

```bash
make deploy
```

This command:
- Deploys AWS infrastructure using Pulumi
- Creates Lambda function, API Gateway, IAM roles, and permissions
- Outputs the API endpoint URL

### Destroy Infrastructure

```bash
make destroy
```

## AWS Infrastructure

The Pulumi infrastructure creates:

- **Lambda Function**: Go runtime with ARM64 architecture
- **API Gateway v2**: HTTP API for routing requests
- **IAM Role**: With permissions for Lambda execution and SSM Parameter Store access
- **Integration**: Between API Gateway and Lambda function
