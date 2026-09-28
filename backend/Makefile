DB_URL ?= mysql://adam:adamramadhans@tcp(localhost:3306)/task_management
MIGRATE := go run -tags mysql github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1

.PHONY: run test tidy lint migrate-up migrate-down

run:
	go run ./cmd/api

test:
	go test ./...

tidy:
	go mod tidy

lint:
	@out="$$(gofmt -l .)"; \
	if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi
	go vet ./...

migrate-up:
	$(MIGRATE) -path migrations -database "$(DB_URL)" up

migrate-down:
	$(MIGRATE) -path migrations -database "$(DB_URL)" down 1
