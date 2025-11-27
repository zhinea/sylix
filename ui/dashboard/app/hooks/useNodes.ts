import { useState, useEffect } from 'react';

export interface Node {
    id: string;
    name: string;
    description: string;
    type: 'compute' | 'pageserver' | 'safekeeper' | 'storage_broker';
    priorityStartup: number;
    fields: string;
    imports: string;
    exports: string;
    serverId: string;
    createdAt: string;
    updatedAt: string;
}

// Simplified hook - will connect to gRPC-Web later
// For now, provides the interface for the UI components
export function useNodes() {
    const [nodes, setNodes] = useState<Node[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const fetchNodes = async () => {
        setLoading(true);
        setError(null);
        try {
            // TODO: Implement actual gRPC-Web call
            // const response = await fetch('http://localhost:8082/...')
            setNodes([]);
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to fetch nodes');
        } finally {
            setLoading(false);
        }
    };

    const createNode = async (data: {
        name: string;
        description: string;
        type: string;
        priorityStartup: number;
        fields: string;
        imports: string;
        exports: string;
        serverId: string;
    }) => {
        setLoading(true);
        setError(null);
        try {
            // TODO: Implement actual gRPC-Web call
            const newNode = {
                ...data,
                id: `node-${Date.now()}`,
                createdAt: new Date().toISOString(),
                updatedAt: new Date().toISOString(),
            } as Node;
            setNodes([...nodes, newNode]);
            return newNode;
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to create node');
            throw err;
        } finally {
            setLoading(false);
        }
    };

    const updateNode = async (data: {
        id: string;
        name: string;
        description: string;
        type: string;
        priorityStartup: number;
        fields: string;
        imports: string;
        exports: string;
        serverId: string;
    }) => {
        setLoading(true);
        setError(null);
        try {
            // TODO: Implement actual gRPC-Web call
            const updatedNode = {
                ...data,
                createdAt: nodes.find((n) => n.id === data.id)?.createdAt || '',
                updatedAt: new Date().toISOString(),
            } as Node;
            setNodes(nodes.map((n) => (n.id === data.id ? updatedNode : n)));
            return updatedNode;
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to update node');
            throw err;
        } finally {
            setLoading(false);
        }
    };

    const deleteNode = async (id: string) => {
        setLoading(true);
        setError(null);
        try {
            // TODO: Implement actual gRPC-Web call
            setNodes(nodes.filter((n) => n.id !== id));
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to delete node');
            throw err;
        } finally {
            setLoading(false);
        }
    };

    const deployNode = async (id: string) => {
        setLoading(true);
        setError(null);
        try {
            // TODO: Implement actual gRPC-Web call
            console.log('Deploying node:', id);
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to deploy node');
            throw err;
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchNodes();
    }, []);

    return {
        nodes,
        loading,
        error,
        fetchNodes,
        createNode,
        updateNode,
        deleteNode,
        deployNode,
    };
}
