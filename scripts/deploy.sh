#!/bin/sh

set -eu

PROJECT_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
DEPLOY_ENV_FILE=${DEPLOY_ENV_FILE:-.env}

cd "$PROJECT_ROOT"

if [ ! -f "$DEPLOY_ENV_FILE" ]; then
    echo "deployment environment file not found: $DEPLOY_ENV_FILE" >&2
    echo "create it from deployments/.env.production.example first" >&2
    exit 1
fi

docker compose \
    --env-file "$DEPLOY_ENV_FILE" \
    -f deployments/docker-compose.yml \
    pull api

docker compose \
    --env-file "$DEPLOY_ENV_FILE" \
    -f deployments/docker-compose.yml \
    up -d --no-build api
