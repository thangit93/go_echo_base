ifneq (,$(wildcard ./.env))
    include .env
    export
endif

build:
	docker compose build --no-cache
build-dev:
	docker compose -f docker-compose.dev.yml build --no-cache
dev-up:
	docker compose -f docker-compose.dev.yml up -d
dev-down:
	docker compose -f docker-compose.dev.yml down
prod-up:
	docker compose up -d
prod-down:
	docker compose down -v
create-migrate:
	migrate create -ext sql -dir database/migrations -seq $(name)
migrate-up:
	migrate -path database/migrations -database "mysql://${MYSQL_DSN}" up
migrate-down:
	migrate -path database/migrations -database "mysql://${MYSQL_DSN}" down
migrate-rollback:
	migrate -path database/migrations -database "mysql://${MYSQL_DSN}" down $(step)