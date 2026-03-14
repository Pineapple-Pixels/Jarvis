#!/bin/bash
# Creates the asistente database alongside the n8n database.
# This script runs automatically on first postgres container startup.

set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE DATABASE asistente;
    GRANT ALL PRIVILEGES ON DATABASE asistente TO $POSTGRES_USER;
EOSQL
