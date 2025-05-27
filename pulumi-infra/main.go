package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/apigatewayv2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Get the AWS region from the Pulumi config.
		// awsRegion := config.Require(ctx, "aws:region") // Unused, so commented out or removed

		// Create an IAM role for the Lambda function.
		lambdaRole, err := iam.NewRole(ctx, "lambdaRole", &iam.RoleArgs{
			AssumeRolePolicy: pulumi.String(`{
				"Version": "2012-10-17",
				"Statement": [{
					"Action": "sts:AssumeRole",
					"Effect": "Allow",
					"Principal": {
						"Service": "lambda.amazonaws.com"
					}
				}]
			}`),
		})
		if err != nil {
			return err
		}

		// Attach the AWSLambdaBasicExecutionRole policy to the Lambda role.
		_, err = iam.NewRolePolicyAttachment(ctx, "lambdaPolicyAttachment", &iam.RolePolicyAttachmentArgs{
			Role:      lambdaRole.Name,
			PolicyArn: pulumi.String("arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"),
		})
		if err != nil {
			return err
		}

		// Create the Lambda function.
		// The Code path is relative to this pulumi_infra.go file, so it points to the parent directory.
		function, err := lambda.NewFunction(ctx, "myGinLambda", &lambda.FunctionArgs{
			Handler: pulumi.String("bootstrap"), // This should be the name of the compiled binary in handler.zip
			Role:    lambdaRole.Arn,
			Runtime: pulumi.String("provided.al2"),              // Amazon Linux 2 custom runtime for Go
			Code:    pulumi.NewFileArchive("../deployment.zip"), // Path to the zipped deployment package
			Architectures: pulumi.StringArray{
				pulumi.String("arm64"), // or "x86_64"
			},
			MemorySize: pulumi.Int(128),
			Timeout:    pulumi.Int(300),
			Environment: &lambda.FunctionEnvironmentArgs{
				Variables: pulumi.StringMap{
					"GIN_MODE": pulumi.String("release"),
				},
			},
		})
		if err != nil {
			return err
		}

		// Create an API Gateway v2 HTTP API.
		api, err := apigatewayv2.NewApi(ctx, "httpApi", &apigatewayv2.ApiArgs{
			ProtocolType: pulumi.String("HTTP"),
			// Target:       function.InvokeArn, // For direct Lambda integration this is simpler
		})
		if err != nil {
			return err
		}

		// Create an integration between API Gateway and the Lambda function.
		lambdaIntegration, err := apigatewayv2.NewIntegration(ctx, "lambdaIntegration", &apigatewayv2.IntegrationArgs{
			ApiId:                api.ID(),
			IntegrationType:      pulumi.String("AWS_PROXY"),
			IntegrationUri:       function.Arn, // Use function.Arn for integration URI
			PayloadFormatVersion: pulumi.String("2.0"),
		})
		if err != nil {
			return err
		}

		// Create a default route that invokes the Lambda integration.
		_, err = apigatewayv2.NewRoute(ctx, "defaultRoute", &apigatewayv2.RouteArgs{
			ApiId:    api.ID(),
			RouteKey: pulumi.String("$default"),
			Target:   pulumi.Sprintf("integrations/%s", lambdaIntegration.ID()),
		})
		if err != nil {
			return err
		}

		// Create a stage and deploy the API.
		_, err = apigatewayv2.NewStage(ctx, "apiStage", &apigatewayv2.StageArgs{
			ApiId:      api.ID(),
			Name:       pulumi.String("app"), // Default stage
			AutoDeploy: pulumi.Bool(true),
		})
		if err != nil {
			return err
		}

		// Grant API Gateway permission to invoke the Lambda function.
		// The SourceArn needs to be constructed carefully for HTTP APIs.
		// It should use the ExecutionArn of the API Gateway API.
		_, err = lambda.NewPermission(ctx, "apiGatewayPermission", &lambda.PermissionArgs{
			Action:    pulumi.String("lambda:InvokeFunction"),
			Function:  function.Name,
			Principal: pulumi.String("apigateway.amazonaws.com"),
			SourceArn: pulumi.Sprintf("%s/*/*", api.ExecutionArn),
		})
		if err != nil {
			return err
		}

		// Export the API endpoint URL.
		ctx.Export("apiUrl", api.ApiEndpoint)

		return nil
	})
}
