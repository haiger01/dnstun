

all: client server

client:
	go build -o bin/client src/client.go

server:
	go build -o bin/server src/server.go
	go build -o bin/debug src/server-debug.go

bin-dir:
	mkdir -p bin

clean:
	rm -rf bin
