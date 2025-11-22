
run:
	go run ./cmd/main.go

build:
	go build -o bin/controlplane cmd/main.go
	go build -o bin/agent cmd/agent/main.go

compile-proto:
	rm -rf ./internal/infra/proto/*
	protoc --proto_path=proto proto/*/*.proto --go_out=module=github.com/zhinea/sylix:. --go-grpc_out=module=github.com/zhinea/sylix:. --experimental_allow_proto3_optional
