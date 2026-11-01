.PHONY: init db run dev test build

# Makes conf.json from conf.example.json with a random 32 byte secret. It does not touch an existing one.
init:
	@[ -f conf.json ] && echo "conf.json already exists" || { \
		sed "s/REPLACE-WITH-32-BYTES-OF-SECRET!!/$$(openssl rand -hex 16)/" conf.example.json > conf.json; echo "made conf.json"; }

# Postgres in docker (see docker-compose.yml).
db:
	docker compose up -d postgres

run:
	go run ./cmd/app

# Rebuilds on change. Needs air: go install github.com/air-verse/air@latest
dev:
	air

test:
	go test -race ./...

build:
	go build -o bin/app ./cmd/app
