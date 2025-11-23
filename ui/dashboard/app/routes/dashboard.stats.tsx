import { useEffect, useState } from "react";
import { useLoaderData } from "react-router";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "~/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "~/components/ui/tabs";
import { serverService } from "~/lib/api";
import { ServerPing, ServerStat } from "~/proto/controlplane/server";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";

export async function clientLoader() {
  try {
    const response = await serverService.All({});
    return { servers: response.servers };
  } catch (error) {
    console.error("Failed to fetch servers:", error);
    return { servers: [] };
  }
}

export default function ServerStatsPage() {
  const { servers } = useLoaderData<typeof clientLoader>();
  const [selectedServerId, setSelectedServerId] = useState<string>(servers[0]?.id || "");
  const [realtimePings, setRealtimePings] = useState<ServerPing[]>([]);
  const [historicalStats, setHistoricalStats] = useState<ServerStat[]>([]);

  // Realtime polling
  useEffect(() => {
    if (!selectedServerId) return;

    const fetchRealtime = async () => {
      try {
        const response = await serverService.GetRealtimeStats({ serverId: selectedServerId, limit: 50 });
        setRealtimePings(response.pings.reverse());
      } catch (error) {
        console.error("Failed to fetch realtime stats:", error);
      }
    };

    fetchRealtime();
    const interval = setInterval(fetchRealtime, 5000);
    return () => clearInterval(interval);
  }, [selectedServerId]);

  // Historical fetch
  useEffect(() => {
    if (!selectedServerId) return;

    const fetchHistorical = async () => {
      try {
        const response = await serverService.GetStats({ serverId: selectedServerId });
        setHistoricalStats(response.stats.reverse());
      } catch (error) {
        console.error("Failed to fetch historical stats:", error);
      }
    };

    fetchHistorical();
  }, [selectedServerId]);

  if (servers.length === 0) {
    return <div className="p-8 text-center">No servers found.</div>;
  }

  return (
    <div className="flex flex-col gap-4 p-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold tracking-tight">Server Statistics</h1>
        <select
          className="border rounded p-2"
          value={selectedServerId}
          onChange={(e) => setSelectedServerId(e.target.value)}
        >
          {servers.map((s) => (
            <option key={s.id} value={s.id}>
              {s.name} ({s.ipAddress})
            </option>
          ))}
        </select>
      </div>

      <Tabs defaultValue="realtime" className="space-y-4">
        <TabsList>
          <TabsTrigger value="realtime">Realtime (Live)</TabsTrigger>
          <TabsTrigger value="historical">Historical (Aggregated)</TabsTrigger>
        </TabsList>

        <TabsContent value="realtime" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Live Response Time</CardTitle>
              <CardDescription>Last 50 pings (updated every 5s)</CardDescription>
            </CardHeader>
            <CardContent className="h-[400px]">
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={realtimePings}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis 
                    dataKey="createdAt" 
                    tickFormatter={(time) => new Date(time).toLocaleTimeString()} 
                  />
                  <YAxis />
                  <Tooltip 
                    labelFormatter={(label) => new Date(label).toLocaleString()}
                    formatter={(value: number) => [`${value}ms`, "Response Time"]}
                  />
                  <Line 
                    type="monotone" 
                    dataKey="responseTime" 
                    stroke="#2563eb" 
                    strokeWidth={2}
                    dot={false}
                  />
                </LineChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="historical" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Average Response Time</CardTitle>
              <CardDescription>15-minute aggregates</CardDescription>
            </CardHeader>
            <CardContent className="h-[400px]">
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={historicalStats}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis 
                    dataKey="timestamp" 
                    tickFormatter={(time) => new Date(time).toLocaleDateString()} 
                  />
                  <YAxis />
                  <Tooltip 
                    labelFormatter={(label) => new Date(label).toLocaleString()}
                    formatter={(value: number) => [`${value.toFixed(2)}ms`, "Avg Response Time"]}
                  />
                  <Line 
                    type="monotone" 
                    dataKey="averageResponseTime" 
                    stroke="#16a34a" 
                    strokeWidth={2} 
                  />
                </LineChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
