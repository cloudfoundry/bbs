The Rep
==============

**Note**: This repository should be imported as `code.cloudfoundry.org/rep`.

<p align="center">
  <img src="http://i.imgur.com/3bd2VFS.jpg" alt="Vote Quimby" title="He'd Vote For You" />
</p>

The Rep bids on tasks and schedules them on an associated Executor.

## Reporting issues and requesting features

Please report all issues and feature requests in [cloudfoundry/diego-release](https://github.com/cloudfoundry/diego-release/issues).

#### Learn more about Diego and its components at [diego-design-notes](https://github.com/cloudfoundry/diego-design-notes)


## Run Tests

1. First setup your [GOPATH and install the necessary dependencies](https://github.com/cloudfoundry/diego-release/blob/develop/CONTRIBUTING.md#initial-setup) for running tests.
1. Setup a MySQL server or a postgres server. [Please follow these instructions.](https://github.com/cloudfoundry/diego-release/blob/develop/CONTRIBUTING.md#running-the-sql-unit-tests)
1. Run the tests from the root directory of the rep repo:
```
SQL_FLAVOR=mysql ginkgo -r -p -race
```
