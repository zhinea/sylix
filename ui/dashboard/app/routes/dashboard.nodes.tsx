import { useCallback, useState } from 'react';
import {
    ReactFlow,
    Background,
    Controls,
    MiniMap,
    useNodesState,
    useEdgesState,
    addEdge,
    type Connection,
    type Edge,
    type Node
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { nodeTypes } from '~/components/node-canvas';
import { Button } from '~/components/ui/button';
import { Plus } from 'lucide-react';
import { DeploymentManager } from '~/components/deployment-manager';

const initialNodes: Node[] = [];
const initialEdges: Edge[] = [];

export default function DashboardNodes() {
    const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
    const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);

    const onConnect = useCallback(
        (params: Connection) => setEdges((eds) => addEdge(params, eds)),
        [setEdges],
    );

    const addNode = (type: string) => {
        const id = `${type}-${nodes.length + 1}`;
        const newNode: Node = {
            id,
            type,
            position: { x: Math.random() * 500, y: Math.random() * 500 },
            data: { label: type, server_id: 'server-1' }, // Default data
        };
        setNodes((nds) => nds.concat(newNode));
    };

    return (
        <div className="h-full flex flex-col">
            <div className="p-4 border-b flex justify-between items-center bg-background">
                <h1 className="text-2xl font-bold">Nodes Management</h1>
                <div className="flex gap-2">
                    <Button size="sm" onClick={() => addNode('compute')}>+ Compute</Button>
                    <Button size="sm" onClick={() => addNode('pageserver')}>+ Pageserver</Button>
                    <Button size="sm" onClick={() => addNode('safekeeper')}>+ Safekeeper</Button>
                    <Button size="sm" onClick={() => addNode('storage_broker')}>+ Broker</Button>
                </div>
            </div>
            <div className="flex-1 bg-muted/20">
                <ReactFlow
                    nodes={nodes}
                    edges={edges}
                    onNodesChange={onNodesChange}
                    onEdgesChange={onEdgesChange}
                    onConnect={onConnect}
                    nodeTypes={nodeTypes}
                    fitView
                >
                    <Background />
                    <Controls />
                    <MiniMap />
                </ReactFlow>
            </div>
            <DeploymentManager graphId="current-graph" />
        </div>
    );
}
