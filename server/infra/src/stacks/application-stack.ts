import { Stack, type StackProps, RemovalPolicy, CfnOutput } from "aws-cdk-lib";
import * as s3 from "aws-cdk-lib/aws-s3";
import * as dynamodb from "aws-cdk-lib/aws-dynamodb";
import * as apigateway from "aws-cdk-lib/aws-apigateway";
import * as cloudfront from "aws-cdk-lib/aws-cloudfront";
import * as origins from "aws-cdk-lib/aws-cloudfront-origins";
import * as certificatemanager from "aws-cdk-lib/aws-certificatemanager";
import type { Construct } from "constructs";
import type { ICertificate } from "aws-cdk-lib/aws-certificatemanager";

import { environmentVars } from "../vars";
import { GoLambda } from "../constructs/lambda";

interface ApplicationStackProps extends StackProps {
  environment: "dev" | "prod";
  apiDomainName: string;
  mediaDomainName: string;
  cloudFrontCertificate: ICertificate;
}

export class ApplicationStack extends Stack {
  constructor(scope: Construct, id: string, props: ApplicationStackProps) {
    super(scope, id, props);

    const isDev = props.environment === "dev";

    const privateBucket = new s3.Bucket(this, "PrivateBucket", {
      bucketName: "goattic-storage-bucket-private",
      versioned: false,
      removalPolicy: isDev ? RemovalPolicy.DESTROY : RemovalPolicy.RETAIN,
      autoDeleteObjects: isDev,
    });

    const publicBucket = new s3.Bucket(this, "PublicBucket", {
      bucketName: "goattic-storage-bucket-public",
      versioned: false,
      removalPolicy: isDev ? RemovalPolicy.DESTROY : RemovalPolicy.RETAIN,
      autoDeleteObjects: isDev,
      blockPublicAccess: s3.BlockPublicAccess.BLOCK_ALL,
    });

    const table = new dynamodb.Table(this, "ApiKeysTable", {
      tableName: "goattic-apikeys",
      partitionKey: {
        name: "id",
        type: dynamodb.AttributeType.STRING,
      },
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
      removalPolicy: isDev ? RemovalPolicy.DESTROY : RemovalPolicy.RETAIN,
    });

    table.addGlobalSecondaryIndex({
      indexName: "ApiKeyIndex",
      partitionKey: {
        name: "apiKey",
        type: dynamodb.AttributeType.STRING,
      },
      projectionType: dynamodb.ProjectionType.ALL,
    });

    table.addGlobalSecondaryIndex({
      indexName: "OwnerIndex",
      partitionKey: {
        name: "owner",
        type: dynamodb.AttributeType.STRING,
      },
      projectionType: dynamodb.ProjectionType.ALL,
    });

    const apiHandler = new GoLambda(this, "ApiHandler", {
      codePath: "../lambdas/api/build",
      environment: {
        PRIVATE_BUCKET_NAME: privateBucket.bucketName,
        PUBLIC_BUCKET_NAME: publicBucket.bucketName,
        TABLE_NAME: table.tableName,
        MEDIA_DOMAIN: props.mediaDomainName,
      },
    });

    const fileValidator = new GoLambda(this, "FileValidator", {
      codePath: "../lambdas/file-validator/build",
      environment: {
        DISCORD_WEBHOOK_URL: environmentVars.DISCORD_WEBHOOK_URL,
        PUBLIC_BUCKET_NAME: publicBucket.bucketName,
        MEDIA_DOMAIN: props.mediaDomainName,
      },
      s3Trigger: {
        bucket: privateBucket,
        events: [s3.EventType.OBJECT_CREATED],
      },
    });

    privateBucket.grantReadWrite(apiHandler.fn);
    publicBucket.grantReadWrite(apiHandler.fn);
    publicBucket.grantReadWrite(fileValidator.fn);
    table.grantReadWriteData(apiHandler.fn);

    const apiCertificate = certificatemanager.Certificate.fromCertificateArn(
      this,
      "ApiCertificate",
      environmentVars.API_CERT_ARN,
    );

    const apiDomain = new apigateway.DomainName(this, "ApiDomain", {
      domainName: props.apiDomainName,
      certificate: apiCertificate,
      endpointType: apigateway.EndpointType.REGIONAL,
    });

    const api = new apigateway.RestApi(this, "RestApi", {
      restApiName: "GoAttic API",
      description: "API Gateway for goattic service",
      deployOptions: {
        stageName: props.environment,
      },
      defaultCorsPreflightOptions: {
        allowOrigins: apigateway.Cors.ALL_ORIGINS,
        allowMethods: apigateway.Cors.ALL_METHODS,
      },
    });

    new apigateway.BasePathMapping(this, "ApiBasePathMapping", {
      domainName: apiDomain,
      restApi: api,
      stage: api.deploymentStage,
    });

    const lambdaIntegration = new apigateway.LambdaIntegration(apiHandler.fn);
    const proxy = api.root.addResource("{proxy+}");
    proxy.addMethod("ANY", lambdaIntegration);
    api.root.addMethod("ANY", lambdaIntegration);

    const distribution = new cloudfront.Distribution(
      this,
      "MediaDistribution",
      {
        defaultBehavior: {
          origin: origins.S3BucketOrigin.withOriginAccessControl(publicBucket),
          viewerProtocolPolicy:
            cloudfront.ViewerProtocolPolicy.REDIRECT_TO_HTTPS,
          allowedMethods: cloudfront.AllowedMethods.ALLOW_GET_HEAD_OPTIONS,
          cachedMethods: cloudfront.CachedMethods.CACHE_GET_HEAD_OPTIONS,
        },
        domainNames: [props.mediaDomainName],
        certificate: props.cloudFrontCertificate,
        priceClass: cloudfront.PriceClass.PRICE_CLASS_100,
      },
    );

    new CfnOutput(this, "ApiEndpoint", {
      value: api.url,
      description: "API Gateway default endpoint URL",
    });

    new CfnOutput(this, "ApiCustomDomain", {
      value: `https://${props.apiDomainName}`,
      description: "API Gateway custom domain URL",
    });

    new CfnOutput(this, "ApiDomainCNAMEName", {
      value: props.apiDomainName,
      description: "DNS CNAME Record - Name",
    });

    new CfnOutput(this, "ApiDomainCNAMEValue", {
      value: apiDomain.domainNameAliasDomainName,
      description: "DNS CNAME Record - Value",
    });

    new CfnOutput(this, "MediaCustomDomain", {
      value: `https://${props.mediaDomainName}`,
      description: "CloudFront custom domain URL",
    });

    new CfnOutput(this, "MediaDomainCNAMEName", {
      value: props.mediaDomainName,
      description: "DNS CNAME Record - Name",
    });

    new CfnOutput(this, "MediaDomainCNAMEValue", {
      value: distribution.domainName,
      description: "DNS CNAME Record - Value",
    });
  }
}
