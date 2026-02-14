BINARY := differ
CMD := ./cmd/differ

.PHONY: build clean test lint

build:
	go build -o $(BINARY) $(CMD)

clean:
	rm -f $(BINARY)

test:
	go test ./...

lint:
	golangci-lint run ./...
