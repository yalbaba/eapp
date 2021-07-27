module eapp

go 1.15

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/coreos/etcd v3.3.22+incompatible
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/protobuf v1.3.3
	github.com/google/uuid v1.1.1 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2 // indirect
	github.com/json-iterator/go v1.1.10
	github.com/lestrrat-go/file-rotatelogs v2.4.0+incompatible
	github.com/lestrrat-go/strftime v1.0.4 // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/nsqio/go-nsq v1.0.8
	go.uber.org/zap v1.15.0
	google.golang.org/grpc v1.29.1
)

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0
