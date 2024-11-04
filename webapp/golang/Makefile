all: app

app: *.go go.mod go.sum
	go build -o app

proto: cache.proto
	protoc --go_out=./grpc --go_opt=paths=source_relative \
	--go-grpc_out=./grpc --go-grpc_opt=paths=source_relative \
	cache.proto