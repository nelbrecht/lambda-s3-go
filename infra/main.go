package main

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/lambda"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type PulumiConfig struct {
	BucketBaseName   string
	AwsPolicyName    string
	LambdaS3RoleName string
	LambdaName       string
}

const (
	policy = `{
  "Statement": [
    {
      "Action": [
        "logs:PutLogEvents",
        "logs:CreateLogGroup",
        "logs:CreateLogStream"
      ],
      "Effect": "Allow",
      "Resource": "arn:aws:logs:*:*:*"
    },
    {
      "Action": [
        "s3:GetObject"
      ],
      "Effect": "Allow",
      "Resource": "arn:aws:s3:::%s/*"
    }
  ],
  "Version": "2012-10-17"
}`
	assumeRolePolicy = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}`
)

func getConfig(ctx *pulumi.Context) PulumiConfig {
	var pulumiConfig PulumiConfig

	cfg := config.New(ctx, "")
	cfg.RequireObject("testlambda", &pulumiConfig)

	return pulumiConfig
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		pcfg := getConfig(ctx)

		sourceBucket, err := s3.NewBucket(ctx, pcfg.BucketBaseName, &s3.BucketArgs{
			ForceDestroy: pulumi.Bool(true),
		})
		if err != nil {
			return err
		}

		awsLambdaS3Policy, err := iam.NewPolicy(ctx, pcfg.AwsPolicyName, &iam.PolicyArgs{
			Policy: pulumi.Any(fmt.Sprintf(policy, pcfg.BucketBaseName)),
		})
		if err != nil {
			return err
		}

		lambdaRole, err := iam.NewRole(ctx, pcfg.LambdaS3RoleName, &iam.RoleArgs{
			AssumeRolePolicy:  pulumi.Any(assumeRolePolicy),
			InlinePolicies:    iam.RoleInlinePolicyArray{nil},
			ManagedPolicyArns: pulumi.StringArray{awsLambdaS3Policy.Arn},
		}, pulumi.DependsOn([]pulumi.Resource{awsLambdaS3Policy}))
		if err != nil {
			return err
		}

		testLambda, err := lambda.NewFunction(ctx, pcfg.LambdaName, &lambda.FunctionArgs{
			Handler: pulumi.String("handler"),
			Role:    lambdaRole.Arn,
			Runtime: pulumi.String("go1.x"),
			Code:    pulumi.NewFileArchive("../testlambda/testlambda.zip"),
			Architectures: pulumi.StringArray{
				pulumi.String("x86_64"),
			},
			Timeout: pulumi.Int(10),
			TracingConfig: &lambda.FunctionTracingConfigArgs{
				Mode: pulumi.String("PassThrough"),
			},
		}, pulumi.Protect(false))
		if err != nil {
			return err
		}

		allowBucketToInvokeF, err := lambda.NewPermission(ctx, "pulumiAllowBucketToInvokeF", &lambda.PermissionArgs{
			Action:    pulumi.String("lambda:InvokeFunction"),
			Function:  testLambda.Arn,
			Principal: pulumi.String("s3.amazonaws.com"),
			SourceArn: sourceBucket.Arn,
		})
		if err != nil {
			return err
		}

		_, err = s3.NewBucketNotification(ctx, "newObjectLambdaTrigger", &s3.BucketNotificationArgs{
			Bucket: sourceBucket.ID(),
			LambdaFunctions: s3.BucketNotificationLambdaFunctionArray{
				&s3.BucketNotificationLambdaFunctionArgs{
					LambdaFunctionArn: testLambda.Arn,
					Events: pulumi.StringArray{
						pulumi.String("s3:ObjectCreated:*"),
					},
				},
			},
		}, pulumi.DependsOn([]pulumi.Resource{allowBucketToInvokeF}))
		if err != nil {
			return err
		}

		return nil
	})
}
