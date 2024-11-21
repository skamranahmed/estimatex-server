dep:
	@go mod tidy
	@go mod download

run:
	@go run main.go || true

.PHONY: dep run