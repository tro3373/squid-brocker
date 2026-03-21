BIN := bin/squid-helper
clean:
	rm -rf bin/
build:
	go build -o $(BIN) ./cmd/squid-helper
test:
	go test ./... -v -count=1
lint:
	golangci-lint run ./...
dev:
	echo "192.168.1.100 www.youtube.com" | go run ./cmd/squid-helper -config rules.yaml
