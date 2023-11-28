FROM golang:1.21.4-bookworm

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GO111MODULE=on \
    GOARCH=amd64 \
    GOPATH=/go

ADD . /build

WORKDIR /build

# run dev tests just in case
RUN go test -count=1 -v -cover ./...

# build
RUN go build -o shorty ./cmd/shortener/*.go \
    && chmod +x /build/shorty

ENTRYPOINT ["/build/shorty"]
