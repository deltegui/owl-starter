.PHONY: init db wasm release run dev test build

# Makes conf.json from conf.example.json with a random 32 byte secret. It does not touch an existing one.
init:
	@[ -f conf.json ] && echo "conf.json already exists" || { \
		sed "s/REPLACE-WITH-32-BYTES-OF-SECRET!!/$$(openssl rand -hex 16)/" conf.example.json > conf.json; echo "made conf.json"; }

# Postgres in docker (see docker-compose.yml).
db:
	docker compose up -d postgres

# The one WASM of the app, and the wasm_exec.js that goes with the Go that built it.
# The server embeds web/static, so build this before `go build`.
wasm:
	mkdir -p web/static/js
	GOOS=js GOARCH=wasm go build -o web/static/js/app.wasm ./client/app
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" web/static/js/

# The same WASM, much smaller, with TinyGo.
release:
	mkdir -p web/static/js
	GOOS=js GOARCH=wasm tinygo build --no-debug -opt=z -o web/static/js/app.wasm ./client/app
	cp "$$(tinygo env TINYGOROOT)/targets/wasm_exec.js" web/static/js/

run: wasm
	go run ./cmd/app

# Rebuilds on change. Needs air: go install github.com/air-verse/air@latest
dev: wasm
	air

test:
	go test -race ./...
	GOOS=js GOARCH=wasm go vet ./...

build: wasm
	go build -o bin/app ./cmd/app
