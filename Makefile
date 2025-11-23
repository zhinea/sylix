VERSION ?= 0.0.0-dev
LDFLAGS := -X github.com/zhinea/sylix/internal/common.Version=$(VERSION)

run:
	go run -ldflags "$(LDFLAGS)" ./cmd/main.go

dev:
	make compile-proto
	make compile-proto-frontend
	go build -ldflags "$(LDFLAGS)" -o bin/agent cmd/agent/main.go
	gowatch -o ./bin/controlplane -p ./cmd/main.go -ldflags "$(LDFLAGS)"

build:
	go build -ldflags "$(LDFLAGS)" -o bin/controlplane cmd/main.go
	go build -ldflags "$(LDFLAGS)" -o bin/agent cmd/agent/main.go

compile-proto:
	rm -rf ./internal/infra/proto/*
	protoc --proto_path=proto proto/*/*.proto --go_out=module=github.com/zhinea/sylix:. --go-grpc_out=module=github.com/zhinea/sylix:. --experimental_allow_proto3_optional

compile-proto-frontend:
	mkdir -p ui/dashboard/app/proto
	protoc --plugin=protoc-gen-ts_proto=./ui/dashboard/node_modules/.bin/protoc-gen-ts_proto --ts_proto_out=./ui/dashboard/app/proto --proto_path=proto proto/controlplane/*.proto proto/common/*.proto --ts_proto_opt=esModuleInterop=true --experimental_allow_proto3_optional
