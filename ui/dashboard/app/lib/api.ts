import { ServerServiceClientImpl } from "../proto/controlplane/server";
import { LogsServiceClientImpl } from "../proto/controlplane/logs";
import { grpcClient } from "./grpc-client";

export const serverService = new ServerServiceClientImpl(grpcClient);
export const logsService = new LogsServiceClientImpl(grpcClient);
