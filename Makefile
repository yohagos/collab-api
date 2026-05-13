.PHONY: build run dev test tidy clean docker-up docker-down migrate

build:
	mkdir -p bin
	go build -o bin/collab-api ./cmd/api

run: build
	./bin/collab-api

dev:
	go run ./cmd/api

test:
	go test ./... -v -race -short

test-integration:
	go test ./.. -race -v

tidy:
	go mod tidy

clean:
	rm -rf bin/ coverage.out

docker-up:
	docker compose up -d

docker-down:
	docker compose down

migrate-up:
	goose -dir migrations postgres "user=postgres password=postgres dbname=collab sslmode=disable host=localhost" up

migrate-down:
	goose -dir migrations postgres "user=postgres password=postgres dbname=collab sslmode=disable host=localhost" down
