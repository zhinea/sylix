
run:
	go run ./cmd/main.go

dev:
	go build -o bin/agent cmd/agent/main.go
	gowatch -o ./bin/controlplane -p ./cmd/main.go

build:
	go build -o bin/controlplane cmd/main.go
	go build -o bin/agent cmd/agent/main.go

compile-proto:
	rm -rf ./internal/infra/proto/*
	protoc --proto_path=proto proto/*/*.proto --go_out=module=github.com/zhinea/sylix:. --go-grpc_out=module=github.com/zhinea/sylix:. --experimental_allow_proto3_optional

compile-proto-frontend:
	mkdir -p ui/dashboard/app/proto
	protoc --plugin=protoc-gen-ts_proto=./ui/dashboard/node_modules/.bin/protoc-gen-ts_proto --ts_proto_out=./ui/dashboard/app/proto --proto_path=proto proto/controlplane/*.proto proto/common/*.proto --ts_proto_opt=esModuleInterop=true --experimental_allow_proto3_optional
