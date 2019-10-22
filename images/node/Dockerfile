# Build Gptn in a stock Go builder container
FROM golang:1.12-alpine as builder

RUN apk add --no-cache make gcc musl-dev linux-headers git

RUN git clone -b testnet https://github.com/palletone/go-palletone.git \
    && cd go-palletone \
    && make gptn

# Pull Gptn into a second stage deploy alpine container
FROM alpine:latest

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/go-palletone/build/bin/gptn /usr/local/bin/

WORKDIR /go-palletone

EXPOSE 8545 8546 30303 30303/udp

ENTRYPOINT ["gptn"]
