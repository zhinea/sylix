import { useEffect, useState } from "react";
import { Activity, CheckCircle, XCircle } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "~/components/ui/dialog";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "~/components/ui/tabs";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "~/components/ui/table";
import { Card, CardContent, CardHeader, CardTitle } from "~/components/ui/card";
import { serverService } from "~/lib/api";
import { ServerStat, ServerAccident } from "~/proto/controlplane/server";

interface ServerStatsDialogProps {
  serverId: string | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function ServerStatsDialog({
  serverId,
  open,
  onOpenChange,
}: ServerStatsDialogProps) {
  const [stats, setStats] = useState<ServerStat[]>([]);
  const [accidents, setAccidents] = useState<ServerAccident[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (open && serverId) {
      fetchData(serverId);
    }
  }, [open, serverId]);

  const fetchData = async (id: string) => {
    setLoading(true);
    try {
      const [statsRes, accidentsRes] = await Promise.all([
        serverService.GetStats({ serverId: id }),
        serverService.GetAccidents({ serverId: id }),
      ]);
      setStats(statsRes.stats);
      setAccidents(accidentsRes.accidents);
    } catch (error) {
      console.error("Failed to fetch stats:", error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Server Analytics</DialogTitle>
          <DialogDescription>
            Performance statistics and incident history.
          </DialogDescription>
        </DialogHeader>

        <Tabs defaultValue="stats" className="w-full">
          <TabsList>
            <TabsTrigger value="stats">Performance Stats</TabsTrigger>
            <TabsTrigger value="accidents">Incidents</TabsTrigger>
          </TabsList>
          
          <TabsContent value="stats" className="space-y-4">
            {stats.length > 0 && (
              <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">
                      Avg Response Time
                    </CardTitle>
                    <Activity className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">
                      {stats[0].averageResponseTime.toFixed(2)}ms
                    </div>
                  </CardContent>
                </Card>
                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">
                      Success Rate
                    </CardTitle>
                    <CheckCircle className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">
                      {stats[0].successRate.toFixed(1)}%
                    </div>
                  </CardContent>
                </Card>
              </div>
            )}

            <div className="rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Time</TableHead>
                    <TableHead>Avg Response</TableHead>
                    <TableHead>Min / Max</TableHead>
                    <TableHead>Success Rate</TableHead>
                    <TableHead>Pings</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {stats.map((stat) => (
                    <TableRow key={stat.id}>
                      <TableCell>{new Date(stat.timestamp).toLocaleString()}</TableCell>
                      <TableCell>{stat.averageResponseTime.toFixed(2)}ms</TableCell>
                      <TableCell>{stat.minResponseTime}ms / {stat.maxResponseTime}ms</TableCell>
                      <TableCell>{stat.successRate.toFixed(1)}%</TableCell>
                      <TableCell>{stat.pingCount}</TableCell>
                    </TableRow>
                  ))}
                  {stats.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={5} className="text-center">
                        No stats available
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </div>
          </TabsContent>

          <TabsContent value="accidents">
            <div className="rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Time</TableHead>
                    <TableHead>Error</TableHead>
                    <TableHead>Response Time</TableHead>
                    <TableHead>Details</TableHead>
                    <TableHead>Status</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {accidents.map((accident) => (
                    <TableRow key={accident.id}>
                      <TableCell>{new Date(accident.createdAt).toLocaleString()}</TableCell>
                      <TableCell className="font-medium text-red-500">{accident.error}</TableCell>
                      <TableCell>{accident.responseTime > 0 ? `${accident.responseTime}ms` : '-'}</TableCell>
                      <TableCell>{accident.details}</TableCell>
                      <TableCell>
                        {accident.resolved ? (
                          <span className="flex items-center text-green-500">
                            <CheckCircle className="mr-1 h-4 w-4" /> Resolved
                          </span>
                        ) : (
                          <span className="flex items-center text-red-500">
                            <XCircle className="mr-1 h-4 w-4" /> Open
                          </span>
                        )}
                      </TableCell>
                    </TableRow>
                  ))}
                  {accidents.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={5} className="text-center">
                        No incidents recorded
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </div>
          </TabsContent>
        </Tabs>
      </DialogContent>
    </Dialog>
  );
}
