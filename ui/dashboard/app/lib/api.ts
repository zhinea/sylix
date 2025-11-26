import { ServerServiceClientImpl } from "../proto/controlplane/server";
import { LogsServiceClientImpl } from "../proto/controlplane/logs";
import { BackupStorageServiceClientImpl } from "../proto/controlplane/backup";
import { NodeServiceClientImpl } from "../proto/controlplane/node";
import { grpcClient } from "./grpc-client";

export const serverService = new ServerServiceClientImpl(grpcClient);
export const logsService = new LogsServiceClientImpl(grpcClient);
export const backupStorageService = new BackupStorageServiceClientImpl(grpcClient);
export const nodeService = new NodeServiceClientImpl(grpcClient);
