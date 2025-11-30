Sylix Engine is a database management postgresql (NeonDB focus) with a `controlplane` (master) and `agent` (node) architecture. It includes a React/Vite frontend (`ui/dashboard`).

so, i want create the core of system, where the system can manage the nodes.

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
            "imports": [
                {
                    "from": "safekeeper",
                    "ports": {
                        "5454/tcp": "PG/WAL listener (compute writes WAL to safekeeper)"
                    }
                },
                {
                    "from": "pageserver",
                    "ports": {
                        "9898/tcp": "Pageserver HTTP API (fetch pages)"
                    }
                },
                {
                    "from": "storage_broker",
                    "ports": {
                        "50051/tcp": "Discovery/coordination (gRPC)"
                    }
                }
            ],
            "exports": [
                {
                    "to": "clients/app",
                    "ports": {
                        "55433/tcp": "PostgreSQL protocol",
                        "3080/tcp": "HTTP admin/metrics (optional)"
                    }
                }
            ]
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
            "imports": [
                {
                    "from": "storage_broker",
                    "ports": {
                        "50051/tcp": "Discovery/coordination (gRPC)"
                    }
                }
            ],
            "exports": [
                {
                    "to": "compute",
                    "ports": {
                        "9898/tcp": "Pageserver HTTP API (fetch pages)"
                    }
                }
            ]
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
            "imports": [
                {
                    "from": "storage_broker",
                    "ports": {
                        "50051/tcp": "Discovery/coordination (gRPC)"
                    }
                }
            ],
            "exports": [
                {
                    "to": "compute",
                    "ports": {
                        "5454/tcp": "WAL acceptor (Postgres protocol)"
                    }
                },
                {
                    "to": "pageserver",
                    "ports": {
                        "7676/tcp": "HTTP API (pull WAL)"
                    }
                }
            ]
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
            "exports": [
                {
                    "to": "safekeeper",
                    "ports": {
                        "50051/tcp": "Discovery/coordination (gRPC)"
                    }
                },
                {
                    "to": "pageserver",
                    "ports": {
                        "50051/tcp": "Discovery/coordination (gRPC)"
                    }
                },
                {
                    "to": "compute",
                    "ports": {
                        "50051/tcp": "Discovery/coordination (gRPC)"
                    }
                }
            ]
        }
    ]
}
```


So, on “ui/dashboard,” create a new page called “nodes” and don't forget to add it to the sidebar.
The “nodes” page will be full like a canvas graph (similar to n8n), so later the nodes can be connected to each other.

For each node, only one connection is needed per node. This means that if A is already connected to B, then B does not need to connect the graph to A again.
So when the graph is connected between nodes, it is like an open connection network between nodes.

Then, import and export ports are ports that will be opened or needed by each node. For example, Compute needs `imports` pageserver, so the port result from pageserver must be set in compute.

Then, on the “ui/dashboard” display, when the node is clicked, a modal will appear for the node settings (the settings are like those in `fields`).

for nodes, It will look something like this

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
3. Users showing realtime logs deploying process
4. The backend parses the Docker Compose file along with the ports to be exposed (the ports to be exposed are random), ensuring that all Docker Compose files are connected to the correct ports. 
5. Deploy according to the server_id using the agent. (agent.deployCompose) by adjusting the `priority_startup` order, where the smallest order must be deployed first.
6. Deploy according to the order, waiting for each node to activate (must wait).
7. When all nodes have been deployed, verify port availability and perform health checks for all services.
8. Send alerts to users.

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

# Phases

## Phase 1
- Create a base `node` entity structure (on agent and controlplane)
- Implement the structure to proto file, based on requirements above.
- After the proto file is implemented, run `make compile-proto` and `make compile-proto-frontend`.

## Phase 2
- Implement the `agent` core system node
- Implement the `controlplane` core system node

## Phase 3
- Implement the `ui/dashboard` canvas node, with already implemented gRPC connection to the controlplane.


## Acceptance criteria
- [ ] Production-grade postgres config
- [ ] Good perfoma and minimum latency
- [ ] High available
- [ ] Not rendudant code or unused code
- [ ] Good maintanable code, follow the DDD architecture app