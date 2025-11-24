import { z } from "zod";

export const databaseFormSchema = z.object({
  name: z.string().min(1, "Name is required"),
  user: z.string().min(1, "User is required"),
  password: z.string().min(1, "Password is required"),
  dbName: z.string().min(1, "Database Name is required"),
  branch: z.string().optional(), // or default to "main"
  serverId: z.string().min(1, "Server is required"),
});

export type DatabaseFormValues = z.infer<typeof databaseFormSchema>;
