.PHONY: all build test clean

all: clean build test


build:
	@echo ">>> Building Application..."
	go build -mod=readonly -a -o publish-failure-resolver-go ./cmd/publish-failure-resolver-go 

test:
	@echo ">>> Running Tests..."
	go test -race -v ./...

clean:
	@echo ">>> Removing binaries..."
	@rm ./publish-failure-resolver-go
