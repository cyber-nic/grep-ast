run:
	go run cmd/main.go main .

test:
	go test -v ./...

build:
	go build -o grep-ast cmd/main.go

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
