module code.cloudfoundry.org/bbs

go 1.26.2

replace (
	code.cloudfoundry.org/bbs/encryption => ./encryption
	code.cloudfoundry.org/bbs/format => ./format
	code.cloudfoundry.org/bbs/models => ./models
)

require (
	code.cloudfoundry.org/auctioneer v0.0.0-20250910193354-1ef7d6c9eefe
	code.cloudfoundry.org/bbs/encryption v0.0.0
	code.cloudfoundry.org/bbs/format v0.0.0
	code.cloudfoundry.org/bbs/models v0.0.0-20260618205254-dc4b9f8d5bc9
	code.cloudfoundry.org/cfhttp/v2 v2.82.0
	code.cloudfoundry.org/clock v1.75.0
	code.cloudfoundry.org/debugserver v0.102.0
	code.cloudfoundry.org/diego-db-helpers v0.3.0
	code.cloudfoundry.org/diego-logging-client v0.112.0
	code.cloudfoundry.org/durationjson v0.77.0
	code.cloudfoundry.org/go-loggregator/v9 v9.2.1
	code.cloudfoundry.org/inigo v0.0.0-20250908175034-b7230e46c815
	code.cloudfoundry.org/lager/v3 v3.74.0
	code.cloudfoundry.org/locket v1.1.0
	code.cloudfoundry.org/rep v0.1442.0
	code.cloudfoundry.org/routing-info v0.1.0
	code.cloudfoundry.org/tlsconfig v0.60.0
	code.cloudfoundry.org/workpool v0.0.0-20250911194158-1489753f182e
	github.com/go-sql-driver/mysql v1.10.0
	github.com/go-test/deep v1.1.1
	github.com/gogo/protobuf v1.3.2
	github.com/jackc/pgx/v5 v5.10.0
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d
	github.com/onsi/ginkgo/v2 v2.31.0
	github.com/onsi/gomega v1.42.0
	github.com/openzipkin/zipkin-go v0.4.3
	github.com/tedsuo/ifrit v0.0.0-20260418191334-846868129986
	github.com/tedsuo/rata v1.0.0
	github.com/vito/go-sse v1.1.3
	google.golang.org/grpc v1.81.1
)

require (
	code.cloudfoundry.org/ecrhelper v0.0.0-20250911193847-5bf65e63bab5 // indirect
	code.cloudfoundry.org/executor v0.1442.0 // indirect
	code.cloudfoundry.org/garden v0.0.0-20260617020226-a9e754564bb5 // indirect
	code.cloudfoundry.org/go-diodes v0.0.0-20260615142411-472d6bcdb3c6 // indirect
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/Masterminds/semver/v3 v3.5.0 // indirect
	github.com/aws/aws-sdk-go v1.55.8 // indirect
	github.com/aws/aws-sdk-go-v2 v1.42.0 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.32.25 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.19.24 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.29 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.29 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.29 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.30 // indirect
	github.com/aws/aws-sdk-go-v2/service/ecr v1.58.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/ecrpublic v1.39.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.12 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.29 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.2.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.31.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.36.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.43.3 // indirect
	github.com/aws/smithy-go v1.27.2 // indirect
	github.com/awslabs/amazon-ecr-credential-helper/ecr-login v0.12.0 // indirect
	github.com/bmizerany/pat v0.0.0-20210406213842-e4b6760bdd6f // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20260604005048-7023385849c0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sirupsen/logrus v1.9.4 // indirect
	github.com/square/certstrap v1.3.0 // indirect
	go.step.sm/crypto v0.83.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/crypto v0.53.0 // indirect
	golang.org/x/mod v0.37.0 // indirect
	golang.org/x/net v0.56.0 // indirect
	golang.org/x/sync v0.21.0 // indirect
	golang.org/x/sys v0.46.0 // indirect
	golang.org/x/text v0.38.0 // indirect
	golang.org/x/tools v0.46.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260618152121-87f3d3e198d3 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
