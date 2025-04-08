#!/bin/sh
set -e

mkdir -p /app/db

if [ ! -f "/app/db/database.sqlite" ]; then
    echo "Initializing database..."
    touch /app/db/database.sqlite
fi


echo "Running migrations..."
migrate -path /app/migrations -database "sqlite3:///app/db/database.sqlite?x-no-tx-wrap=true" up

echo "Starting application..."
exec /app/gobank