include .env
export

export PROJECT_ROOT=${shell pwd}

env-up:
	@docker compose up -d sso-postgres

env-down:
	@docker compose down sso-postgres

migrate-create:
	@if [ -z "${seq}" ]; then \
		echo "pls, try again with seq=value, example: seq=init"; \
		exit 1; \
	fi; \
	docker compose run --rm sso-migrate \
	create -ext sql -dir migrations -seq ${seq}

migrate-up:
	@docker compose run --rm sso-migrate \
	-path migrations \
	-database "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@sso-postgres:5432/${POSTGRES_DB}?sslmode=disable" \
	up

migrate-down:
	@docker compose run --rm sso-migrate \
	-path migrations \
	-database "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@sso-postgres:5432/${POSTGRES_DB}?sslmode=disable" \
	down

migrate-force:
	@docker compose run --rm sso-migrate \
	-path migrations \
	-database "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@sso-postgres:5432/${POSTGRES_DB}?sslmode=disable" \
	force 1

sso-run:
	@go run ${PROJECT_ROOT}/cmd/sso/main.go