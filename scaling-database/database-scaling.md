# Vertical Scaling #

- add more resources to the database server
  - CPU
  - RAM
  - Disk
- easy to perform
- require downtime during reboot
- limited by physical hardware

---

# Horizontal Scaling #

## Read Replica ##

Used when the system has much more reads than writes.

Typical ratio:

```text
Read : Write = 90 : 10
```

Instead of sending everything to one database, reads are moved to replica databases so the master database can focus on writes.

The API server should know:
- which database handles writes
- which database handles reads

---

## Procedure ##

At the API server level:

- create 2 database connection objects
  - master DB connection
  - replica DB connection

Routing logic:

```text
if write query:
    use master connection

else:
    use replica connection
```

---

## Steps to setup master and replica ##

Start replication environment:

```bash
chmod +x master/init-replication.sh

docker compose up -d
```

Connect to master database:

```bash
docker exec -it pg-master psql -U postgres -d app
```

Create test table:

```sql
CREATE TABLE test (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL
);
```

Tear down infrastructure:

```bash
docker compose down -v
```

---

## Synchronous Replication ##

```text
Client <---- w ----> API <---- w ----> MASTER <----W----> REPLICA
```

A write request comes to the API server.

The API writes to the master database.

The write also goes to the replica database.

Either:
- the master sends the write to the replica
- or the API handles writes to both databases

The API does not send a success response back to the client unless the write succeeds on:
- master
- replica

### Characteristics ###

- strong consistency
- zero replication lag
- slower writes

---

## Asynchronous Replication ##

(commonly used)

```text
Client <---- w ----> API <---- w ----> MASTER <------> REPLICA
```

A write request comes to the API server.

The API writes to the master database.

The response is immediately sent back to the client.

The replica pulls data from the master periodically.

Because of this, the replica may temporarily stay behind the master.

### Characteristics ###

- eventual consistency
- some replication lag
- faster writes
- commonly used in production systems
