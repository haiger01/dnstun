

all: client server

libs: lib/ip lib/songgao lib/tonnerre

client: libs
	go build -o bin/client src/drivec.go

server: libs
	go build -o bin/server src/*.go

bin-dir:
	mkdir -p bin

clean:
	rm -rf bin
