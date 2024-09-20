build:
	@go build -o bin/karn cmd/main.go

run: build
	@./bin/karn

test:
	@go test -v ./...