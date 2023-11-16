
test:
	go test ./... -v -count=1 -cover

lint:
	golangci-lint run ./... -p bugs -e G404

docker:
	docker build -t shorty:latest .
