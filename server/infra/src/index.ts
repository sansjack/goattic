import { App, Stack, RemovalPolicy, Duration, CfnOutput } from "aws-cdk-lib";
import * as s3 from "aws-cdk-lib/aws-s3";
import * as dynamodb from "aws-cdk-lib/aws-dynamodb";
import * as lambda from "aws-cdk-lib/aws-lambda";
import * as apigateway from "aws-cdk-lib/aws-apigateway";
import { environmentVars } from "./vars";

const isDev = environmentVars.ENV === "dev";
const isProd = environmentVars.ENV === "prod";

const app = new App();

const stack = new Stack(app, "goattic", {
  env: {
    account: process.env.CDK_DEFAULT_ACCOUNT,
    region: process.env.CDK_DEFAULT_REGION,
  },
});

const bucket = new s3.Bucket(stack, "goattic", {
  bucketName: "goattic-storage-bucket-1",
  versioned: true,
  removalPolicy: isDev ? RemovalPolicy.DESTROY : RemovalPolicy.RETAIN,
  autoDeleteObjects: isDev ? true : false,
});

const table = new dynamodb.Table(stack, "GoAtticTable", {
  tableName: "goattic-apikeys",
  partitionKey: {
    name: "apiKey",
    type: dynamodb.AttributeType.STRING,
  },
  billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
  removalPolicy: isDev ? RemovalPolicy.DESTROY : RemovalPolicy.RETAIN,
});

table.addGlobalSecondaryIndex({
  indexName: "OwnerIndex",
  partitionKey: {
    name: "owner",
    type: dynamodb.AttributeType.STRING,
  },
  projectionType: dynamodb.ProjectionType.ALL,
});

const handler = new lambda.Function(stack, "GoAtticHandler", {
  runtime: lambda.Runtime.PROVIDED_AL2023,
  handler: "bootstrap",
  code: lambda.Code.fromAsset("../lambda/build"),
  functionName: "goattic-handler",
  timeout: Duration.seconds(30),
  memorySize: 256,
  architecture: lambda.Architecture.ARM_64,
  environment: {
    BUCKET_NAME: bucket.bucketName,
    TABLE_NAME: table.tableName,
  },
});

bucket.grantReadWrite(handler);
table.grantReadWriteData(handler);

const api = new apigateway.RestApi(stack, "GoAtticAPI", {
  restApiName: "GoAttic API",
  description: "API Gateway for goattic service",
  deployOptions: {
    stageName: environmentVars.ENV,
  },
  defaultCorsPreflightOptions: {
    allowOrigins: apigateway.Cors.ALL_ORIGINS,
    allowMethods: apigateway.Cors.ALL_METHODS,
  },
});

const lambdaIntegration = new apigateway.LambdaIntegration(handler);

const proxy = api.root.addResource("{proxy+}");
proxy.addMethod("ANY", lambdaIntegration);
api.root.addMethod("ANY", lambdaIntegration);

new CfnOutput(stack, "ApiEndpoint", {
  value: api.url,
  description: "API Gateway endpoint URL",
});
