
# PostgreSQL Read Replica Setup (PostgreSQL 15)

## Goal

Create a read replica for a PostgreSQL primary database using streaming replication.

Architecture:

```text
Primary
Port: 5432
Data Directory:
/opt/homebrew/var/postgresql@15

        |
        | WAL Streaming
        v

Replica
Port: 5433
Data Directory:
~/postgres/replica
```

---

# Concepts

## WAL (Write Ahead Log)

Every write operation:

```sql
INSERT
UPDATE
DELETE
```

is first recorded in PostgreSQL's WAL.

The replica does not receive rows directly.

Instead:

```text
Primary
  |
  | WAL Records
  v
Replica
```

The replica continuously replays WAL records.

---

# Verify Primary Configuration

Connect:

```bash
psql postgres
```

Verify:

```sql
SHOW wal_level;
SHOW max_wal_senders;
SHOW max_replication_slots;
```

Expected:

```text
wal_level = replica
max_wal_senders = 10
max_replication_slots = 10
```

Verify data directory:

```sql
SHOW data_directory;
```

Example:

```text
/opt/homebrew/var/postgresql@15
```

---

# Create Replication User

```sql
CREATE ROLE replicator
WITH REPLICATION
LOGIN
PASSWORD 'replicator';
```

Verify:

```sql
\du
```

You should see:

```text
Replication
```

in the role attributes.

---

# Verify pg_hba.conf

Locate:

```sql
SHOW hba_file;
```

Ensure replication connections are allowed.

Example:

```conf
host replication replicator 127.0.0.1/32 md5
```

For local development, trust authentication also works:

```conf
host replication all 127.0.0.1/32 trust
```

Reload:

```sql
SELECT pg_reload_conf();
```

---

# Ensure PostgreSQL 15 Tools Are Used

Verify:

```bash
which postgres
which psql
which pg_ctl
which pg_basebackup
```

Expected:

```text
/opt/homebrew/opt/postgresql@15/bin/...
```

Verify versions:

```bash
psql --version
pg_ctl --version
pg_basebackup --version
```

Expected:

```text
15.x
```

If not:

```bash
export PATH="/opt/homebrew/opt/postgresql@15/bin:$PATH"
```

Add this to:

```text
~/.zshrc
```

Then:

```bash
source ~/.zshrc
hash -r
```

---

# Create Replica Directory

```bash
mkdir -p ~/postgres/replica
chmod 700 ~/postgres/replica
```

---

# Take Base Backup

This creates the replica.

```bash
pg_basebackup \
  -h localhost \
  -p 5432 \
  -D ~/postgres/replica \
  -U replicator \
  -Fp \
  -Xs \
  -P \
  -R
```

Meaning:

```text
-h localhost     Primary host
-p 5432          Primary port
-D               Replica data directory
-U replicator    Replication user
-P               Show progress
-R               Configure standby automatically
```

---

# Configure Replica Port

Edit:

```text
~/postgres/replica/postgresql.conf
```

Add:

```conf
port = 5433
hot_standby = on
```

---

# Start Replica

```bash
pg_ctl \
  -D ~/postgres/replica \
  -l ~/postgres/replica.log \
  start
```

---

# Verify Replica

Connect:

```bash
psql -p 5433 postgres
```

Run:

```sql
SELECT pg_is_in_recovery();
```

Expected:

```text
t
```

Meaning:

```text
This server is a standby replica.
```

---

# Verify Replication Connection

On Primary:

```sql
SELECT
    client_addr,
    state,
    sync_state
FROM pg_stat_replication;
```

Expected:

```text
1 row
```

showing the replica connection.

---

# Test Replication

## Create Extension

```sql
CREATE EXTENSION IF NOT EXISTS pgcrypto;
```

## Create Table

```sql
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    price NUMERIC(10,2) NOT NULL,
    stock INT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

## Insert Test Data

```sql
INSERT INTO products (name, price, stock)
SELECT
    'Product ' || gs,
    ROUND((random() * 10000)::numeric, 2),
    floor(random() * 1000)::int
FROM generate_series(1, 10000) gs;
```

Verify:

```sql
SELECT COUNT(*) FROM products;
```

Expected:

```text
10000
```

---

# Verify on Replica

Connect:

```bash
psql -p 5433 postgres
```

Run:

```sql
SELECT COUNT(*) FROM products;
```

Expected:

```text
10000
```

Replication is working.

---

# Verify Read-Only Behavior

On Replica:

```sql
INSERT INTO products(name, price, stock)
VALUES ('test', 100, 10);
```

Expected:

```text
ERROR: cannot execute INSERT in a read-only transaction
```

This confirms:

```text
Writes -> Primary
Reads  -> Replica
```

---

# Monitoring Queries

## Primary

Connected replicas:

```sql
SELECT *
FROM pg_stat_replication;
```

---

## Replica

WAL receiver status:

```sql
SELECT *
FROM pg_stat_wal_receiver;
```



# Cleanup

Stop replica:

```bash
pg_ctl -D ~/postgres/replica stop
```

Delete replica:

```bash
rm -rf ~/postgres/replica
rm -f ~/postgres/replica.log
```

Remove replication user:

```sql
DROP ROLE replicator;
```

Primary remains unchanged.
