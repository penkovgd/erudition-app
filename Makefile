tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/yoheimuta/protolint/cmd/protolint@latest


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

down:
	docker compose down

up: down
	docker compose up --build -d

