import { useEffect, useState } from "react";
import { toast } from "sonner";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "~/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/components/ui/select";
import { Button } from "~/components/ui/button";
import { LogFile } from "~/proto/controlplane/logs";
import { logsService } from "~/lib/api";
import { ChevronLeft, ChevronRight, Loader2 } from "lucide-react";

interface ServerLogsDialogProps {
  serverId: string | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function ServerLogsDialog({
  serverId,
  open,
  onOpenChange,
}: ServerLogsDialogProps) {
  const [logs, setLogs] = useState<LogFile[]>([]);
  const [selectedLog, setSelectedLog] = useState<string>("");
  const [logContent, setLogContent] = useState<string[]>([]);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [loading, setLoading] = useState(false);
  const [loadingContent, setLoadingContent] = useState(false);

  useEffect(() => {
    if (open && serverId) {
      fetchLogs();
    } else {
      // Reset state when closed
      setLogs([]);
      setSelectedLog("");
      setLogContent([]);
      setPage(1);
      setTotalPages(1);
    }
  }, [open, serverId]);

  useEffect(() => {
    if (serverId && selectedLog) {
      fetchLogContent(1);
    }
  }, [selectedLog, serverId]);

  const fetchLogs = async () => {
    if (!serverId) return;
    setLoading(true);
    try {
      const response = await logsService.GetServerLogs({ serverId });
      setLogs(response.files);
      if (response.files.length > 0) {
        // Optionally select the first log or the most recent one
        // setSelectedLog(response.files[0].name);
      }
    } catch (error) {
      console.error("Failed to fetch logs:", error);
      toast.error("Failed to fetch logs");
    } finally {
      setLoading(false);
    }
  };

  const fetchLogContent = async (pageNum: number) => {
    if (!serverId || !selectedLog) return;
    setLoadingContent(true);
    try {
      const response = await logsService.ReadServerLog({
        serverId,
        filename: selectedLog,
        page: pageNum,
        pageSize: 100,
      });
      setLogContent(response.lines);
      setPage(response.currentPage);
      setTotalPages(response.totalPages);
    } catch (error) {
      console.error("Failed to read log:", error);
      toast.error("Failed to read log content");
    } finally {
      setLoadingContent(false);
    }
  };

  const handlePageChange = (newPage: number) => {
    if (newPage >= 1 && newPage <= totalPages) {
      fetchLogContent(newPage);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl h-[80vh] flex flex-col">
        <DialogHeader>
          <DialogTitle>Server Logs</DialogTitle>
        </DialogHeader>

        <div className="flex items-center gap-4 py-4">
          <Select
            value={selectedLog}
            onValueChange={(value) => setSelectedLog(value)}
          >
            <SelectTrigger className="w-[300px]">
              <SelectValue placeholder="Select a log file" />
            </SelectTrigger>
            <SelectContent>
              {logs.map((log) => (
                <SelectItem key={log.name} value={log.name}>
                  {log.name} ({formatBytes(log.size)})
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          {loading && <Loader2 className="h-4 w-4 animate-spin" />}
        </div>

        <div className="flex-1 border rounded-md bg-muted/50 p-4 overflow-auto font-mono text-xs">
          {loadingContent ? (
            <div className="flex items-center justify-center h-full">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
          ) : selectedLog ? (
            logContent.length > 0 ? (
              logContent.map((line, index) => (
                <div key={index} className="whitespace-pre-wrap">
                  {line}
                </div>
              ))
            ) : (
              <div className="text-center text-muted-foreground py-8">
                Empty log file
              </div>
            )
          ) : (
            <div className="text-center text-muted-foreground py-8">
              Select a log file to view content
            </div>
          )}
        </div>

        <div className="flex items-center justify-between pt-4">
          <div className="text-sm text-muted-foreground">
            Page {page} of {totalPages}
          </div>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => handlePageChange(page - 1)}
              disabled={page <= 1 || loadingContent}
            >
              <ChevronLeft className="h-4 w-4" />
              Previous
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => handlePageChange(page + 1)}
              disabled={page >= totalPages || loadingContent}
            >
              Next
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}

function formatBytes(bytes: number, decimals = 2) {
  if (!+bytes) return "0 Bytes";

  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ["Bytes", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"];

  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`;
}
