module code.cloudfoundry.org/bbs

go 1.26.2

require (
	code.cloudfoundry.org/bbs/encryption v1.11.0
	code.cloudfoundry.org/bbs/format v1.10.0
	code.cloudfoundry.org/bbs/models v1.12.0
	code.cloudfoundry.org/cfhttp/v2 v2.93.0
	code.cloudfoundry.org/clock v1.86.0
	code.cloudfoundry.org/debugserver v0.113.0
	code.cloudfoundry.org/diego-db-helpers v0.15.0
	code.cloudfoundry.org/diego-logging-client v0.123.0
	code.cloudfoundry.org/durationjson v0.88.0
	code.cloudfoundry.org/go-loggregator/v9 v9.2.1
	code.cloudfoundry.org/inigo v0.0.0-20250908175034-b7230e46c815
	code.cloudfoundry.org/lager/v3 v3.85.0
	code.cloudfoundry.org/locket v1.11.0
	code.cloudfoundry.org/routing-info v1.12.0
	code.cloudfoundry.org/tlsconfig v0.65.0
	code.cloudfoundry.org/workpool v0.0.0-20250911194158-1489753f182e
	github.com/go-sql-driver/mysql v1.10.1
	github.com/go-test/deep v1.1.1
	github.com/gogo/protobuf v1.3.2
	github.com/jackc/pgx/v5 v5.11.0
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d
	github.com/onsi/ginkgo/v2 v2.32.1
	github.com/onsi/gomega v1.43.0
	github.com/openzipkin/zipkin-go v0.4.3
	github.com/tedsuo/ifrit v0.0.0-20260813155221-94822c932811
	github.com/tedsuo/rata v1.0.0
	github.com/vito/go-sse v1.1.3
	google.golang.org/grpc v1.83.2
)

require (
	code.cloudfoundry.org/go-diodes v0.0.0-20260831145205-e8366a756183 // indirect
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/Masterminds/semver/v3 v3.5.0 // indirect
	github.com/bmizerany/pat v0.0.0-20210406213842-e4b6760bdd6f // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20260906184651-6331bc6350fe // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/square/certstrap v1.3.0 // indirect
	go.step.sm/crypto v0.90.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/crypto v0.56.0 // indirect
	golang.org/x/mod v0.40.0 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	golang.org/x/tools v0.49.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260904194346-d0f1323225a4 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
)

// pin ifrit until https://github.com/tedsuo/ifrit/pull/48 is merged
replace github.com/tedsuo/ifrit => github.com/tedsuo/ifrit v0.0.0-20260418191334-846868129986
