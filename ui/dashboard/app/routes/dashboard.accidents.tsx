import { useEffect, useState } from "react";
import { useLoaderData } from "react-router";
import { format } from "date-fns";
import { Calendar as CalendarIcon, CheckCircle2, XCircle, Trash2 } from "lucide-react";

import { Button } from "~/components/ui/button";
import { Calendar } from "~/components/ui/calendar";
import { Checkbox } from "~/components/ui/checkbox";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "~/components/ui/popover";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "~/components/ui/table";
import { Badge } from "~/components/ui/badge";
import { serverService } from "~/lib/api";
import { ServerAccident } from "~/proto/controlplane/server";
import { cn } from "~/lib/utils";
import { IncidentDetailsModal } from "~/components/servers/incident-details-modal";

export async function clientLoader() {
  try {
    const response = await serverService.All({});
    return { servers: response.servers };
  } catch (error) {
    console.error("Failed to fetch servers:", error);
    return { servers: [] };
  }
}

export default function ServerAccidentsPage() {
  const { servers } = useLoaderData<typeof clientLoader>();
  const [accidents, setAccidents] = useState<ServerAccident[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [selectedAccident, setSelectedAccident] = useState<ServerAccident | null>(null);
  const [selectedAccidents, setSelectedAccidents] = useState<string[]>([]);
  
  // Filters
  const [selectedServerId, setSelectedServerId] = useState<string>("all");
  const [date, setDate] = useState<Date | undefined>(undefined);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  useEffect(() => {
    const fetchAccidents = async () => {
      try {
        const response = await serverService.GetAccidents({
          serverId: selectedServerId === "all" ? "" : selectedServerId,
          startDate: date ? date.toISOString() : "",
          endDate: date ? new Date(date.getTime() + 86400000).toISOString() : "", // Next day
          page,
          pageSize,
          resolved: false, // TODO: Add resolved filter UI
        });
        setAccidents(response.accidents);
        setTotalCount(Number(response.totalCount));
      } catch (error) {
        console.error("Failed to fetch accidents:", error);
      }
    };

    fetchAccidents();
  }, [selectedServerId, date, page, pageSize]);

  const handleDelete = async (id: string) => {
    if (!confirm("Are you sure you want to delete this incident?")) return;
    try {
      await serverService.DeleteAccident({ id });
      setAccidents(accidents.filter((a) => a.id !== id));
      setSelectedAccidents(selectedAccidents.filter((sid) => sid !== id));
      setTotalCount(prev => prev - 1);
    } catch (error) {
      console.error("Failed to delete accident:", error);
    }
  };

  const handleBatchDelete = async () => {
    if (!confirm(`Are you sure you want to delete ${selectedAccidents.length} incidents?`)) return;
    try {
      await serverService.BatchDeleteAccidents({ ids: selectedAccidents });
      setAccidents(accidents.filter((a) => !selectedAccidents.includes(a.id)));
      setTotalCount(prev => prev - selectedAccidents.length);
      setSelectedAccidents([]);
    } catch (error) {
      console.error("Failed to batch delete accidents:", error);
    }
  };

  const toggleSelectAll = () => {
    if (selectedAccidents.length === accidents.length && accidents.length > 0) {
      setSelectedAccidents([]);
    } else {
      setSelectedAccidents(accidents.map((a) => a.id));
    }
  };

  const toggleSelect = (id: string) => {
    if (selectedAccidents.includes(id)) {
      setSelectedAccidents(selectedAccidents.filter((sid) => sid !== id));
    } else {
      setSelectedAccidents([...selectedAccidents, id]);
    }
  };

  const totalPages = Math.ceil(totalCount / pageSize);

  return (
    <div className="flex flex-col gap-4 p-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold tracking-tight">Server Incidents</h1>
        {selectedAccidents.length > 0 && (
          <Button 
            variant="destructive" 
            size="sm" 
            onClick={handleBatchDelete}
          >
            <Trash2 className="mr-2 h-4 w-4" />
            Delete Selected ({selectedAccidents.length})
          </Button>
        )}
      </div>

      <div className="flex items-center gap-4 p-4 border rounded-lg bg-card">
        <div className="w-[200px]">
          <Select value={selectedServerId} onValueChange={setSelectedServerId}>
            <SelectTrigger>
              <SelectValue placeholder="Filter by Server" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Servers</SelectItem>
              {servers.map((s) => (
                <SelectItem key={s.id} value={s.id}>
                  {s.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <Popover>
          <PopoverTrigger asChild>
            <Button
              variant={"outline"}
              className={cn(
                "w-[240px] justify-start text-left font-normal",
                !date && "text-muted-foreground"
              )}
            >
              <CalendarIcon className="mr-2 h-4 w-4" />
              {date ? format(date, "PPP") : <span>Pick a date</span>}
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-auto p-0" align="start">
            <Calendar
              mode="single"
              selected={date}
              onSelect={setDate}
              initialFocus
            />
          </PopoverContent>
        </Popover>

        <Button 
          variant="ghost" 
          onClick={() => {
            setSelectedServerId("all");
            setDate(undefined);
            setPage(1);
          }}
        >
          Reset Filters
        </Button>
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[50px]">
                <Checkbox 
                  checked={selectedAccidents.length === accidents.length && accidents.length > 0}
                  onCheckedChange={toggleSelectAll}
                />
              </TableHead>
              <TableHead>Time</TableHead>
              <TableHead>Server</TableHead>
              <TableHead>Error</TableHead>
              <TableHead>Response Time</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Details</TableHead>
              <TableHead className="w-[50px]"></TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {accidents.length === 0 ? (
              <TableRow>
                <TableCell colSpan={8} className="text-center h-24 text-muted-foreground">
                  No incidents found.
                </TableCell>
              </TableRow>
            ) : (
              accidents.map((accident) => {
                const server = servers.find(s => s.id === accident.serverId);
                return (
                  <TableRow 
                    key={accident.id} 
                    className="cursor-pointer hover:bg-muted/50"
                    onClick={() => setSelectedAccident(accident)}
                  >
                    <TableCell onClick={(e) => e.stopPropagation()}>
                      <Checkbox 
                        checked={selectedAccidents.includes(accident.id)}
                        onCheckedChange={() => toggleSelect(accident.id)}
                      />
                    </TableCell>
                    <TableCell>{new Date(accident.createdAt).toLocaleString()}</TableCell>
                    <TableCell className="font-medium">
                      {server ? server.name : accident.serverId}
                    </TableCell>
                    <TableCell className="text-red-500">{accident.error}</TableCell>
                    <TableCell>{accident.responseTime > 0 ? `${accident.responseTime}ms` : '-'}</TableCell>
                    <TableCell>
                      {accident.resolved ? (
                        <Badge variant="outline" className="border-green-500 text-green-500">
                          <CheckCircle2 className="w-3 h-3 mr-1" /> Resolved
                        </Badge>
                      ) : (
                        <Badge variant="outline" className="border-red-500 text-red-500">
                          <XCircle className="w-3 h-3 mr-1" /> Unresolved
                        </Badge>
                      )}
                    </TableCell>
                    <TableCell className="max-w-[300px] truncate" title={accident.details}>
                      {accident.details}
                    </TableCell>
                    <TableCell onClick={(e) => e.stopPropagation()}>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 text-red-500 hover:text-red-600 hover:bg-red-100"
                        onClick={(e) => {
                          e.stopPropagation();
                          handleDelete(accident.id);
                        }}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </TableCell>
                  </TableRow>
                );
              })
            )}
          </TableBody>
        </Table>
      </div>

      <div className="flex items-center justify-between px-2">
        <div className="flex items-center space-x-2">
          <p className="text-sm text-muted-foreground">Rows per page</p>
          <Select
            value={pageSize.toString()}
            onValueChange={(value) => {
              setPageSize(Number(value));
              setPage(1);
            }}
          >
            <SelectTrigger className="h-8 w-[70px]">
              <SelectValue placeholder={pageSize.toString()} />
            </SelectTrigger>
            <SelectContent side="top">
              {[10, 20, 50, 100, 1000].map((pageSize) => (
                <SelectItem key={pageSize} value={`${pageSize}`}>
                  {pageSize}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div className="flex items-center space-x-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setPage(p => Math.max(1, p - 1))}
            disabled={page === 1}
          >
            Previous
          </Button>
          <div className="text-sm text-muted-foreground">
            Page {page} of {totalPages || 1}
          </div>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setPage(p => Math.min(totalPages, p + 1))}
            disabled={page >= totalPages}
          >
            Next
          </Button>
        </div>
      </div>

      <IncidentDetailsModal
        accident={selectedAccident}
        open={!!selectedAccident}
        onOpenChange={(open) => !open && setSelectedAccident(null)}
        serverName={selectedAccident ? servers.find(s => s.id === selectedAccident.serverId)?.name : undefined}
      />
    </div>
  );
}
