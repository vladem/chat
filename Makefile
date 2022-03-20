install_devtools:  # export PATH="$PATH:$(go env GOPATH)/bin" should be run manually, go compiler is considered to be installed
	sudo apt update \
	&& sudo apt install -y protobuf-compiler \
	&& go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.27.1 \
	&& go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1
codegen:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/chat.proto
run_server:
	go run ./server/
run_client:
	go run ./client/ --sender_id=$(me) --receiver_id=$(them)
test:
	go test ./server/storage/ ./server/chatmanager/
