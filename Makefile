.PHONY: swag

swag:
	swag init -g pkg/api/router/router.go

proto:
	protoc --go_out=. --go-grpc_out=. pkg/pb/*.proto
