FROM golang:1.21.4-bookworm as builder

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GO111MODULE=on \
    GOARCH=amd64 \
    GOPATH=/go

ADD . /build

WORKDIR /build

RUN <<EOF
    go mod download
    go build -o shorty -ldflags '-d -w -s' -tags netgo -installsuffix netgo ./cmd/shortener/*.go
    chmod +x /build/shorty
EOF


FROM builder as dev

ENTRYPOINT ["/build/shorty"]


FROM gcr.io/distroless/static as release

WORKDIR /
USER nonroot:nonroot
COPY --from=builder /build/shorty /shorty

ENTRYPOINT ["/shorty"]
