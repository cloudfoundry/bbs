# Diego BBS API Docs

Diego's Bulletin Board System (BBS) is the central data store and orchestrator of a Diego cluster. It communicates via protocol-buffer-encoded RPC-style calls over HTTP.

Diego clients communicate with the BBS via an [ExternalClient](https://godoc.org/github.com/cloudfoundry-incubator/bbs#ExternalClient) interface. This interface allows clients to create, read, update, delete, and subscribe to events about Tasks and LRPs.

## Table of Contents

- [API Overview](overview.md)
    - Implementation Details
- [Overview of Tasks](tasks.md)
    - [Defining Tasks](defining-tasks.md)
    - [Task Examples](task-examples.md)
- [Overview of Long Running Processes](lrps.md) (LRPs)
    - [Defining LRPs](defining-lrps.md)
    - [LRP Examples](lrp-examples.md)
- [Overview of Cells](cells.md)
- [Actions](actions.md)
- Domains
- Event Streams
- The Container Runtime Environment
- API Reference
    - Authorization
    - [Tasks](api-tasks.md)
    - [LRPs](api-lrps.md)
    - [Cells](api-cells.md)
    - Domains
    - Events
- [Fields common to Tasks and LRPs](common-models.md)

Many of the documents are still in the process of being converted from the documents about the now obsolete [receptor](https://github.com/cloudfoundry-incubator/receptor/tree/master/doc) component.
