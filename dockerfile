FROM golang:1.17-alpine

WORKDIR /eth-balance-proxy

COPY . /eth-balance-proxy

RUN GOOS=linux GOARCH=amd64 go build -o bin/service /eth-balance-proxy/main.go

FROM alpine:3.18.3

COPY --from=0 /eth-balance-proxy/bin/service /go/bin/service

ENTRYPOINT ["/go/bin/service", "RPC_ENDPOINT"]