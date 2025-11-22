import { ServerServiceClientImpl } from "../proto/server/server";
import { LogsServiceClientImpl } from "../proto/logs/logs";
import { grpcClient } from "./grpc-client";

export const serverService = new ServerServiceClientImpl(grpcClient);
export const logsService = new LogsServiceClientImpl(grpcClient);
