include .env

PG_URL=postgres://postgres:${PG_PASSWORD}@localhost:5436/spa_dev

.PHONY: compose-up
compose-up: ### up docker-compose
	docker-compose up --build

.PHONY: compose-down
compose-down: ### down docker-compose
	docker-compose down --remove-orphans

.PHONY: migrate-create
migrate-create: ### create new migrations
	migrate create -ext sql -dir migrations 'create_schema'

.PHONY: migrate-up
migrate-up: ### up migrations
	migrate -path migrations -database '${PG_URL}?sslmode=disable' up

.PHONY: migrate-down
migrate-down: ### down migrations
	migrate -path migrations -database '${PG_URL}?sslmode=disable' down
