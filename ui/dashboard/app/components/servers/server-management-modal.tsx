import { useState, useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { toast } from "sonner";
import { Loader2, Save } from "lucide-react";

import { Button } from "~/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "~/components/ui/dialog";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "~/components/ui/form";
import { Input } from "~/components/ui/input";
import { Textarea } from "~/components/ui/textarea";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "~/components/ui/tabs";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "~/components/ui/select";
import { serverService } from "~/lib/api";
import { Server } from "~/proto/controlplane/server";

interface ServerManagementModalProps {
  server: Server | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

const portSchema = z.object({
  port: z.coerce.number().min(1).max(65535),
});



export function ServerManagementModal({ server, open, onOpenChange }: ServerManagementModalProps) {
  const [activeTab, setActiveTab] = useState("overview");

  if (!server) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[80rem] h-[80vh] flex flex-col">
        <DialogHeader>
          <DialogTitle>Manage Server: {server.name}</DialogTitle>
          <DialogDescription>
            Configure agent settings, timezone, and view logs.
          </DialogDescription>
        </DialogHeader>

        <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col overflow-hidden">
          <TabsList>
            <TabsTrigger value="overview">Overview</TabsTrigger>
            <TabsTrigger value="configuration">Configuration</TabsTrigger>
            <TabsTrigger value="logs">Logs</TabsTrigger>
          </TabsList>

          <TabsContent value="overview" className="flex-1 overflow-auto p-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <h3 className="font-medium">Server Details</h3>
                <div className="grid grid-cols-[100px_1fr] gap-2 text-sm">
                  <span className="text-muted-foreground">ID:</span>
                  <span>{server.id}</span>
                  <span className="text-muted-foreground">IP Address:</span>
                  <span>{server.ipAddress}</span>
                  <span className="text-muted-foreground">SSH Port:</span>
                  <span>{server.port}</span>
                  <span className="text-muted-foreground">User:</span>
                  <span>{server.credential?.username}</span>
                </div>
              </div>
            </div>
          </TabsContent>

          <TabsContent value="configuration" className="flex-1 overflow-auto p-4 space-y-6">
            <div className="space-y-4">
              <h3 className="text-lg font-medium">WireGuard Configuration</h3>
              <div className="grid grid-cols-[150px_1fr] gap-2 text-sm border rounded-lg p-4 bg-muted/50">
                <span className="font-medium">Internal IP:</span>
                <span className="font-mono">{server.internalIp || "Not assigned"}</span>
                <span className="font-medium">Public Key:</span>
                <span className="font-mono break-all">{server.wireGuard?.publicKey || "Not generated"}</span>
                <span className="font-medium">Listen Port:</span>
                <span className="font-mono">{server.wireGuard?.listenPort || "Default"}</span>
              </div>
            </div>


          </TabsContent>

          <TabsContent value="logs" className="flex-1 overflow-auto p-4">
            <div className="text-muted-foreground text-center py-10">
              Agent logs viewer coming soon...
            </div>
          </TabsContent>
        </Tabs>
      </DialogContent>
    </Dialog>
  );
}




