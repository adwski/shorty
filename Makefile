COMMIT = $(shell git rev-parse HEAD)
VERSION = v0.0.1
DATETIME = $(shell date -u +"%Y-%m-%dT%H:%M:%S%z")

.PHONY: mock
mock:
	find . -type d -name "mock*" -exec rm -rf {} +
	mockery

.PHONY: unittests
unittests: mock
	go test ./... -v -count=1 -cover

.PHONY: build
build:
	go build -ldflags "-X 'main.buildVer=$(VERSION)' -X 'main.buildCommit=$(COMMIT)' -X 'main.buildTime=$(DATETIME)'" \
		-gcflags "-m" -race -o ./cmd/shortener/shortener ./cmd/shortener/*.go

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
	for num in 10 11 12 13 14 15 16 17 18; do \
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
	go run cmd/staticlint/main.go ./...

.PHONY: goimports
goimports:
	goimports -w  .

.PHONY: test
test: mock goimports lint unittests statictest

.PHONY: integration-tests
integration-tests: docker-dev db-tests shortenertests

.PHONY: db-tests
db-tests:
	go test ./... -v -count=1 -cover --tags=integration -run=TestDatabase*

.PHONY: test-all-cover
test-all-cover: docker-dev
	go test ./... -v -count=1 -cover -coverpkg=./... -coverprofile=profile.cov --tags=integration
	go tool cover -func=profile.cov

.PHONY: image-release
image-release:
	docker build -t shorty:release \
		--build-arg "COMMIT=$(COMMIT)" --build-arg "VERSION=$(VERSION)" \
		--target release -f docker/Dockerfile .

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

.PHONY: godoc
godoc:
	@echo "spawning godoc"
	@echo "navigate to http://localhost:8111/pkg/github.com/adwski/shorty/?m=all"
	godoc -http=:8111

.PHONY: grpc
grpc:
	protoc --go_out=.  --go-grpc_out=.  internal/grpc/protobuf/shorty.proto
