
Untuk database nya konsep nya seperti nodes.
Dimana ada empat tipe nodes, yaitu compute, storage broker, pageserver, dan safekeeper.
dimana setiap node nya harus memiliki server_id (tempat dimana node tersebut di simpan)

berikut spec nodes nya
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


dan setiap node itu disimpan di docker-compose yang berbeda beda, dengan sistem isolated nya sendiri sendiri. tapi khusus untuk `compute` dan `pageserver`, harus berada pada 1 docker-compose. agar perfoma tetap bagus.

untuk flow nya seperti ini.

1. user membuat canvas node
2. flow node dibuat
3. backend parsing ke docker-compose beserta port yang akan di gunakan sebagai expose (port untuk di expose bersifat random), memastikan bahwa semua docker-compose sudah terhubung dengan port yang benar 
4. deploy sesuai server_id menggunakan agent. (agent.deployCompose) dengan menyesuaikan urutan `priority_startup` dimana urutan yang paling kecil yang harus lebih dulu di deploy 
5. di deploy sesuai urutan dengan menunggu node satu persatu aktif (harus menunggu) 
6. ketika semua nodes sudah selesai di deploy, lakukan verifikasi ketersedian port, verifikasi health check untuk semua services
7. berikan alert ke user

semua proses di atas berjalan di latar belakang, dari sisi user mungkin lebih bagus untuk di tampilkan realtime logs seperti pada saat menginstall agent server.

pada canvas setelah di deploy juga bisa di edit atau dihapus pada node nya, ataupun di pindahkan ke server_id lain.

jika aksi nya adalah di hapus, lakukan step berikut
1. verifikasi bahwa container node tersebut masih berjalan atau tidak
2. Jika masih berjalan, lakukan stop container, lalu hapus data pada container tersebut.
3. IMPORTANT: pastikan proses delete harus gracefully delete. dan tidak menganggu node lainya.
4. ubah credential pada node lain agar menghapus credential pada node tersebut. (pastikan juga sudah mengikuti urutan startup)

jika aksi nya adalah mengganti server_id (berarti memindahkan node)
1. lakukan gracefully shutdown pada container node tersebut
2. buat sebuah snapshot pada docker tersebut
3. lakukan backup ke storage backup yang dipilih oleh user
4. pada server yang dituju, download snapshot docker
5. jalankan snapshot yang telah di download lagi
6. ubah credential pada node lain agar mengikuti credential pada server yang baru. (pastikan juga sudah mengikuti urutan startup)


## Acceptance criteria
- [ ] Production-grade postgres config
- [ ] Good perfoma and minimum latency
- [ ] High available
- [ ] Not rendudant code or unused code
- [ ] Good maintanable code, follow the DDD architecture app