# BBS

[![Go Report
Card](https://goreportcard.com/badge/code.cloudfoundry.org/bbs)](https://goreportcard.com/report/code.cloudfoundry.org/bbs)
[![Go
Reference](https://pkg.go.dev/badge/code.cloudfoundry.org/bbs.svg)](https://pkg.go.dev/code.cloudfoundry.org/bbs)

Bulletin Board System (BBS) is the API to access the database for Diego.
It communicates via protocol-buffer-encoded RPC-style calls over HTTP.

Diego clients communicate with the BBS via an
[ExternalClient](https://godoc.org/github.com/cloudfoundry/bbs#ExternalClient)
interface. This interface allows clients to create, read, update,
delete, and subscribe to events about Tasks and LRPs.

> \[!NOTE\]
>
> This repository should be imported as `code.cloudfoundry.org/bbs`.

# Docs

-   [BBS API Overview](./docs/010-overview.md)
-   [The components of a Diego Cell overview](./docs/011-cells.md)
-   [Cells API](./docs/012-api-cells.md)
-   [Overview of Tasks](./docs/020-tasks.md)
-   [Defining Tasks](./docs/021-defining-tasks.md)
-   [Task Examples](./docs/022-task-examples.md)
-   [Tasks API](./docs/023-api-tasks.md)
-   [Tasks Internal API](./docs/024-api-tasks-internal.md)
-   [Overview of LRPs: Long Running Processes](./docs/030-lrps.md)
-   [Defining LRPs](./docs/031-defining-lrps.md)
-   [LRP Examples](./docs/032-lrp-examples.md)
-   [LRP API Reference](./docs/033-api-lrps.md)
-   [Actual LRPs Internal API](./docs/034-api-lrps-internal.md)
-   [BBS DB Schema](./docs/040-schema-description.md)
-   [BBS API Versioning
    Conventions](./docs/041-revisioning-bbs-api-endpoints.md)
-   [BBS Migrations](./docs/042-bbs-migration.md)
-   [Domains](./docs/050-domains.md)
-   [Container Runtime Environment Variables](./docs/051-environment.md)
-   [BBS Events](./docs/052-events.md)
-   [Actions](./docs/053-actions.md)
-   [BBS Models](./docs/054-common-models.md)

# Contributing

See the [Contributing.md](./.github/CONTRIBUTING.md) for more
information on how to contribute.

# Working Group Charter

This repository is maintained by [App Runtime
Platform](https://github.com/cloudfoundry/community/blob/main/toc/working-groups/app-runtime-platform.md)
under `Diego` area.

> \[!IMPORTANT\]
>
> Content in this file is managed by the [CI task
> `sync-readme`](https://github.com/cloudfoundry/wg-app-platform-runtime-ci/blob/main/shared/tasks/sync-readme/metadata.yml)
> and is generated by CI following a convention.
