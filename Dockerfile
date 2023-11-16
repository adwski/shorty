FROM golang:1.21.4-alpine3.18 as builder

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
RUN go build -o shorty ./cmd/shortener/main.go \
    && chmod +x /build/shorty


FROM alpine:3.18
WORKDIR /shorty

USER 65535

# copy files to image
COPY --from=builder /build/shorty ./
ENTRYPOINT ["/shorty/shorty"]
