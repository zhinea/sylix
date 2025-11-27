import { useEffect, useState } from 'react';
import NodeCanvas from '../components/nodes/NodeCanvas';
import DeploymentManager from '../components/nodes/DeploymentManager';

export default function NodesPage() {
    const [selectedNode, setSelectedNode] = useState<string | null>(null);
    const [deploymentStatus, setDeploymentStatus] = useState<{
        isDeploying: boolean;
        logs: string[];
    }>({
        isDeploying: false,
        logs: [],
    });

    return (
        <div className="flex h-screen w-full flex-col bg-gray-50">
            <div className="border-b bg-white px-6 py-4">
                <h1 className="text-2xl font-bold text-gray-900">Node Management</h1>
                <p className="text-sm text-gray-600">
                    Create and manage your database nodes
                </p>
            </div>

            <div className="flex flex-1 overflow-hidden">
                {/* Canvas Area */}
                <div className="flex-1">
                    <NodeCanvas
                        onNodeSelect={setSelectedNode}
                        onDeploy={(nodeId) => {
                            setDeploymentStatus({
                                isDeploying: true,
                                logs: [`Starting deployment for node ${nodeId}...`],
                            });
                        }}
                    />
                </div>

                {/* Deployment Manager Sidebar */}
                {deploymentStatus.isDeploying && (
                    <div className="w-96 border-l bg-white">
                        <DeploymentManager
                            logs={deploymentStatus.logs}
                            onClose={() =>
                                setDeploymentStatus({ isDeploying: false, logs: [] })
                            }
                        />
                    </div>
                )}
            </div>
        </div>
    );
}
