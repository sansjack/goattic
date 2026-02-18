import { App } from "aws-cdk-lib";
import * as certificatemanager from "aws-cdk-lib/aws-certificatemanager";
import { environmentVars } from "./vars";
import { ApplicationStack } from "./stacks/application-stack";

const app = new App();

const apiDomainName = `${environmentVars.API_SUBDOMAIN}.${environmentVars.DOMAIN_NAME}`;
const mediaDomainName = `${environmentVars.MEDIA_SUBDOMAIN}.${environmentVars.DOMAIN_NAME}`;

// Use existing certificate from us-east-1
const cloudFrontCertificate = certificatemanager.Certificate.fromCertificateArn(
  app,
  "ImportedCertificate",
  environmentVars.MEDIA_CERT_ARN,
);

const applicationStack = new ApplicationStack(app, "GoAtticApplication", {
  environment: environmentVars.ENV,
  apiDomainName,
  mediaDomainName,
  cloudFrontCertificate,
  env: {
    account: process.env.CDK_DEFAULT_ACCOUNT,
    region: process.env.CDK_DEFAULT_REGION,
  },
});
