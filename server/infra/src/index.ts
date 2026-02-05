import { App, Stack, RemovalPolicy, Duration, CfnOutput } from "aws-cdk-lib";
import * as s3 from "aws-cdk-lib/aws-s3";
import * as dynamodb from "aws-cdk-lib/aws-dynamodb";
import * as lambda from "aws-cdk-lib/aws-lambda";
import * as apigateway from "aws-cdk-lib/aws-apigateway";
import * as cloudfront from "aws-cdk-lib/aws-cloudfront";
import * as origins from "aws-cdk-lib/aws-cloudfront-origins";
import * as acm from "aws-cdk-lib/aws-certificatemanager";
import * as certificatemanager from "aws-cdk-lib/aws-certificatemanager";
import { environmentVars } from "./vars";

const isDev = environmentVars.ENV === "dev";
const isProd = environmentVars.ENV === "prod";

const apiDomainName = `${environmentVars.API_SUBDOMAIN}.${environmentVars.DOMAIN_NAME}`;
const mediaDomainName = `${environmentVars.MEDIA_SUBDOMAIN}.${environmentVars.DOMAIN_NAME}`;

const app = new App();

const stack = new Stack(app, "goattic", {
  env: {
    account: process.env.CDK_DEFAULT_ACCOUNT,
    region: process.env.CDK_DEFAULT_REGION,
  },
  crossRegionReferences: true, // needed to deploy cert for cloudfront
});

const privateBucket = new s3.Bucket(stack, "goattic-private", {
  bucketName: "goattic-storage-bucket-private",
  versioned: true,
  removalPolicy: isDev ? RemovalPolicy.DESTROY : RemovalPolicy.RETAIN,
  autoDeleteObjects: isDev ? true : false,
});

const publicBucket = new s3.Bucket(stack, "goattic-public", {
  bucketName: "goattic-storage-bucket-public",
  versioned: true,
  removalPolicy: isDev ? RemovalPolicy.DESTROY : RemovalPolicy.RETAIN,
  autoDeleteObjects: isDev ? true : false,
  publicReadAccess: true,
  blockPublicAccess: new s3.BlockPublicAccess({
    blockPublicAcls: false,
    blockPublicPolicy: false,
    ignorePublicAcls: false,
    restrictPublicBuckets: false,
  }),
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
  memorySize: 128,
  architecture: lambda.Architecture.ARM_64,
  environment: {
    PRIVATE_BUCKET_NAME: privateBucket.bucketName,
    PUBLIC_BUCKET_NAME: publicBucket.bucketName,
    TABLE_NAME: table.tableName,
  },
});

privateBucket.grantReadWrite(handler);
publicBucket.grantReadWrite(handler);
table.grantReadWriteData(handler);

const apiCertificate = new certificatemanager.Certificate(
  stack,
  "ApiCertificate",
  {
    domainName: apiDomainName,
    validation: certificatemanager.CertificateValidation.fromDns(),
  },
);

const apiDomain = new apigateway.DomainName(stack, "ApiDomain", {
  domainName: apiDomainName,
  certificate: apiCertificate,
  endpointType: apigateway.EndpointType.REGIONAL,
});

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

new apigateway.BasePathMapping(stack, "ApiBasePathMapping", {
  domainName: apiDomain,
  restApi: api,
  stage: api.deploymentStage,
});

const lambdaIntegration = new apigateway.LambdaIntegration(handler);

const proxy = api.root.addResource("{proxy+}");
proxy.addMethod("ANY", lambdaIntegration);
api.root.addMethod("ANY", lambdaIntegration);

// ACM CERT will be deployd in (us-east-1)
// Create a separate stack context for us-east-1 if needed
const mediaStack = new Stack(app, "GoAtticMediaStack", {
  env: {
    account: process.env.CDK_DEFAULT_ACCOUNT,
    region: "us-east-1", // CloudFront requires certs in us-east-1
  },
});

// ACM CERT for CloudFront in us-east-1
const mediaCertificate = new acm.Certificate(mediaStack, "MediaCertificate", {
  domainName: mediaDomainName,
  validation: acm.CertificateValidation.fromDns(), // keep DNS validation
});

const distribution = new cloudfront.Distribution(stack, "GoAtticDistribution", {
  defaultBehavior: {
    origin: origins.S3BucketOrigin.withBucketDefaults(publicBucket),
    viewerProtocolPolicy: cloudfront.ViewerProtocolPolicy.REDIRECT_TO_HTTPS,
    allowedMethods: cloudfront.AllowedMethods.ALLOW_GET_HEAD_OPTIONS,
    cachedMethods: cloudfront.CachedMethods.CACHE_GET_HEAD_OPTIONS,
  },
  domainNames: [mediaDomainName],
  certificate: mediaCertificate,
  priceClass: cloudfront.PriceClass.PRICE_CLASS_100,
});

new CfnOutput(stack, "ApiEndpoint", {
  value: api.url,
  description: "API Gateway default endpoint URL",
});

new CfnOutput(stack, "ApiCustomDomain", {
  value: `https://${apiDomainName}`,
  description: "API Gateway custom domain URL",
});

new CfnOutput(stack, "DNS_CNAME_API_Name", {
  value: apiDomainName,
  description: "DNS CNAME Record - Name",
});

new CfnOutput(stack, "DNS_CNAME_API_Value", {
  value: apiDomain.domainNameAliasDomainName,
  description: "DNS CNAME Record - Value",
});

new CfnOutput(stack, "MediaCustomDomain", {
  value: `https://${mediaDomainName}`,
  description: "CloudFront custom domain URL",
});

new CfnOutput(stack, "DNS_CNAME_Media_Name", {
  value: mediaDomainName,
  description: "DNS CNAME Record - Name",
});

new CfnOutput(stack, "DNS_CNAME_Media_Value", {
  value: distribution.domainName,
  description: "DNS CNAME Record - Value",
});
