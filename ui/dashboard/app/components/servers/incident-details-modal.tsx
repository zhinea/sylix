import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "~/components/ui/dialog";
import { ServerAccident } from "~/proto/controlplane/server";
import { Badge } from "~/components/ui/badge";
import { CheckCircle2, XCircle } from "lucide-react";

interface IncidentDetailsModalProps {
  accident: ServerAccident | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  serverName?: string;
}

export function IncidentDetailsModal({
  accident,
  open,
  onOpenChange,
  serverName,
}: IncidentDetailsModalProps) {
  if (!accident) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Incident Details</DialogTitle>
          <DialogDescription>
            Detailed information about the server incident.
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-4 py-4">
          <div className="grid grid-cols-4 items-center gap-4">
            <span className="font-medium">Server:</span>
            <span className="col-span-3">{serverName || accident.serverId}</span>
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <span className="font-medium">Time:</span>
            <span className="col-span-3">
              {new Date(accident.createdAt).toLocaleString()}
            </span>
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <span className="font-medium">Status:</span>
            <div className="col-span-3">
              {accident.resolved ? (
                <Badge variant="outline" className="border-green-500 text-green-500">
                  <CheckCircle2 className="w-3 h-3 mr-1" /> Resolved
                </Badge>
              ) : (
                <Badge variant="outline" className="border-red-500 text-red-500">
                  <XCircle className="w-3 h-3 mr-1" /> Unresolved
                </Badge>
              )}
            </div>
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <span className="font-medium">Response Time:</span>
            <span className="col-span-3">
              {accident.responseTime > 0 ? `${accident.responseTime}ms` : "-"}
            </span>
          </div>
          <div className="grid grid-cols-4 items-start gap-4">
            <span className="font-medium pt-1">Error:</span>
            <span className="col-span-3 text-red-500 font-mono text-sm bg-red-50 p-2 rounded border border-red-100">
              {accident.error}
            </span>
          </div>
          <div className="grid grid-cols-4 items-start gap-4">
            <span className="font-medium pt-1">Details:</span>
            <div className="col-span-3 border rounded-md bg-muted/50 overflow-auto max-h-[200px] p-4">
              <pre className="text-sm font-mono whitespace-pre-wrap break-words">
                {accident.details}
              </pre>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
