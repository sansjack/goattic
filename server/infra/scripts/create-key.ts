#!/usr/bin/env bun
import { randomBytes } from "crypto";
import { DynamoDBClient, PutItemCommand } from "@aws-sdk/client-dynamodb";
import { marshall } from "@aws-sdk/util-dynamodb";

const TABLE_NAME = "goattic-apikeys";

const owner = process.argv[2];
if (!owner) {
  console.error("Usage: bun run scripts/create-key.ts <owner>");
  process.exit(1);
}

const dynamoConfig: any = {
  region: process.env.AWS_REGION || "us-east-1",
};

if (process.env.AWS_ENDPOINT_URL) {
  dynamoConfig.endpoint = process.env.AWS_ENDPOINT_URL;
}

const dynamodb = new DynamoDBClient(dynamoConfig);

const id = randomBytes(8).toString("hex");
const apiKey = randomBytes(32).toString("base64");

await dynamodb.send(
  new PutItemCommand({
    TableName: TABLE_NAME,
    Item: marshall({
      id,
      apiKey,
      owner,
      createdAt: new Date().toISOString(),
      rateLimit: 100,
      enabled: true,
    }),
  }),
);

console.log("");
console.log("API Key created:");
console.log(`  ID:    ${id}`);
console.log(`  Key:   ${apiKey}`);
console.log(`  Owner: ${owner}`);
