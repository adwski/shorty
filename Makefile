.PHONY: mock
mock:
	find . -type d -name "mock*" -exec rm -rf {} +
	mockery

.PHONY: unittests
unittests: mock
	go test ./... -v -count=1 -cover

.PHONY: build
build:
	go build -o ./cmd/shortener/shortener ./cmd/shortener/*.go

# Run it like this
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
	go vet -vettool=$$(which statictest) ./...

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: goimports
goimports:
	goimports -w  .

.PHONY: test
test: mock goimports lint unittests statictest
	for num in 1 2 3 4 5 6 7 8 9; do \
		$(MAKE) shortenertest TESTNUM=$$num ; \
	done

.PHONY: docker
docker:
	docker build -t shorty:latest .
