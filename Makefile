VERSION=`cat VERSION`

prepare:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.52.2

modules:
	go mod download
	go mod verify
	go mod tidy -v

lint: prepare
	golangci-lint run

test:
	go test -coverprofile=c.out
	go tool cover -html=c.out -o coverage.html

build:
	go build -ldflags="-X main.Version=v${VERSION}" -o letter-generator
