import { Duration } from "aws-cdk-lib";
import * as lambda from "aws-cdk-lib/aws-lambda";
import * as s3 from "aws-cdk-lib/aws-s3";
import * as s3n from "aws-cdk-lib/aws-s3-notifications";
import { Construct } from "constructs";

interface GoLambdaProps {
  functionName?: string;
  codePath: string;
  environment?: Record<string, string>;
  timeout?: Duration;
  memorySize?: number;
  s3Trigger?: {
    bucket: s3.IBucket;
    events: s3.EventType[];
    prefix?: string;
    suffix?: string;
  };
}

export class GoLambda extends Construct {
  public readonly fn: lambda.Function;

  constructor(scope: Construct, id: string, props: GoLambdaProps) {
    super(scope, id);

    this.fn = new lambda.Function(this, "Function", {
      runtime: lambda.Runtime.PROVIDED_AL2023,
      handler: "bootstrap",
      code: lambda.Code.fromAsset(props.codePath),
      functionName: props.functionName,
      timeout: props.timeout ?? Duration.seconds(30),
      memorySize: props.memorySize ?? 128,
      architecture: lambda.Architecture.ARM_64,
      environment: props.environment,
    });

    if (props.s3Trigger) {
      const { bucket, events, prefix, suffix } = props.s3Trigger;

      bucket.grantReadWrite(this.fn);

      const filters = prefix || suffix ? { prefix, suffix } : undefined;

      for (const event of events) {
        bucket.addEventNotification(
          event,
          new s3n.LambdaDestination(this.fn),
          ...(filters ? [filters] : []),
        );
      }
    }
  }
}
