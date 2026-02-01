tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/yoheimuta/protolint/cmd/protolint@latest
	go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "checking protobuf compiler, if it fails follow guide at https://protobuf.dev/installation/"
	@which -s protoc && echo OK || exit 1


proto:
	protoc \
	--proto_path api/proto \
	--go_out pkg/gen \
	--go-grpc_out pkg/gen \
	--go_opt paths=source_relative \
	--go-grpc_opt paths=source_relative \
	api/proto/*.proto

protolint:
	protolint lint api/proto

golint:
	golangci-lint run \
	./pkg/... \
	./services/knowledge/... \

lint: protolint golint

down:
	docker compose down

up: down
	docker compose up --build -d

