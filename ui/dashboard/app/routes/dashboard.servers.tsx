import { useState } from "react";
import { Form, useActionData, useLoaderData, useNavigation, useSubmit } from "react-router";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { Loader2, MoreHorizontal, Plus, Server as ServerIcon, Trash, RefreshCw } from "lucide-react";
import { toast } from "sonner";

import { Button } from "~/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "~/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "~/components/ui/dropdown-menu";
import { Input } from "~/components/ui/input";
import { Label } from "~/components/ui/label";
import { Textarea } from "~/components/ui/textarea";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "~/components/ui/table";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "~/components/ui/alert-dialog";
import { Badge } from "~/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "~/components/ui/card";

import { serverService } from "~/lib/api";
import { Server, StatusServer } from "~/proto/server/server";
import type { Route } from "./+types/dashboard.servers";

// --- Zod Schema ---
const serverFormSchema = z.object({
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

type ServerFormValues = z.infer<typeof serverFormSchema>;

// --- Loader ---
export async function clientLoader() {
  try {
    const response = await serverService.All({});
    return { servers: response.servers };
  } catch (error) {
    console.error("Failed to fetch servers:", error);
    return { servers: [] };
  }
}

// --- Action ---
export async function clientAction({ request }: Route.ClientActionArgs) {
  const formData = await request.formData();
  const intent = formData.get("intent");

  try {
    if (intent === "create") {
      const data = Object.fromEntries(formData);
      
      const server: Server = {
        id: "",
        name: data.name as string,
        ipAddress: data.ipAddress as string,
        port: Number(data.port),
        protocol: (data.protocol as string) || "ssh",
        credential: {
          username: data.username as string,
          password: (data.password as string) || undefined,
          sshKey: (data.sshKey as string) || undefined,
        },
        isRoot: 0, // Default value
      };

      await serverService.Create(server);
      return { success: true, message: "Server created successfully" };
    }

    if (intent === "delete") {
      const id = formData.get("id") as string;
      await serverService.Delete({ id });
      return { success: true, message: "Server deleted successfully" };
    }
    
    if (intent === "install-agent") {
        const id = formData.get("id") as string;
        await serverService.InstallAgent({ id });
        return { success: true, message: "Agent installation triggered" };
    }

    return { success: false, error: "Unknown intent" };
  } catch (error: any) {
    console.error("Action failed:", error);
    return { success: false, error: error.message || "Operation failed" };
  }
}

// --- Component ---
export default function ServersPage({ loaderData }: Route.ComponentProps) {
  const { servers } = loaderData;
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [serverToDelete, setServerToDelete] = useState<Server | null>(null);
  const submit = useSubmit();
  const navigation = useNavigation();
  const isSubmitting = navigation.state === "submitting";

  // Form handling
  const form = useForm<ServerFormValues>({
    resolver: zodResolver(serverFormSchema as any),
    defaultValues: {
      port: 22,
      username: "root",
      protocol: "ssh",
    },
  });

  const onSubmit = (data: ServerFormValues) => {
    const formData = new FormData();
    formData.append("intent", "create");
    Object.entries(data).forEach(([key, value]) => {
      if (value) formData.append(key, value.toString());
    });
    
    submit(formData, { method: "post" });
    setIsCreateOpen(false);
    form.reset();
    toast.success("Creating server...");
  };

  const handleDelete = () => {
    if (serverToDelete) {
      const formData = new FormData();
      formData.append("intent", "delete");
      formData.append("id", serverToDelete.id);
      submit(formData, { method: "post" });
      setServerToDelete(null);
      toast.success("Deleting server...");
    }
  };
  
  const handleInstallAgent = (id: string) => {
      const formData = new FormData();
      formData.append("intent", "install-agent");
      formData.append("id", id);
      submit(formData, { method: "post" });
      toast.info("Triggering agent installation...");
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">Servers</h2>
          <p className="text-muted-foreground">
            Manage your database servers and agents.
          </p>
        </div>
        <Button onClick={() => setIsCreateOpen(true)}>
          <Plus className="mr-2 h-4 w-4" /> Add Server
        </Button>
      </div>

      {servers.length === 0 ? (
        <Card>
          <CardHeader>
            <CardTitle>No servers found</CardTitle>
            <CardDescription>
              Get started by adding your first server.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button variant="outline" onClick={() => setIsCreateOpen(true)}>
              Add Server
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div className="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Address</TableHead>
                <TableHead>User</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {servers.map((server: Server) => (
                <TableRow key={server.id}>
                  <TableCell className="font-medium">
                    <div className="flex items-center gap-2">
                      <ServerIcon className="h-4 w-4 text-muted-foreground" />
                      {server.name}
                    </div>
                  </TableCell>
                  <TableCell>{server.ipAddress}:{server.port}</TableCell>
                  <TableCell>{server.credential?.username}</TableCell>
                  <TableCell>
                    <Badge variant="secondary">Unknown</Badge>
                  </TableCell>
                  <TableCell className="text-right">
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" className="h-8 w-8 p-0">
                          <span className="sr-only">Open menu</span>
                          <MoreHorizontal className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuLabel>Actions</DropdownMenuLabel>
                        <DropdownMenuItem onClick={() => navigator.clipboard.writeText(server.ipAddress)}>
                          Copy IP Address
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem onClick={() => handleInstallAgent(server.id)}>
                            <RefreshCw className="mr-2 h-4 w-4" /> Install Agent
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem
                          className="text-destructive focus:text-destructive"
                          onClick={() => setServerToDelete(server)}
                        >
                          <Trash className="mr-2 h-4 w-4" /> Delete
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      {/* Create Dialog */}
      <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>Add Server</DialogTitle>
            <DialogDescription>
              Enter the details of the server you want to manage.
            </DialogDescription>
          </DialogHeader>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <div className="grid gap-4 py-4">
              <div className="grid grid-cols-4 items-center gap-4">
                <Label htmlFor="name" className="text-right">
                  Name
                </Label>
                <Input
                  id="name"
                  className="col-span-3"
                  {...form.register("name")}
                />
              </div>
              <div className="grid grid-cols-4 items-center gap-4">
                <Label htmlFor="ipAddress" className="text-right">
                  IP Address
                </Label>
                <Input
                  id="ipAddress"
                  className="col-span-3"
                  {...form.register("ipAddress")}
                />
              </div>
              <div className="grid grid-cols-4 items-center gap-4">
                <Label htmlFor="port" className="text-right">
                  Port
                </Label>
                <Input
                  id="port"
                  type="number"
                  className="col-span-3"
                  {...form.register("port")}
                />
              </div>
              <div className="grid grid-cols-4 items-center gap-4">
                <Label htmlFor="username" className="text-right">
                  Username
                </Label>
                <Input
                  id="username"
                  className="col-span-3"
                  {...form.register("username")}
                />
              </div>
              <div className="grid grid-cols-4 items-center gap-4">
                <Label htmlFor="password" className="text-right">
                  Password
                </Label>
                <div className="col-span-3">
                  <Input
                    id="password"
                    type="password"
                    {...form.register("password")}
                  />
                  {form.formState.errors.password && (
                    <p className="text-sm text-destructive mt-1">
                      {form.formState.errors.password.message}
                    </p>
                  )}
                </div>
              </div>
              <div className="grid grid-cols-4 items-center gap-4">
                <Label htmlFor="sshKey" className="text-right">
                  SSH Key
                </Label>
                <div className="col-span-3">
                  <Textarea
                    id="sshKey"
                    placeholder="-----BEGIN OPENSSH PRIVATE KEY-----"
                    className="min-h-[100px] font-mono text-xs"
                    {...form.register("sshKey")}
                  />
                  {form.formState.errors.sshKey && (
                    <p className="text-sm text-destructive mt-1">
                      {form.formState.errors.sshKey.message}
                    </p>
                  )}
                </div>
              </div>
            </div>
            <DialogFooter>
              <Button type="submit" disabled={isSubmitting}>
                {isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                Save changes
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Delete Alert */}
      <AlertDialog open={!!serverToDelete} onOpenChange={(open: boolean) => !open && setServerToDelete(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
            <AlertDialogDescription>
              This action cannot be undone. This will permanently delete the server
              <span className="font-semibold"> {serverToDelete?.name} </span>
              from the dashboard.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={handleDelete} className="bg-destructive text-destructive-foreground hover:bg-destructive/90">
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
