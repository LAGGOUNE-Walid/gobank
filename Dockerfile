FROM golang:1.24.1-alpine AS builder
ENV CGO_ENABLED=1
    
WORKDIR /app

RUN apk update && apk add --no-cache gcc musl-dev git


COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o ./bin/gobank ./


RUN go get -u -d github.com/golang-migrate/migrate/v4/cmd/migrate && \
cd $GOPATH/src/github.com/golang-migrate/migrate/cmd/migrate && \
git checkout v4.18.2 && \
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.2


FROM alpine:latest


RUN apk update && apk add --no-cache libgcc sqlite-libs

WORKDIR /app


COPY --from=builder /app/bin/gobank .
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate

COPY ./db ./db
COPY ./migrations ./migrations

COPY ./entrypoint.sh /entrypoint.sh

RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]

EXPOSE 8080
    