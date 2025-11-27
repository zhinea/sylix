import { useState } from 'react';

interface Node {
    id: string;
    type: 'compute' | 'pageserver' | 'safekeeper' | 'storage_broker';
    name: string;
    x: number;
    y: number;
}

interface NodeModalProps {
    node: Node;
    onClose: () => void;
    onSave: (node: Node) => void;
}

export default function NodeModal({ node, onClose, onSave }: NodeModalProps) {
    const [formData, setFormData] = useState<{
        name: string;
        serverId: string;
        pgVersion?: string;
        pgPort?: string;
        exposeInternet?: boolean;
        backupStorageId?: string;
    }>({
        name: node.name,
        serverId: '',
        ...getNodeSpecificFields(node.type),
    });

    function getNodeSpecificFields(type: Node['type']) {
        switch (type) {
            case 'compute':
                return {
                    pgVersion: 'postgres-17',
                    pgPort: '5432',
                    exposeInternet: false,
                };
            case 'pageserver':
            case 'safekeeper':
                return {
                    backupStorageId: '',
                };
            default:
                return {};
        }
    }

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        onSave({
            ...node,
            name: formData.name,
        });
    };

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50">
            <div className="w-full max-w-md rounded-lg bg-white p-6 shadow-xl">
                <h2 className="mb-4 text-xl font-bold text-gray-900">
                    Configure {node.type}
                </h2>

                <form onSubmit={handleSubmit}>
                    <div className="space-y-4">
                        <div>
                            <label className="block text-sm font-medium text-gray-700">
                                Name
                            </label>
                            <input
                                type="text"
                                value={formData.name}
                                onChange={(e) =>
                                    setFormData({ ...formData, name: e.target.value })
                                }
                                className="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 focus:border-indigo-500 focus:outline-none"
                            />
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-gray-700">
                                Server ID
                            </label>
                            <select
                                value={formData.serverId}
                                onChange={(e) =>
                                    setFormData({ ...formData, serverId: e.target.value })
                                }
                                className="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 focus:border-indigo-500 focus:outline-none"
                            >
                                <option value="">Select a server</option>
                                <option value="server-1">Server 1</option>
                                <option value="server-2">Server 2</option>
                            </select>
                        </div>

                        {node.type === 'compute' && (
                            <>
                                <div>
                                    <label className="block text-sm font-medium text-gray-700">
                                        PostgreSQL Version
                                    </label>
                                    <select
                                        value={formData.pgVersion}
                                        onChange={(e) =>
                                            setFormData({ ...formData, pgVersion: e.target.value })
                                        }
                                        className="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 focus:border-indigo-500 focus:outline-none"
                                    >
                                        <option value="postgres-14">PostgreSQL 14</option>
                                        <option value="postgres-15">PostgreSQL 15</option>
                                        <option value="postgres-16">PostgreSQL 16</option>
                                        <option value="postgres-17">PostgreSQL 17</option>
                                    </select>
                                </div>

                                <div>
                                    <label className="block text-sm font-medium text-gray-700">
                                        PostgreSQL Port
                                    </label>
                                    <input
                                        type="number"
                                        value={formData.pgPort}
                                        onChange={(e) =>
                                            setFormData({ ...formData, pgPort: e.target.value })
                                        }
                                        className="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 focus:border-indigo-500 focus:outline-none"
                                    />
                                </div>

                                <div className="flex items-center">
                                    <input
                                        type="checkbox"
                                        checked={formData.exposeInternet}
                                        onChange={(e) =>
                                            setFormData({
                                                ...formData,
                                                exposeInternet: e.target.checked,
                                            })
                                        }
                                        className="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
                                    />
                                    <label className="ml-2 text-sm text-gray-700">
                                        Expose to Internet
                                    </label>
                                </div>
                            </>
                        )}

                        {(node.type === 'pageserver' || node.type === 'safekeeper') && (
                            <div>
                                <label className="block text-sm font-medium text-gray-700">
                                    Backup Storage ID
                                </label>
                                <select
                                    value={formData.backupStorageId}
                                    onChange={(e) =>
                                        setFormData({
                                            ...formData,
                                            backupStorageId: e.target.value,
                                        })
                                    }
                                    className="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 focus:border-indigo-500 focus:outline-none"
                                >
                                    <option value="">Select backup storage</option>
                                    <option value="backup-1">S3 Backup 1</option>
                                    <option value="backup-2">S3 Backup 2</option>
                                </select>
                            </div>
                        )}
                    </div>

                    <div className="mt-6 flex justify-end gap-3">
                        <button
                            type="button"
                            onClick={onClose}
                            className="rounded-md border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            className="rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700"
                        >
                            Save
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
}
