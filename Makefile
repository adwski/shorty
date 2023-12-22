.PHONY: mock
mock:
	find . -type d -name "mock*" -exec rm -rf {} +
	mockery

.PHONY: unittests
unittests: mock
	go test ./... -v -count=1 -cover

.PHONY: build
build:
	go build -race -o ./cmd/shortener/shortener ./cmd/shortener/*.go

# Run it like this
# > make shortenertest TESTNUM=7
.PHONY: shortenertest
shortenertest: build
		shortenertestbeta -test.v -test.run=^TestIteration$$TESTNUM$$ \
                      -binary-path=cmd/shortener/shortener \
                      -source-path=. \
                      -database-dsn='postgres://shorty:shorty@127.0.0.1/shorty?sslmode=disable'

.PHONY: shortenertests
shortenertests: build
	for num in 1 2 3 4 5 6 7 8 9; do \
		shortenertestbeta -test.v -test.run=^TestIteration$$num$$ \
                      -binary-path=cmd/shortener/shortener \
                      -source-path=. \
                      -server-port=$$(random unused-port) \
                      -file-storage-path=/tmp/short-url-db-test.json || exit 1 ; \
	done
	for num in 10 11 12 13 14 15; do \
		shortenertestbeta -test.v -test.run=^TestIteration$$num$$ \
                      -binary-path=cmd/shortener/shortener \
                      -source-path=. \
                      -database-dsn='postgres://shorty:shorty@127.0.0.1/shorty?sslmode=disable' || exit 1 ; \
	done

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
test: mock goimports lint unittests statictest shortenertests

.PHONY: integration-tests
integration-tests: docker-dev
	go test ./... -v -count=1 -cover --tags=integration -run=TestDatabase*

.PHONY: image-release
image-release:
	docker build -t shorty:release --target release -f docker/Dockerfile .

.PHONY: image-dev
image-dev:
	docker build -t shorty:dev --target dev -f docker/Dockerfile .

.PHONY: docker-dev
docker-dev:
	cd docker ;\
	docker compose up -d
	docker ps

.PHONY: docker-dev-clean
docker-dev-clean:
	cd docker ;\
	docker compose down -v
