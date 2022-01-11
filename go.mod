module go.ligato.io/cn-infra/v2

go 1.13

require (
	cloud.google.com/go v0.85.0 // indirect
	github.com/DataDog/zstd v1.3.5 // indirect
	github.com/Shopify/sarama v1.20.0
	github.com/Shopify/toxiproxy v2.1.4+incompatible // indirect
	github.com/Songmu/prompter v0.0.0-20150725163906-b5721e8d5566
	github.com/alicebob/gopher-json v0.0.0-20180125190556-5a6b3ba71ee6 // indirect
	github.com/alicebob/miniredis v2.4.5+incompatible
	github.com/armon/go-metrics v0.3.9 // indirect
	github.com/bitly/go-hostpool v0.0.0-20171023180738-a3a6125de932 // indirect
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/boltdb/bolt v1.3.2-0.20180302180052-fd01fc79c553
	github.com/bshuster-repo/logrus-logstash-hook v0.4.1
	github.com/bsm/sarama-cluster v2.1.15+incompatible
	github.com/coreos/go-systemd v0.0.0-20181031085051-9002847aa142 // indirect
	github.com/coreos/pkg v0.0.0-20180108230652-97fdf19511ea // indirect
	github.com/eapache/go-resiliency v1.1.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/eknkc/amber v0.0.0-20171010120322-cdade1c07385 // indirect
	github.com/evalphobia/logrus_fluent v0.4.0
	github.com/fluent/fluent-logger-golang v1.3.0 // indirect
	github.com/frankban/quicktest v1.7.2 // indirect
	github.com/fsnotify/fsnotify v1.4.7
	github.com/ghodss/yaml v1.0.0
	github.com/go-redis/redis v6.14.2+incompatible
	github.com/gocql/gocql v0.0.0-20181030013202-a84ce58083d3
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/golang/protobuf v1.5.2
	github.com/gomodule/redigo v2.0.0+incompatible // indirect
	github.com/gorilla/context v1.1.1 // indirect
	github.com/gorilla/mux v1.6.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.1-0.20190118093823-f849b5445de4
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/hashicorp/consul/api v1.12.0
	github.com/hashicorp/consul/sdk v0.8.0
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-msgpack v0.5.5 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/howeyc/crc16 v0.0.0-20171223171357-2b2a61e366a6
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/maraino/go-mock v0.0.0-20180321183845-4c74c434cd3a
	github.com/mitchellh/mapstructure v1.1.2
	github.com/namsral/flag v1.7.4-pre
	github.com/onsi/gomega v1.4.3
	github.com/philhofer/fwd v1.0.0 // indirect
	github.com/pierrec/lz4 v2.3.0+incompatible // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.4.0
	github.com/rcrowley/go-metrics v0.0.0-20181016184325-3113b8401b8a // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/pflag v1.0.3 // indirect
	github.com/tinylib/msgp v1.0.2 // indirect
	github.com/unrolled/render v0.0.0-20180914162206-b9786414de4d
	github.com/willfaught/gockle v0.0.0-20160623235217-4f254e1e0f0a
	github.com/yuin/gopher-lua v0.0.0-20181031023651-12c4817b42c5 // indirect
	go.etcd.io/etcd v0.5.0-alpha.5.0.20210419091813-4276c3302675
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/net v0.0.0-20210503060351-7fd8e65b6420
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
	google.golang.org/grpc v1.38.0
	google.golang.org/protobuf v1.27.1
)

replace github.com/openhistogram/circonusllhist => github.com/circonus-labs/circonusllhist v0.3.0

replace go.etcd.io/etcd => github.com/coreos/etcd v0.5.0-alpha.5.0.20210419091813-4276c3302675

replace google.golang.org/grpc => google.golang.org/grpc v1.27.0
