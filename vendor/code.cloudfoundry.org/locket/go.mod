module code.cloudfoundry.org/locket

go 1.16

require (
	code.cloudfoundry.org/cfhttp/v2 v2.0.0
	code.cloudfoundry.org/clock v1.0.0
	code.cloudfoundry.org/consuladapter v0.0.0-20200131002136-ac1daf48ba97
	code.cloudfoundry.org/debugserver v0.0.0-20200131002057-141d5fa0e064
	code.cloudfoundry.org/diego-logging-client v0.0.0-20201207211221-6526582b708b
	code.cloudfoundry.org/diegosqldb v0.0.0-00010101000000-000000000000
	code.cloudfoundry.org/durationjson v0.0.0-20200131001738-04c274cd71ed
	code.cloudfoundry.org/executor v0.0.0-20201214152003-d98dd1d962d6 // indirect
	code.cloudfoundry.org/go-diodes v0.0.0-20190809170250-f77fb823c7ee // indirect
	code.cloudfoundry.org/go-loggregator v7.4.0+incompatible
	code.cloudfoundry.org/inigo v0.0.0-20210329165136-c8d03e725437
	code.cloudfoundry.org/lager v2.0.0+incompatible
	code.cloudfoundry.org/rep v0.0.0-20210428021924-ca9cdfaff888 // indirect
	code.cloudfoundry.org/rfc5424 v0.0.0-20201103192249-000122071b78 // indirect
	code.cloudfoundry.org/tlsconfig v0.0.0-20200131000646-bbe0f8da39b3
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/go-test/deep v1.0.7 // indirect
	github.com/gogo/protobuf v1.3.2
	github.com/hashicorp/consul/api v1.8.1
	github.com/jackc/pgx v3.6.2+incompatible // indirect
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d
	github.com/onsi/ginkgo v1.16.1
	github.com/onsi/gomega v1.11.0
	github.com/pkg/errors v0.9.1
	github.com/tedsuo/ifrit v0.0.0-20191009134036-9a97d0632f00
	golang.org/x/net v0.0.0-20210503060351-7fd8e65b6420
	google.golang.org/grpc v1.37.0
)

replace code.cloudfoundry.org/diegosqldb => ../diegosqldb
