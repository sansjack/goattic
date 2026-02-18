import { z } from "zod";

const vars = z.object({
  ENV: z.enum(["dev", "prod"]).default("dev"),
  CDK_DEFAULT_ACCOUNT: z.string().optional(),
  CDK_DEFAULT_REGION: z.string().optional(),
  DOMAIN_NAME: z.string(),
  MEDIA_SUBDOMAIN: z.string(),
  API_SUBDOMAIN: z.string(),
  MEDIA_CERT_ARN: z.string(),
  API_CERT_ARN: z.string(),
  DISCORD_WEBHOOK_URL: z.string(),
});

export const environmentVars = vars.parse(process.env);
