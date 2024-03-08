run_postgres_docker:
	docker run --name go-backend-postgres -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -p 54321:5432 -d postgres:12-alpine

stop_postgres_docker:
	docker stop go-backend-postgres

createdb:
	docker exec -it go-backend-postgres createdb --username=root simple_bank

dropdb:
	docker exec -it go-backend-postgres dropdb --username=root simple_bank

migrate:
	migrate -path db/migration -database "postgresql://root:secret@localhost:54321/simple_bank?sslmode=disable" --verbose up

migrate_down:
	migrate -path db/migration -database "postgresql://root:secret@localhost:54321/simple_bank?sslmode=disable" --verbose down