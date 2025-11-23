const GRPC_WEB_URL = import.meta.env.VITE_GRPC_URL || "http://localhost:8082";

export interface Rpc {
  request(service: string, method: string, data: Uint8Array): Promise<Uint8Array>;
}

export class GrpcWebClient implements Rpc {
  async request(service: string, method: string, data: Uint8Array): Promise<Uint8Array> {
    const url = `${GRPC_WEB_URL}/${service}/${method}`;
    
    // Create gRPC-Web frame
    // 1 byte: 0 (compressed flag, 0=none)
    // 4 bytes: length (big endian)
    // N bytes: data
    const frame = new Uint8Array(5 + data.length);
    frame[0] = 0;
    const view = new DataView(frame.buffer);
    view.setUint32(1, data.length, false); // big endian
    frame.set(data, 5);

    const response = await fetch(url, {
      method: "POST",
      headers: {
        "Content-Type": "application/grpc-web+proto",
        "x-grpc-web": "1",
      },
      body: frame,
    });

    if (!response.ok) {
      throw new Error(`gRPC request failed: ${response.statusText}`);
    }

    const responseBuffer = await response.arrayBuffer();
    const responseData = new Uint8Array(responseBuffer);

    // Parse gRPC-Web response frames
    // We expect at least one data frame (flag=0) and maybe trailers (flag=0x80)
    
    let offset = 0;
    let responseMessage: Uint8Array | null = null;
    let grpcStatus = 0;
    let grpcMessage = "";

    while (offset < responseData.length) {
      const flag = responseData[offset];
      const length = new DataView(responseData.buffer, responseData.byteOffset + offset + 1).getUint32(0, false);
      
      if (flag === 0 || flag === 0x00) { // Data frame
        responseMessage = responseData.slice(offset + 5, offset + 5 + length);
      } else if (flag === 0x80) { // Trailers
        const trailersData = responseData.slice(offset + 5, offset + 5 + length);
        const trailers = new TextDecoder().decode(trailersData);
        trailers.split("\r\n").forEach(line => {
          const [key, ...values] = line.split(":");
          if (key && values.length > 0) {
            const value = values.join(":").trim();
            if (key === "grpc-status") {
              grpcStatus = parseInt(value, 10);
            } else if (key === "grpc-message") {
              grpcMessage = value;
            }
          }
        });
      }
      
      // Skip to next frame
      offset += 5 + length;
    }

    if (grpcStatus !== 0) {
      throw new Error(grpcMessage || `gRPC error code: ${grpcStatus}`);
    }

    if (responseMessage) {
      return responseMessage;
    }

    throw new Error("No data frame found in gRPC response");
  }
}

export const grpcClient = new GrpcWebClient();
