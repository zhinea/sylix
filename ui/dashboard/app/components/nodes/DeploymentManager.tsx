import { useEffect, useRef } from 'react';

interface DeploymentManagerProps {
    logs: string[];
    onClose: () => void;
}

export default function DeploymentManager({
    logs,
    onClose,
}: DeploymentManagerProps) {
    const logsEndRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        logsEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    }, [logs]);

    return (
        <div className="flex h-full flex-col">
            <div className="flex items-center justify-between border-b bg-gray-50 px-4 py-3">
                <h3 className="font-semibold text-gray-900">Deployment Logs</h3>
                <button
                    onClick={onClose}
                    className="text-gray-500 hover:text-gray-700"
                >
                    <svg
                        className="h-5 w-5"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                    >
                        <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M6 18L18 6M6 6l12 12"
                        />
                    </svg>
                </button>
            </div>

            <div className="flex-1 overflow-y-auto bg-gray-900 p-4 font-mono text-sm">
                {logs.map((log, index) => (
                    <div key={index} className="text-green-400">
                        <span className="text-gray-500">[{new Date().toLocaleTimeString()}]</span>{' '}
                        {log}
                    </div>
                ))}
                <div ref={logsEndRef} />
            </div>

            <div className="border-t bg-gray-50 px-4 py-3">
                <div className="flex items-center gap-2">
                    <div className="h-2 w-2 animate-pulse rounded-full bg-green-500"></div>
                    <span className="text-sm text-gray-600">Deployment in progress...</span>
                </div>
            </div>
        </div>
    );
}
