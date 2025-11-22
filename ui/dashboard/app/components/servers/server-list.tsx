import { useState } from "react";
import { MoreHorizontal, RefreshCw, Server as ServerIcon, Trash, FileText } from "lucide-react";
import { Badge } from "~/components/ui/badge";
import { Button } from "~/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "~/components/ui/dropdown-menu";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "~/components/ui/table";
import { Server, StatusServer, AgentStatusServer } from "~/proto/server/server";
import { ServerLogsDialog } from "./server-logs-dialog";

interface ServerListProps {
  servers: Server[];
  onDelete: (server: Server) => void;
  onInstallAgent: (id: string) => void;
}

export function ServerList({ servers, onDelete, onInstallAgent }: ServerListProps) {
  const [logsServerId, setLogsServerId] = useState<string | null>(null);

  const getStatusBadge = (status: StatusServer) => {
    switch (status) {
      case StatusServer.CONNECTED:
        return <Badge variant="default" className="bg-green-500 hover:bg-green-600">Connected</Badge>;
      case StatusServer.DISCONNECTED:
        return <Badge variant="destructive">Disconnected</Badge>;
      default:
        return <Badge variant="secondary">Unknown</Badge>;
    }
  };

  const getAgentStatusBadge = (status: AgentStatusServer) => {
    switch (status) {
      case AgentStatusServer.SUCCESS:
        return <Badge variant="outline" className="border-green-500 text-green-500">Installed</Badge>;
      case AgentStatusServer.INSTALLING:
      case AgentStatusServer.CONFIGURING:
      case AgentStatusServer.FINALIZING_SETUP:
        return <Badge variant="outline" className="border-blue-500 text-blue-500 animate-pulse">Installing</Badge>;
      case AgentStatusServer.FAILED:
        return <Badge variant="outline" className="border-red-500 text-red-500">Failed</Badge>;
      default:
        return <Badge variant="outline" className="text-muted-foreground">Not Installed</Badge>;
    }
  };

  return (
    <>
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Address</TableHead>
              <TableHead>User</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Agent Status</TableHead>
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
                <TableCell>
                  {server.ipAddress}:{server.port}
                </TableCell>
                <TableCell>{server.credential?.username}</TableCell>
                <TableCell>
                  {getStatusBadge(server.status)}
                </TableCell>
                <TableCell>
                  {getAgentStatusBadge(server.agentStatus)}
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
                      <DropdownMenuItem
                        onClick={() => navigator.clipboard.writeText(server.ipAddress)}
                      >
                        Copy IP Address
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem onClick={() => setLogsServerId(server.id)}>
                        <FileText className="mr-2 h-4 w-4" /> Logs
                      </DropdownMenuItem>
                      <DropdownMenuItem onClick={() => onInstallAgent(server.id)}>
                        <RefreshCw className="mr-2 h-4 w-4" /> Install Agent
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem
                        className="text-destructive focus:text-destructive"
                        onClick={() => onDelete(server)}
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
      <ServerLogsDialog
        serverId={logsServerId}
        open={!!logsServerId}
        onOpenChange={(open) => !open && setLogsServerId(null)}
      />
    </>
  );
}
