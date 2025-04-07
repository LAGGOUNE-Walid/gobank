#!/bin/sh

# Run migrations
echo "Running migrations..."
migrate -path ./migrations -database "sqlite3://file:./db/database.sqlite3" up

# Start the Go application
echo "Starting the application..."
exec ./gobank
