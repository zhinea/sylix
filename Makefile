
run:
	go run ./cmd/main.go

compile-proto-frontend:
	rm -rf ./internal/infra/proto/* && \
	protoc --proto_path=proto proto/frontend/*.proto proto/common/*.proto \
		--go_out=. \
		--go-grpc_out=. \
		--experimental_allow_proto3_optional