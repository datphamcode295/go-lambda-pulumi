package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/datphamcode295/go-lambda-pulumi/internal/adapters/handler"
	"github.com/datphamcode295/go-lambda-pulumi/internal/adapters/repository"
	"github.com/datphamcode295/go-lambda-pulumi/internal/config"
	"github.com/datphamcode295/go-lambda-pulumi/internal/core/domain"
	"github.com/datphamcode295/go-lambda-pulumi/internal/core/services"
	"github.com/datphamcode295/go-lambda-pulumi/internal/logger"
	util "github.com/datphamcode295/go-lambda-pulumi/internal/utils"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

var (
	patientService *services.PatientService
	ginLambda      *ginadapter.GinLambdaV2
)

func Handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// parse json request for debugging
	reqJson, err := json.Marshal(req)
	if err != nil {
		log.Println("Error marshalling request", err)
	}

	log.Println("Request received", string(reqJson))
	return ginLambda.ProxyWithContext(ctx, req)
}

func init() {
	// cfg := config.NewConfig()
	cfg := &config.Config{
		DatabaseURL: "",
		APIKey:      "1234567890",
	}
	db, err := gorm.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		panic(err)
	}

	logger.SetupLogger()

	// Create or modify the database tables based on the model structs found in the imported package
	db.AutoMigrate(&domain.User{}, &domain.Patient{}, &domain.Transaction{})

	store := repository.NewDB(db)

	patientService = services.NewPatientService(cfg, store, store)

	InitRoutes()
}

func InitRoutes() {
	router := gin.Default()
	// Register custom validator
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("ddmmyyyy", util.ValidateDDMMYYYY)
	}

	pprof.Register(router)

	v1 := router.Group("/app")

	patientHandler := handler.NewPatientHandler(*patientService)
	v1.POST("/patients/pay-transaction", patientHandler.PayTransaction)

	// err := router.Run(":4242")
	// if err != nil {
	// 	log.Fatalf("Error starting server: %v", err)
	// }

	ginLambda = ginadapter.NewV2(router)
}

func main() {
	lambda.Start(Handler)
}
