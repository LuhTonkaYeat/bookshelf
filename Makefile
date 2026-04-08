.PHONY: proto
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       proto/*.proto

.PHONY: run-validator
run-validator:
	go run cmd/validator/main.go

.PHONY: run-bookshelf
run-bookshelf:
	go run cmd/bookshelf/main.go

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: grpcurl-test
grpcurl-test:
	grpcurl -plaintext -d '{"author": "Tolkien"}' localhost:50051 validator.AuthorValidator/Validate