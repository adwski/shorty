
.PHONY: test
test:
	go test ./... -v -count=1 -cover

.PHONY: build
build:
	go build -o ./cmd/shortener/shortener ./cmd/shortener/main.go

# Run this like
# > make shortenertest TESTNUM=7
.PHONY: shortenertest
shortenertest: build
	shortenertestbeta -test.v -test.run=^TestIteration$(TESTNUM)$$ \
                  -binary-path=cmd/shortener/shortener \
                  -source-path=. \
                  -server-port=$$(random unused-port) \
                  -file-storage-path=/tmp/short-url-db-test.json

.PHONY: statictest
statictest:
	go vet -vettool=$(which statictest) ./...

.PHONY: lint
lint:
	golangci-lint run ./... -c .golangci.yml

goimports:
	goimports -v -w  .

.PHONY: docker
docker:
	docker build -t shorty:latest .
