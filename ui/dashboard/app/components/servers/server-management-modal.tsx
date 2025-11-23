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

const timezoneSchema = z.object({
  timezone: z.string().min(1),
});

const configSchema = z.object({
  config: z.string().min(1),
});

export function ServerManagementModal({ server, open, onOpenChange }: ServerManagementModalProps) {
  const [activeTab, setActiveTab] = useState("overview");

  if (!server) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl h-[80vh] flex flex-col">
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
              <h3 className="text-lg font-medium">Agent Port</h3>
              <AgentPortForm serverId={server.id} initialPort={server.agent?.port ?? 8083} />
            </div>
            
            <div className="space-y-4">
              <h3 className="text-lg font-medium">Timezone (Chrony)</h3>
              <TimezoneForm serverId={server.id} />
            </div>

            <div className="space-y-4">
              <h3 className="text-lg font-medium">Agent Configuration</h3>
              <AgentConfigForm serverId={server.id} />
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

function AgentPortForm({ serverId, initialPort }: { serverId: string, initialPort: number }) {
  const form = useForm<z.infer<typeof portSchema>>({
    resolver: zodResolver(portSchema as any),
    defaultValues: { port: initialPort || 8083 },
  });

  async function onSubmit(values: z.infer<typeof portSchema>) {
    try {
      await serverService.UpdateAgentPort({ serverId, port: values.port });
      toast.success("Agent port updated successfully");
    } catch (error) {
      toast.error("Failed to update agent port");
      console.error(error);
    }
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="flex items-end gap-4">
        <FormField
          control={form.control}
          name="port"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Port</FormLabel>
              <FormControl>
                <Input {...field} type="number" />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <Button type="submit" disabled={form.formState.isSubmitting}>
          {form.formState.isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          Update Port
        </Button>
      </form>
    </Form>
  );
}

function TimezoneForm({ serverId }: { serverId: string }) {
  const form = useForm<z.infer<typeof timezoneSchema>>({
    resolver: zodResolver(timezoneSchema as any),
    defaultValues: { timezone: "UTC" },
  });

  useEffect(() => {
    const fetchConfig = async () => {
      try {
        const response = await serverService.GetAgentConfig({ id: serverId });
        if (response.timezone) {
          form.setValue("timezone", response.timezone);
        }
      } catch (error) {
        console.error("Failed to fetch agent config", error);
      }
    };
    fetchConfig();
  }, [serverId, form]);

  async function onSubmit(values: z.infer<typeof timezoneSchema>) {
    try {
      await serverService.UpdateServerTimeZone({ serverId, timezone: values.timezone });
      toast.success("Timezone updated successfully");
    } catch (error) {
      toast.error("Failed to update timezone");
      console.error(error);
    }
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="flex items-end gap-4">
        <FormField
          control={form.control}
          name="timezone"
          render={({ field }) => (
            <FormItem className="w-[200px]">
              <FormLabel>Timezone</FormLabel>
              <Select onValueChange={field.onChange} defaultValue={field.value}>
                <FormControl>
                  <SelectTrigger>
                    <SelectValue placeholder="Select timezone" />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  <SelectItem value="UTC">UTC</SelectItem>
                  <SelectItem value="Asia/Jakarta">Asia/Jakarta</SelectItem>
                  <SelectItem value="America/New_York">America/New_York</SelectItem>
                  <SelectItem value="Europe/London">Europe/London</SelectItem>
                  {/* Add more timezones as needed */}
                </SelectContent>
              </Select>
              <FormMessage />
            </FormItem>
          )}
        />
        <Button type="submit" disabled={form.formState.isSubmitting}>
          {form.formState.isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          Set Timezone
        </Button>
      </form>
    </Form>
  );
}

function AgentConfigForm({ serverId }: { serverId: string }) {
  const form = useForm<z.infer<typeof configSchema>>({
    resolver: zodResolver(configSchema as any),
    defaultValues: { config: "" },
  });

  useEffect(() => {
    const fetchConfig = async () => {
      try {
        const response = await serverService.GetAgentConfig({ id: serverId });
        if (response.config) {
          form.setValue("config", response.config);
        }
      } catch (error) {
        console.error("Failed to fetch agent config", error);
      }
    };
    fetchConfig();
  }, [serverId, form]);

  async function onSubmit(values: z.infer<typeof configSchema>) {
    try {
      await serverService.ConfigureAgent({ serverId, config: values.config });
      toast.success("Configuration applied successfully");
    } catch (error) {
      toast.error("Failed to apply configuration");
      console.error(error);
    }
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
        <FormField
          control={form.control}
          name="config"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Configuration (YAML/JSON)</FormLabel>
              <FormControl>
                <Textarea {...field} className="font-mono min-h-[200px]" />
              </FormControl>
              <FormDescription>
                Enter the agent configuration content.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
        <Button type="submit" disabled={form.formState.isSubmitting}>
          {form.formState.isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          <Save className="mr-2 h-4 w-4" />
          Apply Configuration
        </Button>
      </form>
    </Form>
  );
}
