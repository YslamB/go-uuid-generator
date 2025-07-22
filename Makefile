generate:
	protoc --go_out=./snowflake-pb --go_opt=paths=source_relative --go-grpc_out=./snowflake-pb --go-grpc_opt=paths=source_relative --proto_path=./proto ./proto/snowflake.proto

install:
	@echo "Installing protoc-gen-go and protoc-gen-go-grpc..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Add this to your ~/.zshrc if not set:"
	@echo 'export PATH="$$PATH:$$HOME/go/bin"'
