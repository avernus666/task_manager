#!/bin/sh
set -eu

if [ "${AUTO_MIGRATE:-true}" = "true" ]; then
  if [ -z "${DATABASE_URL:-}" ]; then
    echo "DATABASE_URL is required to run database migrations" >&2
    exit 1
  fi
  echo "Applying database migrations..."
  migrate -path /app/migrations -database "$DATABASE_URL" up
fi

exec /app/task-manager
