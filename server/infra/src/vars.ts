import { z } from "zod";

const vars = z.object({
  ENV: z.enum(["dev", "prod"]).default("dev"),
  CDK_DEFAULT_ACCOUNT: z.string().optional(),
  CDK_DEFAULT_REGION: z.string().optional(),
  DOMAIN_NAME: z.string(),
  MEDIA_SUBDOMAIN: z.string(),
  API_SUBDOMAIN: z.string(),
});

export const environmentVars = vars.parse(process.env);
