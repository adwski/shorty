
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
                  -server-port=$$(random unused-port)

.PHONY: statictest
statictest:
	go vet -vettool=$(which statictest) ./...

.PHONY: lint
lint:
	golangci-lint run ./... -p bugs -e G404

.PHONY: docker
docker:
	docker build -t shorty:latest .
