run:
	go run cmd/kwik-mq/main.go

test:
	go test -v ./...

build:
	go build -o kwik-mq cmd/kwik-mq/main.go