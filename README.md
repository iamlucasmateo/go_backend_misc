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