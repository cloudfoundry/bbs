module code.cloudfoundry.org/bbs/models

go 1.26.2

replace (
	code.cloudfoundry.org/bbs => ..
	code.cloudfoundry.org/bbs/encryption => ../encryption
	code.cloudfoundry.org/bbs/format => ../format
)

require (
	code.cloudfoundry.org/bbs v0.0.0-00010101000000-000000000000
	code.cloudfoundry.org/bbs/format v0.0.0
	code.cloudfoundry.org/lager/v3 v3.74.0
	github.com/gogo/protobuf v1.3.2
	github.com/onsi/ginkgo/v2 v2.31.0
	github.com/onsi/gomega v1.42.0
)

require (
	code.cloudfoundry.org/bbs/encryption v0.0.0 // indirect
	code.cloudfoundry.org/clock v1.75.0 // indirect
	code.cloudfoundry.org/diego-db-helpers v0.4.0 // indirect
	code.cloudfoundry.org/locket v1.2.0 // indirect
	code.cloudfoundry.org/tlsconfig v0.60.0 // indirect
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/Masterminds/semver/v3 v3.5.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-sql-driver/mysql v1.10.0 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/go-test/deep v1.1.1 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20260604005048-7023385849c0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.10.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d // indirect
	github.com/openzipkin/zipkin-go v0.4.3 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/tedsuo/ifrit v0.0.0-20260418191334-846868129986 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/mod v0.37.0 // indirect
	golang.org/x/net v0.56.0 // indirect
	golang.org/x/sync v0.21.0 // indirect
	golang.org/x/sys v0.46.0 // indirect
	golang.org/x/text v0.38.0 // indirect
	golang.org/x/tools v0.46.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260618152121-87f3d3e198d3 // indirect
	google.golang.org/grpc v1.81.1 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
