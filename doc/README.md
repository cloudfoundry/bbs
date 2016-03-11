# Diego BBS API Docs

Diego's BBS is the central data store and orchestrator of a Diego cluster. It can be communicated with via protobuf-based RPC.

Consumers of Diego communicate with it via an ExternalClient, the golang interface for which can be found [here](https://godoc.org/github.com/cloudfoundry-incubator/bbs#ExternalClient). This client allows you to schedule Tasks and LRPs and to fetch information about running Tasks and LRPs.

Note: These docs are thouroughly incomplete, and are in the process of being moved over from Diego's deprecated JSON receptor API docs. For the incomplete sections, it's often useful to reference these old docs, which can be found [here](https://github.com/cloudfoundry-incubator/receptor/tree/master/doc). However, we make no guarantees about the accuracy of these old docs, and in many cases they reference components and data types that have changed significantly since they were last changed.

Here's an outline:

- [API Overview](overview.md)
   - Implementation Details
- [Understanding Tasks](tasks.md)
   - [Defining Tasks](defining-tasks.md)
   - [Task Examples](task-examples.md)
- Understanding Long Running Processes (LRPs)
   - Defining LRPs
   - LRP Examples
- Container Runtime Environment
- [Understanding Actions](actions.md)
