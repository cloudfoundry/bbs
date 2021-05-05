module code.cloudfoundry.org/bbs

go 1.16

require (
	code.cloudfoundry.org/auctioneer v0.0.0-20201204183739-8cd9c800fbf9
	code.cloudfoundry.org/cfhttp/v2 v2.0.1-0.20210126204043-e96011109905
	code.cloudfoundry.org/clock v1.0.1-0.20200131002207-86534f4ca3a5
	code.cloudfoundry.org/consuladapter v0.0.0-20200131002136-ac1daf48ba97
	code.cloudfoundry.org/debugserver v0.0.0-20200131002057-141d5fa0e064
	code.cloudfoundry.org/diego-logging-client v0.0.0-20201207211221-6526582b708b
	code.cloudfoundry.org/diegosqldb v0.0.0-00010101000000-000000000000
	code.cloudfoundry.org/durationjson v0.0.0-20200131001738-04c274cd71ed
	code.cloudfoundry.org/ecrhelper v0.0.0-20200131001657-9a7c7e5a931d // indirect
	code.cloudfoundry.org/executor v0.0.0-20201214152003-d98dd1d962d6
	code.cloudfoundry.org/garden v0.0.0-20210208153517-580cadd489d2 // indirect
	code.cloudfoundry.org/go-loggregator v7.4.0+incompatible
	code.cloudfoundry.org/inigo v0.0.0-20210329165136-c8d03e725437
	code.cloudfoundry.org/lager v2.0.0+incompatible
	code.cloudfoundry.org/locket v0.0.0-20210126204241-74d8e4fe8d79
	code.cloudfoundry.org/tlsconfig v0.0.0-20200131000646-bbe0f8da39b3
	code.cloudfoundry.org/workpool v0.0.0-20200131000409-2ac56b354115
	github.com/aws/aws-sdk-go v1.38.29 // indirect
	github.com/awslabs/amazon-ecr-credential-helper/ecr-login v0.0.0-20210324191134-efd1603705e9 // indirect
	github.com/bmizerany/pat v0.0.0-20210406213842-e4b6760bdd6f // indirect
	github.com/go-sql-driver/mysql v1.6.0
	github.com/go-test/deep v1.0.7
	github.com/gogo/protobuf v1.3.2
	github.com/hashicorp/consul/api v1.8.1
	github.com/jackc/pgx v3.6.2+incompatible
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d
	github.com/onsi/ginkgo v1.16.1
	github.com/onsi/gomega v1.11.0
	github.com/tedsuo/ifrit v0.0.0-20191009134036-9a97d0632f00
	github.com/tedsuo/rata v1.0.0
	github.com/vito/go-sse v1.0.0
	golang.org/x/net v0.0.0-20210503060351-7fd8e65b6420
	google.golang.org/grpc v1.37.0
)

replace code.cloudfoundry.org/locket => ../locket

replace code.cloudfoundry.org/diegosqldb => ../diegosqldb
