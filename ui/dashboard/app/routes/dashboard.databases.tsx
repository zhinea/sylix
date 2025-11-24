import type { Route } from "./+types/dashboard.databases";
import { databaseService, serverService } from "~/lib/api";
import { DatabaseFormDialog } from "~/components/databases/database-form-dialog";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "~/components/ui/table";
import { Badge } from "~/components/ui/badge";
import { Button } from "~/components/ui/button";
import { Trash2 } from "lucide-react";
import { useFetcher } from "react-router";
import { databaseFormSchema } from "~/components/databases/schema";

export async function clientLoader() {
  const [dbResponse, serverResponse] = await Promise.all([
    databaseService.All({}),
    serverService.All({}),
  ]);
  return {
    databases: dbResponse.databases,
    servers: serverResponse.servers,
  };
}

export async function clientAction({ request }: Route.ClientActionArgs) {
  const formData = await request.formData();
  const intent = formData.get("intent");

  if (intent === "create") {
    const values = Object.fromEntries(formData);
    const result = databaseFormSchema.safeParse(values);

    if (!result.success) {
      return { success: false, errors: result.error.flatten() };
    }

    try {
      await databaseService.Create({
        name: result.data.name,
        user: result.data.user,
        password: result.data.password,
        dbName: result.data.dbName,
        branch: result.data.branch,
        serverId: result.data.serverId,
        // These are set by backend/agent
        id: "",
        status: "",
        containerId: "",
        port: 0,
      });
      return { success: true };
    } catch (error) {
      console.error("Failed to create database:", error);
      return { success: false, error: "Failed to create database" };
    }
  }

  if (intent === "delete") {
    const id = formData.get("id") as string;
    try {
      await databaseService.Delete({ id });
      return { success: true };
    } catch (error) {
      console.error("Failed to delete database:", error);
      return { success: false, error: "Failed to delete database" };
    }
  }

  return { success: false };
}

export default function DatabasesPage({ loaderData }: Route.ComponentProps) {
  const { databases, servers } = loaderData;
  const fetcher = useFetcher();

  const getServerName = (id: string) => {
    return servers.find((s) => s.id === id)?.name || id;
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case "RUNNING":
        return "default"; // primary/black
      case "CREATING":
        return "secondary"; // gray
      case "ERROR":
        return "destructive"; // red
      default:
        return "outline";
    }
  };

  return (
    <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Databases</h1>
          <p className="text-muted-foreground">
            Manage your Postgres instances and branches.
          </p>
        </div>
        <DatabaseFormDialog servers={servers} />
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Database</TableHead>
              <TableHead>Branch</TableHead>
              <TableHead>Server</TableHead>
              <TableHead>Port</TableHead>
              <TableHead>Status</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {databases.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} className="h-24 text-center">
                  No databases found.
                </TableCell>
              </TableRow>
            ) : (
              databases.map((db) => (
                <TableRow key={db.id}>
                  <TableCell className="font-medium">{db.name}</TableCell>
                  <TableCell>{db.dbName}</TableCell>
                  <TableCell>
                    <Badge variant="outline">{db.branch}</Badge>
                  </TableCell>
                  <TableCell>{getServerName(db.serverId)}</TableCell>
                  <TableCell>{db.port > 0 ? db.port : "-"}</TableCell>
                  <TableCell>
                    <Badge variant={getStatusColor(db.status) as any}>
                      {db.status}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-right">
                    <fetcher.Form method="post" className="inline-block">
                      <input type="hidden" name="intent" value="delete" />
                      <input type="hidden" name="id" value={db.id} />
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 text-destructive"
                        type="submit"
                        disabled={fetcher.state !== "idle"}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </fetcher.Form>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
