# go_backend_misc
Simple bank backend with a variety of backend technologies, written in Go 


## Run postgresql in Docker locally

- Pull image: `docker pull postgres:12-alpine`
- To exec into the container: `docker exec -it <container_hash> sh`
- To log into the postgres CLI: `psql` (password is not required when connecting from localhost, this is the default for for the Postgres image)

## DB migrations

- `migrate create -ext sql -dir <folder_name, e.g. db/migration> -seq <migration_name>`
- `migrate -path db/migration -database "postgresql://root:secret@localhost:54321/simple_bank?sslmode=disable" --verbose up`

## DB tests
- `go get github.com/lib/pq` to get package `lib/pq`

## Transaction in postgresql
- Uses `FOR NO KEY UPDATE` to prevent concurrency + deadlock issues:
    - `FOR UPDATE` will create a lock in `SELECT` until and `UPDATE` occurs in the transacion
    - Deadlock (`NO KEY` prevents a deadlock between an `INSERT` and the concurrent `SELECT FOR UPDATE`)
    - Order is important: if the two same accounts are updated concurrently, make sure they are updated in the same order (to prevent a situation where `processA` locks `account1`, `processB` locks `account2`, and then `processA` cannot access `account2` and viceversa)
- In order to debug concurrent DB transactions, some useful queries ((postgres wiki)[https://wiki.postgresql.org/wiki/Lock_Monitoring]):
```
SELECT blocked_locks.pid     AS blocked_pid,
         blocked_activity.usename  AS blocked_user,
         blocking_locks.pid     AS blocking_pid,
         blocking_activity.usename AS blocking_user,
         blocked_activity.query    AS blocked_statement,
         blocking_activity.query   AS current_statement_in_blocking_process
   FROM  pg_catalog.pg_locks         blocked_locks
    JOIN pg_catalog.pg_stat_activity blocked_activity  ON blocked_activity.pid = blocked_locks.pid
    JOIN pg_catalog.pg_locks         blocking_locks 
        ON blocking_locks.locktype = blocked_locks.locktype
        AND blocking_locks.database IS NOT DISTINCT FROM blocked_locks.database
        AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation
        AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page
        AND blocking_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple
        AND blocking_locks.virtualxid IS NOT DISTINCT FROM blocked_locks.virtualxid
        AND blocking_locks.transactionid IS NOT DISTINCT FROM blocked_locks.transactionid
        AND blocking_locks.classid IS NOT DISTINCT FROM blocked_locks.classid
        AND blocking_locks.objid IS NOT DISTINCT FROM blocked_locks.objid
        AND blocking_locks.objsubid IS NOT DISTINCT FROM blocked_locks.objsubid
        AND blocking_locks.pid != blocked_locks.pid

    JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
   WHERE NOT blocked_locks.granted;
```

```
SELECT blocked_locks.pid     AS blocked_pid,
         blocked_activity.usename  AS blocked_user,
         blocking_locks.pid     AS blocking_pid,
         blocking_activity.usename AS blocking_user,
         blocked_activity.query    AS blocked_statement,
         blocking_activity.query   AS current_statement_in_blocking_process,
         blocked_activity.application_name AS blocked_application,
         blocking_activity.application_name AS blocking_application
   FROM  pg_catalog.pg_locks         blocked_locks
    JOIN pg_catalog.pg_stat_activity blocked_activity  ON blocked_activity.pid = blocked_locks.pid
    JOIN pg_catalog.pg_locks         blocking_locks 
        ON blocking_locks.locktype = blocked_locks.locktype
        AND blocking_locks.DATABASE IS NOT DISTINCT FROM blocked_locks.DATABASE
        AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation
        AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page
        AND blocking_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple
        AND blocking_locks.virtualxid IS NOT DISTINCT FROM blocked_locks.virtualxid
        AND blocking_locks.transactionid IS NOT DISTINCT FROM blocked_locks.transactionid
        AND blocking_locks.classid IS NOT DISTINCT FROM blocked_locks.classid
        AND blocking_locks.objid IS NOT DISTINCT FROM blocked_locks.objid
        AND blocking_locks.objsubid IS NOT DISTINCT FROM blocked_locks.objsubid
        AND blocking_locks.pid != blocked_locks.pid
 
    JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
   WHERE NOT blocked_locks.GRANTED;
```

Deadlock may come from an INSERT query with a foreign key constraint, concurrent with a SELECT query with a FOR UPDATE constraint
