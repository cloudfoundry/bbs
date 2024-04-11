---
title: Where does BBS fits in Cloud Foundry
expires_at: never
tags: [diego-release, bbs]
---

# Where does BBS fits in Cloud Foundry

Diego is Cloud Foundry's container runtime, and the BBS server is its central orchestrator.

The BBS accepts client requests from both inside and outside a Diego cluster, dispatching work to the [auctioneer](http://github.com/cloudfoundry/auctioneer) so it can balance work among the [cells](05-cells-overview.md). In Diego, there are two distinct types of work:

- [**Tasks**](03-a-tasks-overview.md) are one-off processes that Diego guarantees will run at most once.
- [**Long-Running Processes**](04-a-lrps-overview.md) (LRPs) are processes that Diego monitors for health continually.  Diego can distribute, run, and monitor several identical instances of a given LRP. When an LRP instance crashes, Diego restarts it automatically.

Tasks and LRP instances run in [Garden](http://github.com/cloudfoundry/garden) containers on Diego Cells.  The filesystem mounted into these containers can be either a 'preloaded' rootfs colocated with the Diego cell or an arbitrary Docker image. Diego also provides some additional [environment variables](07-environment-overview.md) to processes running in its containers.

In addition to launching and monitoring Tasks and LRPs, Diego streams logs from containers and cells to end users via the [Loggregator system](http://github.com/cloudfoundry/loggregator). Diego also allows clients to store routing data on LRPs. In Cloud Foundry, routing tiers such as the [HTTP Gorouter](http://github.com/cloudfoundry/gorouter) and the [TCP router](https://github.com/cloudfoundry-incubator/cf-tcp-router) use this data to route external traffic to container processes.

Diego provides only a basic notion of client multitenancy via the concept of a [domain](02-domains-overview.md). Enforcement of richer multitenancy, such as quotas for organizations or visibility restrictions for different users, falls on the [Cloud Controller](http://github.com/cloudfoundry/cloud_controller_ng) in the case of Cloud Foundry.
