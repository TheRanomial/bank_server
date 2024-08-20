postgres:
	docker run --name postgres --network bank_network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres


createdb:
	docker exec -it postgres createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgres psql -U postgres -c "DROP DATABASE simple_bank;"

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up

migrateup1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up 1

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down

migratedown1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down 1

sqlc:
	sqlc generate

test:
	go test -v ./...

server:
	go run main.go

serverprod:
	docker run --name simplebank --network bank_network -p 8080:8080 -e DB_SOURCE=postgresql://root:secret@postgres:5432/simple_bank?sslmode=disable him4462/bank_server

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/TheRanomial/bank_server/db/sqlc Store

.PHONY: postgres createdb dropdb migrateup migratedown sqlc test server mock migrateup1 migratedown1

