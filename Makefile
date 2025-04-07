BINARY_NAME=gobank
BIN_DIR=bin

build:
	go build -o $(BIN_DIR)/$(BINARY_NAME)
run: build
	./$(BIN_DIR)/$(BINARY_NAME)
test:
	go test -v ./...
benchmark:
	go test -bench=. ./...
coverage:
	go test -coverprofile=coverage.out ./... -covermode=count
	go tool cover -html=coverage.out
fmt:
	go fmt ./...
lint:
	staticcheck ./...
clean:
	rm -rf $(BIN_DIR) coverage.out
deps:
	go mod tidy
