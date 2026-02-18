import { Stack, type StackProps } from "aws-cdk-lib";
import * as acm from "aws-cdk-lib/aws-certificatemanager";
import type { Construct } from "constructs";

interface CloudFrontCertificateStackProps extends StackProps {
  domainName: string;
}

/**
 * Stack for CloudFront SSL certificate
 * Must be deployed in us-east-1, needed for cloudfront
 */
export class CloudFrontCertificateStack extends Stack {
  public readonly certificate: acm.Certificate;

  constructor(scope: Construct, id: string, props: CloudFrontCertificateStackProps) {
    super(scope, id, {
      ...props,
      env: {
        account: props.env?.account,
        region: "us-east-1",
      },
    });

    this.certificate = new acm.Certificate(this, "Certificate", {
      domainName: props.domainName,
      validation: acm.CertificateValidation.fromDns(),
    });
  }
}
