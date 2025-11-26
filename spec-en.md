
The database concept is similar to nodes.
There are four types of nodes: compute, storage broker, pageserver, and safekeeper.
Each node must have a server_id (where the node is stored).

The following are the node specifications:

```json
{
    "nodes": [
        {
            "name": "Compute Engine",
            "description": "The main computing neondb, where this node will run postgres. use ghcr.io/neondatabase/compute-node-vxx",
            "type": "compute",
            "priority_startup": 4,
            "fields": {
                "server_id": {
                    "description": "server where the node will be placed",
                    "relatedTable": "servers"
                },
                "pg_version": {
                    "description": "The version of Postgres to be used, the type is options with several choices.",
                    "options": [
                        "postgres-14",
                        "postgres-15",
                        "postgres-16",
                        "postgres-17",
                        "postgres-18"
                    ]
                },
                "pg_port": {
                    "description": "port to be used by postgres"
                },
                "expose_internet": {
                    "description": "Will the Postgres port be exposed to the internet? The input is a checkbox type."
                }
            },
            "pipes": {
                "in": [
                    {
                        "from": "safekeeper",
                        "imported": {
                            "5454/tcp": "PG/WAL listener (compute writes WAL to safekeeper)"
                        }
                    },
                    {
                        "from": "pageserver",
                        "imported": {
                            "9898/tcp": "Pageserver HTTP API (fetch pages)"
                        }
                    },
                    {
                        "from": "storage_broker",
                        "imported": {
                            "50051/tcp": "Discovery/coordination (gRPC)"
                        }
                    }
                ],
                "out": [
                    {
                        "to": "clients/app",
                        "exported": {
                            "55433/tcp": "PostgreSQL protocol",
                            "3080/tcp": "HTTP admin/metrics (optional)"
                        }
                    }
                ]
            }
        },
        {
            "name": "Pageserver",
            "description": "The main storage engine for neondb, good when one server with Compute Engine.",
            "type": "pageserver",
            "priority_startup": 3,
            "fields": {
                "server_id": {
                    "required": true,
                    "description": "server where the node will be placed",
                    "relatedTable": "servers"
                },
                "backup_storage_id": {
                    "required": true,
                    "description": "The base backup account, use account same as like on the Safekeeper 1",
                    "relatedTable": "backup_storage"
                }
            },
            "pipes": {
                "in": [
                    {
                        "from": "storage_broker",
                        "imported": {
                            "50051/tcp": "Discovery/coordination (gRPC)"
                        }
                    }
                ],
                "out": [
                    {
                        "to": "compute",
                        "exported": {
                            "9898/tcp": "Pageserver HTTP API (fetch pages)"
                        }
                    }
                ]
            }
        },
        {
            "name": "Safekeeper",
            "description": "Safekeepers are the redundant WAL storage service. They receive WAL from the compute node and durably store it.",
            "type": "safekeeper",
            "priority_startup": 2,
            "fields": {
                "server_id": {
                    "required": true,
                    "description": "server where the node will be placed",
                    "relatedTable": "servers"
                },
                "backup_storage_id": {
                    "required": true,
                    "description": "The base backup account for offloading WAL to S3.",
                    "relatedTable": "backup_storage"
                }
            },
            "pipes": {
                "in": [
                    {
                        "from": "storage_broker",
                        "imported": {
                            "50051/tcp": "Discovery/coordination (gRPC)"
                        }
                    }
                ],
                "out": [
                    {
                        "to": "compute",
                        "exported": {
                            "5454/tcp": "WAL acceptor (Postgres protocol)"
                        }
                    },
                    {
                        "to": "pageserver",
                        "exported": {
                            "7676/tcp": "HTTP API (pull WAL)"
                        }
                    }
                ]
            }
        },
        {
            "name": "Storage Broker",
            "description": "The storage broker is a service that coordinates the safekeepers and pageservers.",
            "type": "storage_broker",
            "priority_startup": 1,
            "fields": {
                "server_id": {
                    "required": true,
                    "description": "server where the node will be placed",
                    "relatedTable": "servers"
                }
            },
            "pipes": {
                "out": [
                    {
                        "to": "safekeeper",
                        "exported": {
                            "50051/tcp": "Discovery/coordination (gRPC)"
                        }
                    },
                    {
                        "to": "pageserver",
                        "exported": {
                            "50051/tcp": "Discovery/coordination (gRPC)"
                        }
                    },
                    {
                        "to": "compute",
                        "exported": {
                            "50051/tcp": "Discovery/coordination (gRPC)"
                        }
                    }
                ]
            }
        }
    ]
}
```


jadi pada "ui/dashboard" dibuat halaman baru yaitu "nodes" dan jangan lupa untuk menambahkan nya di sidebar.
halaman "nodes" ini akan full seperti canvas graph (mirip seperti n8n) jadi nanti antar node bisa di sambung sambungkan.

kira kira seperti ini
```
storage broker
    |
compute <-> pageserver
                |
            --------------------------------------------
            safekeeper 1        safekeeper 2        safekeeper 3
```


Each node is stored in a different docker-compose, with its own isolated system. However, specifically for `compute` and `pageserver`, they must be on one docker-compose to maintain good performance.

The flow is as follows.

1. The user creates a canvas node.
2. The flow node is created.
3. The backend parses the Docker Compose file along with the ports to be exposed (the ports to be exposed are random), ensuring that all Docker Compose files are connected to the correct ports. 
4. Deploy according to the server_id using the agent. (agent.deployCompose) by adjusting the `priority_startup` order, where the smallest order must be deployed first.
5. Deploy according to the order, waiting for each node to activate (must wait).
6. When all nodes have been deployed, verify port availability and perform health checks for all services.
7. Send alerts to users.

All of the above processes run in the background. From the user's perspective, it may be better to display real-time logs, as when installing the server agent.

On the canvas, after deployment, it can also be edited or deleted on the node, or moved to another server_id.

If the action is to delete, perform the following steps
1. Verify that the node container is still running or not.
2. If it is still running, stop the container, then delete the data in the container.
3. IMPORTANT: Ensure that the deletion process is performed gracefully and does not interfere with other nodes.
4. Change the credentials on other nodes to delete the credentials on that node. (Also ensure that the startup sequence has been followed.)

If the action is to change the server_id (meaning moving the node)
1. Perform a graceful shutdown on the container node.
2. Create a snapshot on the Docker.
3. Back up to the storage backup selected by the user.
4. On the target server, download the Docker snapshot.
5. Run the downloaded snapshot again.
6. Change the credentials on other nodes to follow the credentials on the new server. (Also ensure that the startup sequence is followed.)

## Acceptance criteria
- [ ] Production-grade postgres config
- [ ] Good perfoma and minimum latency
- [ ] High available
- [ ] Not rendudant code or unused code
- [ ] Good maintanable code, follow the DDD architecture app