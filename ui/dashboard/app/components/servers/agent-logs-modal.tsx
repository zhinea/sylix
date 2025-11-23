import { useEffect, useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "~/components/ui/dialog";
import { serverService, logsService } from "~/lib/api";
import { AgentStatusServer } from "~/proto/controlplane/server";

interface AgentLogsModalProps {
  serverId: string | null;
  onClose: () => void;
}

export function AgentLogsModal({ serverId, onClose }: AgentLogsModalProps) {
  const [logs, setLogs] = useState("");
  const [status, setStatus] = useState<AgentStatusServer>(
    AgentStatusServer.AGENT_STATUS_SERVER_UNSPECIFIED
  );

  useEffect(() => {
    if (!serverId) return;

    const fetchData = async () => {
      try {
        const serverResponse = await serverService.Get({ id: serverId });
        if (serverResponse.server) {
          setStatus(serverResponse.server.agentStatus);
        }

        const logsResponse = await logsService.ReadServerLog({
          serverId,
          filename: "setup_agent.log",
          page: 1,
          pageSize: 10000,
        });
        setLogs(logsResponse.lines.join("\n"));
      } catch (e) {
        console.error(e);
      }
    };

    fetchData();
    const interval = setInterval(fetchData, 2000);

    return () => clearInterval(interval);
  }, [serverId]);

  return (
    <Dialog open={!!serverId} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-w-3xl h-[80vh] flex flex-col">
        <DialogHeader>
          <DialogTitle>Agent Installation Logs</DialogTitle>
          <DialogDescription>
            Status: {AgentStatusServer[status]}
          </DialogDescription>
        </DialogHeader>
        <div className="flex-1 bg-black text-white p-4 rounded overflow-auto font-mono text-sm whitespace-pre-wrap">
          {logs || "Waiting for logs..."}
        </div>
      </DialogContent>
    </Dialog>
  );
}
