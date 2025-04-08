FROM golang:1.24.1-alpine AS builder
ENV CGO_ENABLED=1

WORKDIR /app

RUN apk update && apk add --no-cache gcc musl-dev git sqlite-dev
RUN go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.2

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o gobank .

FROM alpine:latest
RUN apk update && apk add --no-cache libgcc sqlite-libs

WORKDIR /app

COPY --from=builder /app/gobank .
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY entrypoint.sh .
COPY migrations/ ./migrations/

RUN chmod +x entrypoint.sh

ENTRYPOINT ["./entrypoint.sh"]
EXPOSE 8080