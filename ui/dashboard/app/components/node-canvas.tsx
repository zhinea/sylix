import { Handle, Position } from '@xyflow/react';
import { Card, CardContent, CardHeader, CardTitle } from "~/components/ui/card";
import { Badge } from "~/components/ui/badge";
import { Server, Database, HardDrive, Shield, Network } from "lucide-react";

// Custom Node Components

// Local NodeProps type compatible with xyflow/react props (we only use data here)
export type LocalNodeProps<T = any> = { data: T };

export function ComputeNode({ data }: LocalNodeProps<{ pg_version?: string; pg_port?: string; expose_internet?: boolean }>) {
    return (
        <Card className="w-64 border-2 border-primary/50 shadow-lg bg-card text-card-foreground">
            <CardHeader className="p-3 pb-0">
                <CardTitle className="text-sm font-bold flex items-center gap-2">
                    <Database className="w-4 h-4" />
                    Compute Engine
                </CardTitle>
            </CardHeader>
            <CardContent className="p-3 text-xs">
                <div className="flex flex-col gap-1">
                    <div className="flex justify-between">
                        <span className="text-muted-foreground">PG Version:</span>
                        <span>{data.pg_version as string}</span>
                    </div>
                    <div className="flex justify-between">
                        <span className="text-muted-foreground">Port:</span>
                        <span>{data.pg_port as string}</span>
                    </div>
                    {data.expose_internet && <Badge variant="outline" className="mt-1 w-fit">Exposed</Badge>}
                </div>
            </CardContent>

            {/* Inputs */}
            <Handle type="target" position={Position.Left} id="safekeeper" style={{ top: '30%' }} />
            <Handle type="target" position={Position.Left} id="pageserver" style={{ top: '50%' }} />
            <Handle type="target" position={Position.Left} id="storage_broker" style={{ top: '70%' }} />

            {/* Outputs */}
            <Handle type="source" position={Position.Right} id="clients" />
        </Card>
    );
}

export function PageserverNode({ data }: LocalNodeProps<{ server_id?: string }>) {
    return (
        <Card className="w-64 border-2 border-blue-500/50 shadow-lg bg-card">
            <CardHeader className="p-3 pb-0">
                <CardTitle className="text-sm font-bold flex items-center gap-2">
                    <HardDrive className="w-4 h-4" />
                    Pageserver
                </CardTitle>
            </CardHeader>
            <CardContent className="p-3 text-xs">
                <div className="flex flex-col gap-1">
                    <div className="flex justify-between">
                        <span className="text-muted-foreground">Server ID:</span>
                        <span>{data.server_id as string}</span>
                    </div>
                </div>
            </CardContent>
            <Handle type="target" position={Position.Left} id="storage_broker" />
            <Handle type="source" position={Position.Right} id="compute" />
        </Card>
    );
}

export function SafekeeperNode({ data }: LocalNodeProps<{ server_id?: string }>) {
    return (
        <Card className="w-64 border-2 border-green-500/50 shadow-lg bg-card">
            <CardHeader className="p-3 pb-0">
                <CardTitle className="text-sm font-bold flex items-center gap-2">
                    <Shield className="w-4 h-4" />
                    Safekeeper
                </CardTitle>
            </CardHeader>
            <CardContent className="p-3 text-xs">
                <div className="flex flex-col gap-1">
                    <div className="flex justify-between">
                        <span className="text-muted-foreground">Server ID:</span>
                        <span>{data.server_id as string}</span>
                    </div>
                </div>
            </CardContent>
            <Handle type="target" position={Position.Left} id="storage_broker" />
            <Handle type="source" position={Position.Right} id="compute" style={{ top: '30%' }} />
            <Handle type="source" position={Position.Right} id="pageserver" style={{ top: '70%' }} />
        </Card>
    );
}

export function StorageBrokerNode({ data }: LocalNodeProps<{ server_id?: string }>) {
    return (
        <Card className="w-64 border-2 border-orange-500/50 shadow-lg bg-card">
            <CardHeader className="p-3 pb-0">
                <CardTitle className="text-sm font-bold flex items-center gap-2">
                    <Network className="w-4 h-4" />
                    Storage Broker
                </CardTitle>
            </CardHeader>
            <CardContent className="p-3 text-xs">
                <div className="flex flex-col gap-1">
                    <div className="flex justify-between">
                        <span className="text-muted-foreground">Server ID:</span>
                        <span>{data.server_id as string}</span>
                    </div>
                </div>
            </CardContent>
            <Handle type="source" position={Position.Right} id="safekeeper" style={{ top: '30%' }} />
            <Handle type="source" position={Position.Right} id="pageserver" style={{ top: '50%' }} />
            <Handle type="source" position={Position.Right} id="compute" style={{ top: '70%' }} />
        </Card>
    );
}

export const nodeTypes = {
    compute: ComputeNode,
    pageserver: PageserverNode,
    safekeeper: SafekeeperNode,
    storage_broker: StorageBrokerNode,
};
