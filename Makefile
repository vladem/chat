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
acceptance:  # docker_uds arg is required, which is, in my case, '/run/user/"$(id -u)"/docker.sock'
	docker run --name=chat_tests --rm -v $(docker_uds):/var/run/docker.sock chat-tests:latest
build_server:
	docker build --tag chat-server -f dockerfiles/server.Dockerfile .
build_client:
	docker build --tag chat-client -f dockerfiles/client.Dockerfile .
build_tests_only:
	docker build -f dockerfiles/tests.Dockerfile --tag chat-tests .
build_tests: build_server build_client build_tests_only
run_server_d:
	docker run -d --name=chat_server --rm chat-server:latest
stop_server:
	docker kill chat_server
run_client_d:
	docker run -i --net=container:chat_server chat-client:latest /client --receiver_id $(them) --sender_id $(me)
