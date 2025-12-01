import type { UseFormReturn } from "react-hook-form";
import { Badge } from "~/components/ui/badge";
import { Button } from "~/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/components/ui/dialog";
import { Input } from "~/components/ui/input";
import { Label } from "~/components/ui/label";
import { Textarea } from "~/components/ui/textarea";
import { StatusServer } from "~/proto/controlplane/server";
import type { ServerFormValues } from "./schema";

interface ServerFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  form: UseFormReturn<ServerFormValues>;
  onSubmit: (data: ServerFormValues) => void;
  isSubmitting: boolean;
  mode?: "create" | "edit";
  status?: StatusServer;
}

export function ServerFormDialog({
  open,
  onOpenChange,
  form,
  onSubmit,
  isSubmitting,
  mode = "create",
  status = StatusServer.STATUS_SERVER_UNSPECIFIED,
}: ServerFormDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>{mode === "create" ? "Add Server" : "Edit Server"}</DialogTitle>
          <DialogDescription>
            {mode === "create"
              ? "Enter the details of the server you want to manage."
              : "Update the details of your server."}
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
                SSH Port
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
          <DialogFooter className="flex items-center justify-between sm:justify-between">
            <div className="flex items-center gap-2">
              {status === StatusServer.CONNECTED && (
                <Badge variant="default" className="bg-green-500 hover:bg-green-600">
                  Connected
                </Badge>
              )}
              {status === StatusServer.DISCONNECTED && (
                <Badge variant="destructive">Disconnected</Badge>
              )}
            </div>
            <Button type="submit" isLoading={isSubmitting}>
              {mode === "create" ? "Add Server" : "Save Changes"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
