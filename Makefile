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