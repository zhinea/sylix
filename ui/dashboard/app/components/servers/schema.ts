import { z } from "zod";

export const serverFormSchema = z.object({
  name: z.string().min(1, "Name is required"),
  ipAddress: z.string().min(1, "IP Address is required"),
  port: z.coerce.number().min(1).max(65535).default(22),
  username: z.string().min(1, "Username is required"),
  password: z.string().optional(),
  sshKey: z.string().optional(),
  protocol: z.string().default("ssh"),
}).refine((data) => data.password || data.sshKey, {
  message: "Either password or SSH key is required",
  path: ["password"],
});

export type ServerFormValues = z.infer<typeof serverFormSchema>;
