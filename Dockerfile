# -------- Builder Stage --------
    FROM golang:1.24.1-alpine AS builder

    # Enable CGO
    ENV CGO_ENABLED=1
    
    # Set working directory
    WORKDIR /app
    
    # Install build dependencies
    RUN apk update && apk add --no-cache gcc musl-dev git
    
    # Copy go.mod and go.sum to download dependencies
    COPY go.mod go.sum ./
    RUN go mod download
    
    # Copy source code and build the application
    COPY . .
    RUN go build -o ./bin/gobank ./
    
    # Install the golang-migrate tool
    RUN go get -u -d github.com/golang-migrate/migrate/v4/cmd/migrate && \
        cd $GOPATH/src/github.com/golang-migrate/migrate/cmd/migrate && \
        git checkout v4.18.2 && \
        go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.2
    
    # -------- Final Stage --------
    FROM alpine:latest
    
    # Install runtime dependencies
    RUN apk update && apk add --no-cache libgcc sqlite-libs
    
    WORKDIR /app
    
    # Copy the compiled binary and migrate CLI
    COPY --from=builder /app/bin/gobank .
    COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
    
    # Copy migrations and SQLite DB if needed
    COPY ./db ./db
    COPY ./migrations ./migrations
    
    COPY ./entrypoint.sh /entrypoint.sh
    
    # Make the entrypoint script executable
    RUN chmod +x /entrypoint.sh
    
    # Set the entrypoint
    ENTRYPOINT ["/entrypoint.sh"]
    
    EXPOSE 8080
    