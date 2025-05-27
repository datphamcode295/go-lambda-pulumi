package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/apigatewayv2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
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

		// Create a custom policy for Systems Manager Parameter Store access
		ssmPolicy, err := iam.NewPolicy(ctx, "lambdaSSMPolicy", &iam.PolicyArgs{
			Description: pulumi.String("Allow Lambda to read from Systems Manager Parameter Store"),
			Policy: pulumi.String(`{
				"Version": "2012-10-17",
				"Statement": [
					{
						"Effect": "Allow",
						"Action": [
							"ssm:GetParameter",
							"ssm:GetParameters",
							"ssm:GetParametersByPath"
						],
						"Resource": [
							"arn:aws:ssm:*:*:parameter/app/*"
						]
					}
				]
			}`),
		})
		if err != nil {
			return err
		}

		// Attach the SSM policy to the Lambda role
		_, err = iam.NewRolePolicyAttachment(ctx, "lambdaSSMPolicyAttachment", &iam.RolePolicyAttachmentArgs{
			Role:      lambdaRole.Name,
			PolicyArn: ssmPolicy.Arn,
		})
		if err != nil {
			return err
		}

		// Create the Lambda function.
		function, err := lambda.NewFunction(ctx, "myGinLambda", &lambda.FunctionArgs{
			Handler: pulumi.String("bootstrap"),
			Role:    lambdaRole.Arn,
			Runtime: pulumi.String("provided.al2"),
			Code:    pulumi.NewFileArchive("../deployment.zip"),
			Architectures: pulumi.StringArray{
				pulumi.String("arm64"),
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
		})
		if err != nil {
			return err
		}

		// Create an integration between API Gateway and the Lambda function.
		lambdaIntegration, err := apigatewayv2.NewIntegration(ctx, "lambdaIntegration", &apigatewayv2.IntegrationArgs{
			ApiId:                api.ID(),
			IntegrationType:      pulumi.String("AWS_PROXY"),
			IntegrationUri:       function.Arn,
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
