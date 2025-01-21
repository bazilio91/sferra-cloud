.PHONY: swag proto

swag:
	swag init --parseDependency -g pkg/api/router/router.go -o pkg/api/docs

proto:
	#mkdir -p pkg/pb/models
	$(eval proto_path := $(shell go list -m -f '{{.Dir}}' github.com/cosmos/gogoproto))

	protoc -I=. -I=$(proto_path)/protobuf -I=$(proto_path) \
		--gogofaster_out=Mgoogle/protobuf/any.proto=github.com/cosmos/gogoproto/types,Mgoogle/protobuf/struct.proto=github.com/cosmos/gogoproto/types:. proto/data.proto

	$(eval gorm_proto_path := $(shell go list -m -f '{{.Dir}}' github.com/infobloxopen/protoc-gen-gorm))
	protoc -I=. -I=$(gorm_proto_path)/proto -I=$(proto_path)/protobuf -I=$(proto_path) --go_out=. --gorm_out="engine=postgres:." proto/models.proto

	protoc -I=. -I=$(proto_path) --go_out=. --go-grpc_out=. proto/image_service.proto

	protoc -I=. -I=$(proto_path) -I=$(gorm_proto_path)/proto --go_out=. --go-grpc_out=. proto/image_service.proto
	protoc -I=. -I=$(proto_path) -I=$(gorm_proto_path)/proto --go_out=. --go-grpc_out=. proto/task_service.proto
