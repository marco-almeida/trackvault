lint:
	golangci-lint run --fix
	golangci-lint fmt

run:
	go run main.go