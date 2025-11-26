import { useState, useEffect } from 'react';
import { Button } from "~/components/ui/button";
import { ScrollArea } from "~/components/ui/scroll-area";
import { Loader2, Terminal } from "lucide-react";
import { toast } from "sonner";

// Mock API for now, replace with actual gRPC client later
const deployGraph = async (graphId: string) => {
    // Simulate API call
    return new Promise((resolve) => setTimeout(resolve, 1000));
};

const getDeploymentLogs = async (graphId: string) => {
    // Simulate API call
    return ["Deploying Storage Broker...", "Storage Broker started.", "Deploying Safekeeper...", "Safekeeper started."];
};

export function DeploymentManager({ graphId }: { graphId: string }) {
    const [isDeploying, setIsDeploying] = useState(false);
    const [logs, setLogs] = useState<string[]>([]);
    const [isOpen, setIsOpen] = useState(false);

    const handleDeploy = async () => {
        setIsDeploying(true);
        setIsOpen(true);
        setLogs(["Starting deployment..."]);

        try {
            await deployGraph(graphId);
            toast.success("Deployment initiated");

            // Start polling logs
            const interval = setInterval(async () => {
                const newLogs = await getDeploymentLogs(graphId);
                setLogs(prev => [...prev, ...newLogs]);
                // In real app, stop interval when deployment finishes
            }, 2000);

            // Stop polling after 10s for demo
            setTimeout(() => {
                clearInterval(interval);
                setIsDeploying(false);
                setLogs(prev => [...prev, "Deployment completed successfully."]);
            }, 10000);

        } catch (error) {
            toast.error("Deployment failed");
            setIsDeploying(false);
        }
    };

    return (
        <div className="fixed bottom-4 right-4 z-50 flex flex-col gap-2 items-end">
            {isOpen && (
                <div className="w-96 h-64 bg-black/90 text-green-400 p-4 rounded-md border border-green-500/30 shadow-2xl font-mono text-xs flex flex-col">
                    <div className="flex justify-between items-center mb-2 border-b border-green-500/30 pb-1">
                        <span className="flex items-center gap-2"><Terminal className="w-3 h-3" /> Deployment Logs</span>
                        <Button variant="ghost" size="sm" className="h-4 w-4 p-0 text-green-400 hover:text-green-300" onClick={() => setIsOpen(false)}>x</Button>
                    </div>
                    <ScrollArea className="flex-1">
                        {logs.map((log, i) => (
                            <div key={i}>{log}</div>
                        ))}
                        {isDeploying && <Loader2 className="w-3 h-3 animate-spin mt-2" />}
                    </ScrollArea>
                </div>
            )}
            <Button onClick={handleDeploy} disabled={isDeploying} className="shadow-lg">
                {isDeploying ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : "Deploy Changes"}
            </Button>
        </div>
    );
}
