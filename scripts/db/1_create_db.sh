#!/bin/bash

# Create the db if first run and not exists
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE DATABASE letsgo;
EOSQL