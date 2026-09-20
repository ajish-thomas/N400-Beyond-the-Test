.PHONY: build test lint ingest build-all
build:
	CGO_ENABLED=0 go build -trimpath -o bin/n400 ./cmd/n400
test:
	go test -race ./...
lint:
	go vet ./...
	staticcheck ./...
ingest:
	go run ./cmd/ingest
build-all:
	@set -e; for os in linux darwin windows; do for arch in amd64 arm64; do suffix=; if [ "$$os" = windows ]; then suffix=.exe; fi; CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build -trimpath -o bin/n400-$$os-$$arch$$suffix ./cmd/n400; done; done
