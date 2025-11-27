import { useState } from 'react';
import NodeModal from './NodeModal';

interface Node {
    id: string;
    type: 'compute' | 'pageserver' | 'safekeeper' | 'storage_broker';
    name: string;
    x: number;
    y: number;
}

interface Connection {
    from: string;
    to: string;
}

interface NodeCanvasProps {
    onNodeSelect: (nodeId: string | null) => void;
    onDeploy: (nodeId: string) => void;
}

export default function NodeCanvas({ onNodeSelect, onDeploy }: NodeCanvasProps) {
    const [nodes, setNodes] = useState<Node[]>([]);
    const [connections, setConnections] = useState<Connection[]>([]);
    const [selectedNode, setSelectedNode] = useState<string | null>(null);
    const [showModal, setShowModal] = useState(false);
    const [modalNode, setModalNode] = useState<Node | null>(null);

    const handleAddNode = (type: Node['type']) => {
        const newNode: Node = {
            id: `node-${Date.now()}`,
            type,
            name: `${type}-${nodes.length + 1}`,
            x: 100 + nodes.length * 50,
            y: 100 + nodes.length * 50,
        };
        setNodes([...nodes, newNode]);
    };

    const handleNodeClick = (node: Node) => {
        setSelectedNode(node.id);
        setModalNode(node);
        setShowModal(true);
        onNodeSelect(node.id);
    };

    const handleDeploy = () => {
        if (nodes.length > 0) {
            onDeploy('all-nodes');
        }
    };

    return (
        <div className="relative h-full w-full bg-gray-100">
            {/* Toolbar */}
            <div className="absolute left-4 top-4 z-10 flex gap-2 rounded-lg bg-white p-2 shadow-lg">
                <button
                    onClick={() => handleAddNode('storage_broker')}
                    className="rounded bg-blue-500 px-3 py-2 text-xs font-medium text-white hover:bg-blue-600"
                >
                    + Storage Broker
                </button>
                <button
                    onClick={() => handleAddNode('pageserver')}
                    className="rounded bg-green-500 px-3 py-2 text-xs font-medium text-white hover:bg-green-600"
                >
                    + Pageserver
                </button>
                <button
                    onClick={() => handleAddNode('safekeeper')}
                    className="rounded bg-yellow-500 px-3 py-2 text-xs font-medium text-white hover:bg-yellow-600"
                >
                    + Safekeeper
                </button>
                <button
                    onClick={() => handleAddNode('compute')}
                    className="rounded bg-purple-500 px-3 py-2 text-xs font-medium text-white hover:bg-purple-600"
                >
                    + Compute
                </button>
                <div className="mx-2 border-l"></div>
                <button
                    onClick={handleDeploy}
                    disabled={nodes.length === 0}
                    className="rounded bg-indigo-600 px-4 py-2 text-xs font-medium text-white hover:bg-indigo-700 disabled:opacity-50"
                >
                    Deploy All
                </button>
            </div>

            {/* Canvas */}
            <svg className="h-full w-full">
                {/* Connections */}
                {connections.map((conn, idx) => {
                    const fromNode = nodes.find((n) => n.id === conn.from);
                    const toNode = nodes.find((n) => n.id === conn.to);
                    if (!fromNode || !toNode) return null;

                    return (
                        <line
                            key={idx}
                            x1={fromNode.x + 60}
                            y1={fromNode.y + 30}
                            x2={toNode.x + 60}
                            y2={toNode.y + 30}
                            stroke="#94a3b8"
                            strokeWidth="2"
                        />
                    );
                })}

                {/* Nodes */}
                {nodes.map((node) => (
                    <g
                        key={node.id}
                        transform={`translate(${node.x}, ${node.y})`}
                        onClick={() => handleNodeClick(node)}
                        className="cursor-pointer"
                    >
                        <rect
                            width="120"
                            height="60"
                            rx="8"
                            fill={
                                node.type === 'storage_broker'
                                    ? '#3b82f6'
                                    : node.type === 'pageserver'
                                        ? '#10b981'
                                        : node.type === 'safekeeper'
                                            ? '#f59e0b'
                                            : '#a855f7'
                            }
                            stroke={selectedNode === node.id ? '#1e293b' : 'transparent'}
                            strokeWidth="3"
                        />
                        <text
                            x="60"
                            y="30"
                            textAnchor="middle"
                            fill="white"
                            fontSize="12"
                            fontWeight="600"
                        >
                            {node.name}
                        </text>
                        <text
                            x="60"
                            y="45"
                            textAnchor="middle"
                            fill="white"
                            fontSize="10"
                        >
                            {node.type}
                        </text>
                    </g>
                ))}
            </svg>

            {/* Node Modal */}
            {showModal && modalNode && (
                <NodeModal
                    node={modalNode}
                    onClose={() => {
                        setShowModal(false);
                        setModalNode(null);
                    }}
                    onSave={(updatedNode: Node) => {
                        setNodes(nodes.map((n) => (n.id === updatedNode.id ? updatedNode : n)));
                        setShowModal(false);
                    }}
                />
            )}
        </div>
    );
}
