module github.com/nats-io/nats.go

go 1.16

replace (
	github.com/nats-io/nats-server/v2 v2.3.2 => github.com/dongnguyenvt/nats-server/v2 v2.3.3
	github.com/nats-io/nkeys v0.3.0 => github.com/dongnguyenvt/nkeys v0.3.3
)

require (
	github.com/golang/protobuf v1.4.3
	github.com/nats-io/nats-server/v2 v2.3.2
	github.com/nats-io/nkeys v0.3.0
	github.com/nats-io/nuid v1.0.1
	google.golang.org/protobuf v1.23.0
)
