.PHONY: build install uninstall test check

BIN := whyor

build:
	go build -trimpath -o bin/$(BIN) ./cmd/whyor

install:
	go install -trimpath ./cmd/whyor

uninstall:
	rm -f "$$(go env GOBIN)"/$(BIN) "$$(go env GOPATH)"/bin/$(BIN)

test:
	go vet ./...
	go test ./...

check: test
	go run ./cmd/whyor check ./...
	go run ./cmd/whyor unused ./...
