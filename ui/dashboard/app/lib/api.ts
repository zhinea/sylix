import { ServerServiceClientImpl } from "../proto/server/server";
import { grpcClient } from "./grpc-client";

export const serverService = new ServerServiceClientImpl(grpcClient);
