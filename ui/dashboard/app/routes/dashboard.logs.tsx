import { useEffect, useState, useRef } from "react";
import { toast } from "sonner";
import { LogFile } from "~/proto/controlplane/logs";
import { logsService } from "~/lib/api";
import { Card, CardContent, CardHeader, CardTitle } from "~/components/ui/card";
import { Button } from "~/components/ui/button";
import { ChevronLeft, ChevronRight, FileText, Loader2, RefreshCw } from "lucide-react";
import { Separator } from "~/components/ui/separator";

export default function LogsPage() {
  const [logs, setLogs] = useState<LogFile[]>([]);
  const [selectedLog, setSelectedLog] = useState<string | null>(null);
  const [logContent, setLogContent] = useState<string[]>([]);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [loading, setLoading] = useState(false);
  const [loadingContent, setLoadingContent] = useState(false);
  const logContainerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    fetchLogs();
  }, []);

  useEffect(() => {
    if (selectedLog) {
      fetchLogContent(1000000);
    } else {
      setLogContent([]);
    }
  }, [selectedLog]);

  const fetchLogs = async () => {
    setLoading(true);
    try {
      const response = await logsService.GetSystemLogs({});
      setLogs(response.files);
      toast.success("Logs refreshed");
    } catch (error) {
      console.error("Failed to fetch logs:", error);
      toast.error("Failed to fetch logs");
    } finally {
      setLoading(false);
    }
  };

  const fetchLogContent = async (pageNum: number) => {
    if (!selectedLog) return;
    setLoadingContent(true);
    try {
      const response = await logsService.ReadSystemLog({
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

  useEffect(() => {
    if (logContainerRef.current) {
      logContainerRef.current.scrollTop = logContainerRef.current.scrollHeight;
    }
  }, [logContent]);

  return (
    <div className="flex h-full flex-col space-y-4 p-8 md:flex">
      <div className="flex items-center justify-between space-y-2">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">System Logs</h2>
          <p className="text-muted-foreground">
            View and analyze system and server logs.
          </p>
        </div>
        <div className="flex items-center space-x-2">
          <Button onClick={fetchLogs} isLoading={loading}>
            {!loading && <RefreshCw className="mr-2 h-4 w-4" />}
            Refresh
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-12 gap-4 h-[calc(100vh-200px)]">
        <Card className="col-span-4 h-full flex flex-col">
          <CardHeader>
            <CardTitle>Log Files</CardTitle>
          </CardHeader>
          <CardContent className="flex-1 overflow-hidden p-0">
            <div className="h-full overflow-auto">
              <div className="flex flex-col gap-1 p-4 pt-0">
                {logs.map((log) => (
                  <Button
                    key={log.name}
                    variant={selectedLog === log.name ? "secondary" : "ghost"}
                    className="justify-start h-auto py-3 px-4"
                    onClick={() => setSelectedLog(log.name)}
                  >
                    <FileText className="mr-2 h-4 w-4 shrink-0" />
                    <div className="flex flex-col items-start overflow-hidden">
                      <span className="truncate w-full text-sm font-medium">{log.name}</span>
                      <span className="text-xs text-muted-foreground">
                        {formatBytes(log.size)} • {new Date(log.lastModified).toLocaleString()}
                      </span>
                    </div>
                  </Button>
                ))}
                {logs.length === 0 && !loading && (
                  <div className="text-center text-muted-foreground py-8">
                    No logs found
                  </div>
                )}
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="col-span-8 h-full flex flex-col">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-base font-medium">
              {selectedLog ? selectedLog : "Select a log file"}
            </CardTitle>
            {selectedLog && (
              <div className="flex items-center gap-2">
                <span className="text-xs text-muted-foreground">
                  Page {page} of {totalPages}
                </span>
                <div className="flex gap-1">
                  <Button
                    variant="outline"
                    size="icon"
                    className="h-8 w-8"
                    onClick={() => handlePageChange(page - 1)}
                    disabled={page <= 1 || loadingContent}
                  >
                    <ChevronLeft className="h-4 w-4" />
                  </Button>
                  <Button
                    variant="outline"
                    size="icon"
                    className="h-8 w-8"
                    onClick={() => handlePageChange(page + 1)}
                    disabled={page >= totalPages || loadingContent}
                  >
                    <ChevronRight className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            )}
          </CardHeader>
          <Separator />
          <CardContent className="flex-1 overflow-hidden p-0 font-mono text-xs">
            <div className="h-full overflow-auto p-4" ref={logContainerRef}>
              {loadingContent ? (
                <div className="flex items-center justify-center h-full">
                  <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
                </div>
              ) : selectedLog ? (
                logContent.length > 0 ? (
                  logContent.map((line, index) => (
                    <div key={index} className="whitespace-pre-wrap border-b border-border/50 py-0.5 last:border-0">
                      {line}
                    </div>
                  ))
                ) : (
                  <div className="text-center text-muted-foreground py-8">
                    Empty log file
                  </div>
                )
              ) : (
                <div className="flex items-center justify-center h-full text-muted-foreground">
                  Select a log file to view content
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
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
