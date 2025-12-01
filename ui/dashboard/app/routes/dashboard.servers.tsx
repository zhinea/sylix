import { zodResolver } from "@hookform/resolvers/zod";
import { Plus } from "lucide-react";
import { useEffect, useState, useRef } from "react";
import { useForm } from "react-hook-form";
import { useActionData, useNavigation, useSubmit } from "react-router";
import { toast } from "sonner";

import { Button } from "~/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "~/components/ui/card";

import { AgentLogsModal } from "~/components/servers/agent-logs-modal";
import { DeleteServerAlert } from "~/components/servers/delete-server-alert";
import { serverFormSchema } from "~/components/servers/schema";
import type { ServerFormValues } from "~/components/servers/schema";
import { ServerFormDialog } from "~/components/servers/server-form-dialog";
import { ServerList } from "~/components/servers/server-list";
import { serverService } from "~/lib/api";
import { AgentStatusServer, Server, StatusCode, StatusServer } from "~/proto/controlplane/server";
import type { Route } from "./+types/dashboard.servers";

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
        status: StatusServer.STATUS_SERVER_UNSPECIFIED,
        agent: {
          port: 0,
          status: AgentStatusServer.AGENT_STATUS_SERVER_UNSPECIFIED,
          logs: "",
        },
      };

      const response = await serverService.Create(server);

      if (response.status === StatusCode.CREATED || response.status === StatusCode.OK) {
        if (response.server?.status === StatusServer.CONNECTED) {
          return { success: true, message: "Server credentials have been successfully connected", close: true, server: response.server };
        } else {
          return { success: true, warning: "Connection failed, credentials may be incorrect, or UFW is running", close: false, server: response.server };
        }
      }
      return { success: false, error: response.error || "Failed to create server" };
    }

    if (intent === "delete") {
      const id = formData.get("id") as string;
      await serverService.Delete({ id });
      return { success: true, message: "Server deleted successfully" };
    }

    if (intent === "retry-connection") {
      const id = formData.get("id") as string;
      const response = await serverService.RetryConnection({ id });
      if (response.status === StatusCode.OK) {
        if (response.server?.status === StatusServer.CONNECTED) {
          return { success: true, message: "Server connected successfully" };
        } else {
          return { success: true, warning: "Connection failed" };
        }
      }
      return { success: false, error: response.error || "Failed to retry connection" };
    }

    if (intent === "update") {
      const data = Object.fromEntries(formData);
      const server: Server = {
        id: data.id as string,
        name: data.name as string,
        ipAddress: data.ipAddress as string,
        port: Number(data.port),
        protocol: (data.protocol as string) || "ssh",
        credential: {
          username: data.username as string,
          password: (data.password as string) || undefined,
          sshKey: (data.sshKey as string) || undefined,
        },
        isRoot: 0,
        status: StatusServer.STATUS_SERVER_UNSPECIFIED,
        agent: {
          port: 0,
          status: AgentStatusServer.AGENT_STATUS_SERVER_UNSPECIFIED,
          logs: "",
        },
      };

      const response = await serverService.Update(server);
      if (response.status === StatusCode.OK) {
        if (response.server?.status === StatusServer.CONNECTED) {
          return { success: true, message: "Server updated successfully", close: true, server: response.server };
        } else {
          return { success: true, warning: "Server updated but connection failed", close: false, server: response.server };
        }
      }
      return { success: false, error: response.error || "Failed to update server" };
    }

    if (intent === "provision-node") {
      const id = formData.get("id") as string;
      await serverService.InstallAgent({ id }); // Maps to ProvisionNode in backend
      return { success: true, message: "Node provisioning triggered" };
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
  const [serverToUpdate, setServerToUpdate] = useState<Server | null>(null);
  const [installingServerId, setInstallingServerId] = useState<string | null>(null);
  const submit = useSubmit();
  const navigation = useNavigation();
  const isSubmitting = navigation.state === "submitting";
  const isDeleting =
    navigation.state === "submitting" &&
    navigation.formData?.get("intent") === "delete";
  const retryingServerId =
    navigation.state === "submitting" &&
      navigation.formData?.get("intent") === "retry-connection"
      ? navigation.formData.get("id") as string
      : null;
  const actionData = useActionData<typeof clientAction>();
  const wasDeleting = useRef(false);

  // Form handling
  const form = useForm<ServerFormValues>({
    resolver: zodResolver(serverFormSchema as any),
    defaultValues: {
      port: 22,
      username: "root",
      protocol: "ssh",
    },
  });

  useEffect(() => {
    if (isDeleting) {
      wasDeleting.current = true;
    }
    if (!isDeleting && wasDeleting.current && navigation.state === "idle") {
      wasDeleting.current = false;
      if (actionData?.success) {
        setServerToDelete(null);
        toast.success("Server deleted successfully");
      } else if (actionData?.error) {
        toast.error(actionData.error);
      }
    }
  }, [isDeleting, navigation.state, actionData]);

  useEffect(() => {
    if (actionData) {
      if (actionData.success) {
        if (actionData.warning) {
          toast.error(actionData.warning);
        } else if (actionData.message) {
          toast.success(actionData.message);
        }

        if (actionData.close) {
          setIsCreateOpen(false);
          setServerToUpdate(null);
          form.reset();
        } else if (actionData.server) {
          if (isCreateOpen) {
            setIsCreateOpen(false);
            handleUpdate(actionData.server);
          } else if (serverToUpdate) {
            handleUpdate(actionData.server);
          }
        }
      } else if (actionData.error) {
        toast.error(actionData.error);
      }
    }
  }, [actionData, form]);

  const onSubmit = (data: ServerFormValues) => {
    const formData = new FormData();
    formData.append("intent", "create");
    Object.entries(data).forEach(([key, value]) => {
      if (value) formData.append(key, value.toString());
    });

    submit(formData, { method: "post" });
    toast.info("Creating server...");
  };

  const onUpdateSubmit = (data: ServerFormValues) => {
    const formData = new FormData();
    formData.append("intent", "update");
    formData.append("id", serverToUpdate?.id || "");
    Object.entries(data).forEach(([key, value]) => {
      if (value) formData.append(key, value.toString());
    });

    submit(formData, { method: "post" });
    toast.info("Updating server...");
  };

  const handleDelete = () => {
    if (serverToDelete) {
      const formData = new FormData();
      formData.append("intent", "delete");
      formData.append("id", serverToDelete.id);
      submit(formData, { method: "post" });
      toast.info("Deleting server...");
    }
  };

  const handleProvisionNode = (id: string) => {
    const formData = new FormData();
    formData.append("intent", "provision-node");
    formData.append("id", id);
    submit(formData, { method: "post" });
    setInstallingServerId(id);
    toast.info("Triggering node provisioning...");
  }

  const handleRetryConnection = (id: string) => {
    const formData = new FormData();
    formData.append("intent", "retry-connection");
    formData.append("id", id);
    submit(formData, { method: "post" });
    toast.info("Retrying connection...");
  };

  const handleUpdate = (server: Server) => {
    setServerToUpdate(server);
    form.reset({
      name: server.name,
      ipAddress: server.ipAddress,
      port: server.port,
      username: server.credential?.username,
      password: server.credential?.password || "",
      sshKey: server.credential?.sshKey || "",
      protocol: server.protocol,
    });
  };

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
        <ServerList
          servers={servers}
          onDelete={setServerToDelete}
          onInstallAgent={handleProvisionNode}
          onRetryConnection={handleRetryConnection}
          onUpdate={handleUpdate}
          retryingServerId={retryingServerId}
        />
      )}

      <ServerFormDialog
        open={isCreateOpen}
        onOpenChange={setIsCreateOpen}
        form={form as any}
        onSubmit={onSubmit}
        isSubmitting={isSubmitting}
        mode="create"
        status={StatusServer.STATUS_SERVER_UNSPECIFIED}
      />

      <ServerFormDialog
        open={!!serverToUpdate}
        onOpenChange={(open) => {
          if (!open) {
            setServerToUpdate(null);
            form.reset();
          }
        }}
        form={form as any}
        onSubmit={onUpdateSubmit}
        isSubmitting={isSubmitting}
        mode="edit"
        status={serverToUpdate?.status || StatusServer.STATUS_SERVER_UNSPECIFIED}
      />

      <DeleteServerAlert
        server={serverToDelete}
        onClose={() => setServerToDelete(null)}
        onConfirm={handleDelete}
        isDeleting={isDeleting}
      />

      <AgentLogsModal
        serverId={installingServerId}
        onClose={() => setInstallingServerId(null)}
      />
    </div>
  );
}
