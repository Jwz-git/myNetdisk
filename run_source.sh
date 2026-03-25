#!/usr/bin/env bash
set -euo pipefail

# Local source run script.
# Usage:
#   ./run_source.sh
# Optional environment variables:
#   DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, INIT_DB

DB_HOST="${DB_HOST:-127.0.0.1}"
DB_PORT="${DB_PORT:-3306}"
DB_USER="${DB_USER:-root}"
DB_PASSWORD="${DB_PASSWORD:-}"
DB_NAME="${DB_NAME:-personal_disk}"
INIT_DB="${INIT_DB:-1}"

export DB_HOST DB_PORT DB_USER DB_PASSWORD DB_NAME

echo "[1/4] Ensure uploads directory exists..."
mkdir -p uploads

echo "[2/4] Start MySQL service..."
if command -v brew >/dev/null 2>&1; then
  # Use Homebrew to start MySQL
  brew services start mysql
  # Wait a bit for MySQL to start
  sleep 3
elif command -v mysql.server >/dev/null 2>&1; then
  # Use mysql.server to start MySQL
  sudo mysql.server start
  # Wait a bit for MySQL to start
  sleep 3
else
  echo "[2/4] Warning: Could not find MySQL service management tool. Please start MySQL manually."
fi

if [[ "${INIT_DB}" == "1" ]]; then
  if command -v mysql >/dev/null 2>&1; then
    echo "[3/4] Ensure database exists (${DB_NAME})..."
    MYSQL_PWD="${DB_PASSWORD}" mysql \
      -h "${DB_HOST}" \
      -P "${DB_PORT}" \
      -u "${DB_USER}" \
      -e "CREATE DATABASE IF NOT EXISTS \`${DB_NAME}\` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
  else
    echo "[3/4] Skip database initialization: mysql client not found."
  fi
else
  echo "[3/4] Skip database initialization: INIT_DB=${INIT_DB}."
fi

echo "[4/4] Download Go modules..."
go mod download

echo "[5/5] Start app..."
go run .
