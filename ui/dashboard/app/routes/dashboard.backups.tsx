import { Plus, Trash2, Edit } from "lucide-react";
import { useState, useEffect } from "react";
import { useActionData, useLoaderData, useNavigation, useSubmit } from "react-router";
import { toast } from "sonner";

import { Button } from "~/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "~/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "~/components/ui/table";
import { Badge } from "~/components/ui/badge";

import { BackupFormDialog } from "~/components/backups/backup-form-dialog";
import type { BackupFormValues } from "~/components/backups/schema";
import { backupStorageService } from "~/lib/api";
import { BackupStorage, BackupStatusCode } from "~/proto/controlplane/backup";
import type { Route } from "./+types/dashboard.backups";

// --- Loader ---
export async function clientLoader() {
  try {
    const response = await backupStorageService.All({});
    return { backups: response.data || [] };
  } catch (error) {
    console.error("Failed to fetch backups:", error);
    return { backups: [] };
  }
}

// --- Action ---
export async function clientAction({ request }: Route.ClientActionArgs) {
  const formData = await request.formData();
  const intent = formData.get("intent");

  try {
    if (intent === "create") {
      const data = Object.fromEntries(formData);
      const backup: BackupStorage = {
        id: "",
        name: data.name as string,
        endpoint: data.endpoint as string,
        region: data.region as string,
        bucket: data.bucket as string,
        accessKey: data.accessKey as string,
        secretKey: data.secretKey as string,
        status: "",
        errorMessage: "",
      };

      const response = await backupStorageService.Create(backup);
      
      if (response.status === BackupStatusCode.BACKUP_CREATED || response.status === BackupStatusCode.BACKUP_OK) {
          return { success: true, message: "Backup storage created and connected successfully", close: true };
      }
      return { success: false, error: response.error || "Failed to create backup storage" };
    }

    if (intent === "update") {
      const data = Object.fromEntries(formData);
      const backup: BackupStorage = {
        id: data.id as string,
        name: data.name as string,
        endpoint: data.endpoint as string,
        region: data.region as string,
        bucket: data.bucket as string,
        accessKey: data.accessKey as string,
        secretKey: data.secretKey as string,
        status: "",
        errorMessage: "",
      };

      const response = await backupStorageService.Update(backup);
      
      if (response.status === BackupStatusCode.BACKUP_OK) {
          return { success: true, message: "Backup storage updated and connected successfully", close: true };
      }
      return { success: false, error: response.error || "Failed to update backup storage" };
    }

    if (intent === "delete") {
      const id = formData.get("id") as string;
      await backupStorageService.Delete({ id });
      return { success: true, message: "Backup storage deleted successfully" };
    }

  } catch (error: any) {
    return { success: false, error: error.message || "An unexpected error occurred" };
  }
  return null;
}

export default function DashboardBackups() {
  const { backups } = useLoaderData<typeof clientLoader>();
  const actionData = useActionData<typeof clientAction>();
  const submit = useSubmit();
  const navigation = useNavigation();
  
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [editingBackup, setEditingBackup] = useState<BackupStorage | null>(null);

  const isSubmitting = navigation.state === "submitting";

  useEffect(() => {
    if (actionData?.success) {
      toast.success(actionData.message);
      if (actionData.close) {
        setIsCreateOpen(false);
        setEditingBackup(null);
      }
    } else if (actionData?.error) {
      toast.error(actionData.error);
    }
  }, [actionData]);

  const handleCreate = (values: BackupFormValues) => {
    const formData = new FormData();
    formData.append("intent", "create");
    Object.entries(values).forEach(([key, value]) => formData.append(key, value));
    submit(formData, { method: "post" });
  };

  const handleUpdate = (values: BackupFormValues) => {
    if (!editingBackup) return;
    const formData = new FormData();
    formData.append("intent", "update");
    formData.append("id", editingBackup.id);
    Object.entries(values).forEach(([key, value]) => formData.append(key, value));
    submit(formData, { method: "post" });
  };

  const handleDelete = (id: string) => {
    if (confirm("Are you sure you want to delete this backup storage?")) {
      const formData = new FormData();
      formData.append("intent", "delete");
      formData.append("id", id);
      submit(formData, { method: "post" });
    }
  };

  return (
    <div className="flex flex-col gap-4 p-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Backup Storage</h1>
          <p className="text-muted-foreground">
            Manage your S3-compatible backup storage locations.
          </p>
        </div>
        <Button onClick={() => setIsCreateOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Add Storage
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Storage Locations</CardTitle>
          <CardDescription>
            List of configured backup storage locations.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Endpoint</TableHead>
                <TableHead>Bucket</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {backups.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={5} className="text-center py-8 text-muted-foreground">
                    No backup storage configured.
                  </TableCell>
                </TableRow>
              ) : (
                backups.map((backup) => (
                  <TableRow key={backup.id}>
                    <TableCell className="font-medium">{backup.name}</TableCell>
                    <TableCell>{backup.endpoint}</TableCell>
                    <TableCell>{backup.bucket}</TableCell>
                    <TableCell>
                      {backup.status === "CONNECTED" ? (
                        <Badge variant="default" className="bg-green-500 hover:bg-green-600">Connected</Badge>
                      ) : (
                        <Badge variant="destructive">Error</Badge>
                      )}
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-2">
                        <Button variant="ghost" size="icon" onClick={() => setEditingBackup(backup)}>
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button variant="ghost" size="icon" onClick={() => handleDelete(backup.id)}>
                          <Trash2 className="h-4 w-4 text-red-500" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <BackupFormDialog
        open={isCreateOpen}
        onOpenChange={setIsCreateOpen}
        onSubmit={handleCreate}
        isSubmitting={isSubmitting}
      />

      <BackupFormDialog
        open={!!editingBackup}
        onOpenChange={(open) => !open && setEditingBackup(null)}
        onSubmit={handleUpdate}
        initialData={editingBackup}
        isSubmitting={isSubmitting}
      />
    </div>
  );
}
