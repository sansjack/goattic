#!/usr/bin/env bun
import { DynamoDBClient, PutItemCommand } from "@aws-sdk/client-dynamodb";
import { marshall } from "@aws-sdk/util-dynamodb";
import seedData from "../config/seed-data.json";

const TABLE_NAME = "goattic-apikeys";

// local stack
const dynamoConfig: any = {
  region: process.env.AWS_REGION || "us-east-1",
};

if (process.env.AWS_ENDPOINT_URL) {
  dynamoConfig.endpoint = process.env.AWS_ENDPOINT_URL;
  console.log(`Using LocalStack endpoint: ${process.env.AWS_ENDPOINT_URL}`);
}

const dynamodb = new DynamoDBClient(dynamoConfig);

async function seedDatabase() {
  console.log(`Seeding table ${TABLE_NAME} with ${seedData.length} items...`);

  for (const item of seedData) {
    try {
      await dynamodb.send(
        new PutItemCommand({
          TableName: TABLE_NAME,
          Item: marshall(item),
        }),
      );
      console.log(`✓ Seeded: ${item.apiKey} (${item.owner})`);
    } catch (error) {
      console.error(`✗ Failed to seed ${item.apiKey}:`, error);
    }
  }

  console.log("Seeding complete!");
}

seedDatabase().catch((error) => {
  console.error("Seeding failed:", error);
  process.exit(1);
});
