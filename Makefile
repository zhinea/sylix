
run:
	go run ./cmd/main.go

compile-proto-frontend:
	rm -rf ./internal/infra/proto/* && \
	protoc --proto_path=proto proto/frontend/*.proto \
		--go_out=/home/adzin/projects/sylix \
		--go-grpc_out=/home/adzin/projects/sylix \
		--experimental_allow_proto3_optional