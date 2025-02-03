.DEFAULT_GOAL := build

.PHONY:fmt vet build

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

build: vet
	go build -o bin/main cmd/main.go

clean:
	go clean
	rm -f bin/main