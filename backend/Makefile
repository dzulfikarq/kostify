.PHONY: run migrate-up migrate-down seed test vet tidy

run:
	go run ./cmd/server

migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down 1

seed:
	go run ./cmd/seed

test:
	go test ./...

vet:
	go vet ./...

tidy:
	go mod tidy
