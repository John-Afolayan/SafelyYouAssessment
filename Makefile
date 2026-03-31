.PHONY: build run test test-race lint clean

BINARY := fleet-monitor
ADDR   := 127.0.0.1:6733

build:
	go build -o $(BINARY) ./main

run: build
	./$(BINARY) --addr $(ADDR)

test:
	go test ./... -v -count=1

test-race:
	go test ./... -v -race -count=1

lint:
	go vet ./...

clean:
	rm -f $(BINARY)
	