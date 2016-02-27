# Diego's BBS API Overview

Diego is Cloud Foundry's next generation runtime, and the BBS (amongst devs, "the babies") is its central orchestrator.

The BBS accepts client requests from within and without a Diego cluster, dispatching work to the [auctioneer](http://github.com/cloudfoundry-incubator/auctioneer) so it can balance work among the [cells](http://github.com/cloudfoundry-incubator/rep). In Diego, there are two distinct types of work:

- [**Tasks**](tasks.md) are one-off processes that Diego guarantees will run at most once.
- [**Long-Running Processes**](lrps.md) (LRPs) are processes that Diego launches and monitors.  Diego can distribute, run, and monitor `N` instances of a given LRP.  When an LRP instance crashes, Diego restarts it automatically.

Tasks and LRPs ultimately run in [Garden](http://github.com/cloudfoundry-incubator/garden) containers on Diego Cells.  The filesystem mounted into these containers can either be a generic rootfs that ships with Diego or an arbitrary Docker image.  Processes spawned in these containers are provided with a set of [environment variables](environment.md) to aid in configuration.

In addition to launching and monitoring Tasks and LRPs, Diego can stream logs (via [doppler](http://github.com/cloudfoundry/loggregator)) out of the container processes to end users, and Diego can route (via the [router](http://github.com/cloudfoundry/gorouter)) incoming web traffic to container processes.

While it is possible to run multi-tenant workload on Diego, the API does not provide strong abstractions and protections around managing such work (e.g. users, organizations, quotas, etc...).  Diego simply runs Tasks and LRPs and it is up to the consumer to provide these additional abstractions.  In the case of Cloud Foundry these responsibilities fall on the [Cloud Controller](http://github.com/cloudfoundry/cloud_controller_ng).

[back](README.md)
