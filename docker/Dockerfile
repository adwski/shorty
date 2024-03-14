FROM golang:1.21.4-bookworm as builder

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GO111MODULE=on \
    GOARCH=amd64 \
    GOPATH=/go

ARG COMMIT="N/A"
ARG VERSION="N/A"

ADD . /build

WORKDIR /build

RUN <<EOF
    DATETIME=$(date -u +"%Y-%m-%dT%H:%M:%S")
    go mod download
    go build -o shorty \
      -ldflags "-d -w -s -X 'main.buildVer=${VERSION}' -X 'main.buildCommit=${COMMIT}' -X 'main.buildTime=${DATETIME}'" \
      -tags netgo -installsuffix netgo ./cmd/shortener/*.go
    chmod +x /build/shorty
EOF


FROM builder as dev

ENTRYPOINT ["/build/shorty"]


FROM gcr.io/distroless/static as release

WORKDIR /
USER nonroot:nonroot
COPY --from=builder /build/shorty /shorty

ENTRYPOINT ["/shorty"]
