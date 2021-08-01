# syntax=docker/dockerfile:1
FROM golang:1.16.6 as nats-build

COPY . /build
WORKDIR /build

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o nats-client ./examples/nats-cluster-bench/client/cmd

FROM scratch
#FROM debian:stretch

COPY --from=nats-build /build/nats-client /nats-client

ENTRYPOINT ["/nats-client"]
CMD ["-p", "8080"]
