#!/bin/sh
set -eu

command_name="${1:-up}"
database_url="${MIGRATION_DATABASE_URL:-postgres://cash:cash@postgres:5432/cash?sslmode=disable}"

case "$command_name" in
  up)
    exec make migrate-up MIGRATION_DATABASE_URL="$database_url"
    ;;
  down)
    exec make migrate-down MIGRATION_DATABASE_URL="$database_url"
    ;;
  *)
    echo "用法: scripts/migrate.sh [up|down]" >&2
    exit 2
    ;;
esac
