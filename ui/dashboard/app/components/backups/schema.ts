import { z } from "zod";

export const backupFormSchema = z.object({
  name: z.string().min(1, "Name is required"),
  endpoint: z.string().min(1, "Endpoint is required"),
  region: z.string().min(1, "Region is required"),
  bucket: z.string().min(1, "Bucket is required"),
  accessKey: z.string().min(1, "Access Key is required"),
  secretKey: z.string().min(1, "Secret Key is required"),
});

export type BackupFormValues = z.infer<typeof backupFormSchema>;
