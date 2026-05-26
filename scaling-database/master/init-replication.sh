#!/bin/bash
set -e

echo "host replication replicator all trust" >> "$PGDATA/pg_hba.conf"

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE ROLE replicator WITH REPLICATION LOGIN PASSWORD 'postgres';
EOSQL
