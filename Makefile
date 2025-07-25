

build:
	go build -o ./bin/loadtester ./cmd/loadtester

test:
	go test ./...
