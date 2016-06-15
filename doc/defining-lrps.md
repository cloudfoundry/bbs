## Defining LRPs

This document explains the fields available when defining a new LRP. For a higher-level overview of the Diego LRP API, see the [LRPs Overview](lrps.md).

```go
client := bbs.NewClient(url)
err := client.DesireLRP(logger, &models.DesiredLRP{
	ProcessGuid:          "some-guid",
	Domain:               "some-domain",
	RootFs:               "some-rootfs",
	Instances:            1,
	EnvironmentVariables: []*models.EnvironmentVariable{{Name: "FOO", Value: "bar"}},
	CachedDependencies: []*models.CachedDependency{
		{Name: "app bits", From: "blobstore.com/bits/app-bits", To: "/usr/local/app", CacheKey: "cache-key", LogSource: "log-source"},
		{Name: "app bits with checksum", From: "blobstore.com/bits/app-bits-checksum", To: "/usr/local/app-checksum", CacheKey: "cache-key", LogSource: "log-source", ChecksumAlgorithm: "md5", ChecksumValue: "checksum-value"},
	},
	Setup:          models.WrapAction(&models.RunAction{Path: "ls", User: "name"}),
	Action:         models.WrapAction(&models.RunAction{Path: "ls", User: "name"}),
	StartTimeoutMs: 15000,
	Monitor: models.WrapAction(models.EmitProgressFor(
		models.Timeout(models.Try(models.Parallel(models.Serial(&models.RunAction{Path: "ls", User: "name"}))),
			10*time.Second,
		),
		"start-message",
		"success-message",
		"failure-message",
	)),
	DiskMb:      512,
	MemoryMb:    1024,
	Privileged:  true,
	CpuWeight:   42,
	Ports:       []uint32{8080, 9090},
	Routes:      &models.Routes{"my-router": json.RawMessage(`{"foo":"bar"}`)},
	LogSource:   "some-log-source",
	LogGuid:     "some-log-guid",
	MetricsGuid: "some-metrics-guid",
	Annotation:  "some-annotation",
	Network: &models.Network{
		Properties: map[string]string{
			"some-key":       "some-value",
			"some-other-key": "some-other-value",
		},
	},
	EgressRules: []*models.SecurityGroupRule{{
		Protocol:     models.TCPProtocol,
		Destinations: []string{"1.1.1.1/32", "2.2.2.2/32"},
		PortRange:    &models.PortRange{Start: 10, End: 16000},
	}},
	ModificationTag:               &models.NewModificationTag("epoch", 0),
	LegacyDownloadUser:            "legacy-dan",
	TrustedSystemCertificatesPath: "/etc/somepath",
	VolumeMounts: []*models.VolumeMount{
		{
			Driver:        "my-driver",
			VolumeId:      "my-volume",
			ContainerPath: "/mnt/mypath",
			Mode:          models.BindMountMode_RO,
		},
	},
})
```

### LRP Identifiers

#### `process_guid` [required]

It is up to the consumer of Diego to provide a *globally unique* `process_guid`.  To subsequently fetch the DesiredLRP and its ActualLRP you refer to it by its `process_guid`.

- The `process_guid` must include only the characters `a-z`, `A-Z`, `0-9`, `_` and `-`.
- The `process_guid` must not be empty
- If you attempt to create a DesiredLRP with a `process_guid` that matches that of an existing DesiredLRP, Diego will attempt to update the existing DesiredLRP.  This is subject to the rules described in [updating DesiredLRPs](#updating-desiredlrps) below.


#### `domain` [required]

The consumer of Diego may organize LRPs into groupings called 'domains'.  These are purely organizational (for example, for enabling multiple consumers to use Diego without colliding) and have no implications on the ActualLRP's placement or lifecycle.  It is possible to fetch all LRPs in a given domain.

- It is an error to provide an empty `domain` field.

### LRP Placement

In the future Diego will support the notion of Placement Pools via arbitrary tags associated with Cells.

### Instances

#### `instances` [required]

Diego can run and manage multiple instances (`ActualLRP`s) for each `DesiredLRP`.  `instances` specifies the number of desired instances and must not be less than zero.

### Container Contents and Environment

#### `rootfs` [required]

The `rootfs` field specifies the root filesystem to mount into the container.  Diego can be configured with a set of *preloaded* RootFSes.  These are named root filesystems that are already on the Diego Cells.

Preloaded root filesystems look like:

```
"rootfs": "preloaded:ROOTFS-NAME"
```

Diego's [BOSH release](https://github.com/cloudfoundry-incubator/diego-release) ships with the `cflinuxfs2` filesystem root filesystem built to work with the Cloud Foundry buildpacks, which can be specified via
```
"rootfs": "preloaded:cflinuxfs2"
```

It is possible to provide a custom root filesystem by specifying a Docker image for `rootfs`:

```
"rootfs": "docker:///docker-user/docker-image#docker-tag"
```

To pull the image from a different registry than the default (Docker Hub), specify it as the host in the URI string, e.g.:

```
"rootfs": "docker://index.myregistry.gov/docker-user/docker-image#docker-tag"
```

> You *must* provide the dockerimage `rootfs` uri as above, including the leading `docker://`!

> [Lattice](https://github.com/cloudfoundry-incubator/lattice) does not ship with any preloaded root filesystems. You must specify a Docker image when using Lattice. You can mount the filesystem provided by diego-release by specifying `"rootfs": "docker:///cloudfoundry/cflinuxfs2"`.


#### `env` [optional]

Diego supports the notion of container-level environment variables.  All processes that run in the container will inherit these environment variables.

For more details on the environment variables provided to processes in the container, read [Container Runtime Environment](environment.md).

### Container Limits

#### `cpu_weight` [optional]

To control the CPU shares provided to a container, set `cpu_weight`.  This must be a positive number between `1` and `100`, inclusive.  The `cpu_weight` enforces a relative fair share of the CPU among containers.  It's best explained with examples.  Consider the following scenarios (we shall assume that each container is running a busy process that is attempting to consume as many CPU resources as possible):

- Two containers, with equal values of `cpu_weight`: both containers will receive equal shares of CPU time.
- Two containers, one with `"cpu_weight": 50` and the other with `"cpu_weight": 100`: the later will get (roughly) 2/3 of the CPU time, the former 1/3.

#### `disk_mb` [optional]

A disk quota applied to the entire container.  Any data written on top of the RootFS counts against the Disk Quota.  Processes that attempt to exceed this limit will not be allowed to write to disk.

- `disk_mb` must be an integer >= 0
- If set to 0 no disk constraints are applied to the container
- The units are megabytes

#### `memory_mb` [optional]

A memory limit applied to the entire container.  If the aggregate memory consumption by all processs running in the container exceeds this value, the container will be destroyed.

- `memory_mb` must be an integer >= 0
- If set to 0 no memory constraints are applied to the container
- The units are megabytes

#### `privileged` [optional]

If false, Diego will create a container that is in a user namespace.  Processes that succesfully obtain escalated privileges (i.e. root access) will actually only be root within the user namespace and will not be able to maliciously modify the host VM.  If true, Diego creates a container with no user namespace -- escalating to root gives the user *real* root access.

### Actions

When an LRP instance is instantiated, a container is created with the specified `rootfs` mounted.  Diego is responsible for performing any container setup necessary to successfully launch processes and monitor said processes.

#### `setup` [optional]

After creating a container, Diego will first run the action specified in the `setup` field.  This field is optional and is typically used to download files and run (short-lived) processes that configure the container.  For more details on the available actions see [actions](actions.md).

- If the `setup` action fails the `ActualLRP` is considered to have crashed and will be restarted

#### `action` [required]

After completing any `setup` action, Diego will launch the `action` action.  This `action` is intended to launch any long-running processes.  For more details on the available actions see [actions](actions.md).

#### `monitor` [optional]

If provided, Diego will monitor the long running processes encoded in `action` by periodically invoking the `monitor` action.  If the `monitor` action returns succesfully (exit status code 0), the container is deemed "healthy", otherwise the container is deemed "unhealthy".  Monitoring is quite flexible in Diego and is outlined in more detail [below](#monitoring-health).

#### `start_timeout` [optional]

If provided, Diego will give the `action` action up to `start_timeout` seconds to become healthy before marking the LRP as failed.

### Networking

Diego can open and expose arbitrary `ports` inside the container.  There are plans to generalize this support and make it possible to build custom service discovery solutions on top of Diego.  The API is likely to change in backward-incompatible ways as we work these requirements out.

By default network access for any container is limited but some LRPs might need specific network access and that can be setup using `egress_rules` field.  Rules are evaluated in reverse order of their position, i.e., the last one takes precedence.

> Lattice users: Lattice is intended to be a single-tenant cluster environment.  In Lattice there are no network-access constraints on the containers so there is no need to specify `egress_rules`.

#### `ports` [optional]

`ports` is a list of ports to open in the container.  Processes running in the container can bind to these ports to receive incoming traffic.  These ports are only valid within the container namespace and an arbitrary host-side port is created when the container is created.  This host-side port is made available on the `ActualLRP`.

#### `routes` [optional]

`routes` is a map where the keys identify route providers and the values hold information for the providers to consume.  The information in the map must be valid JSON but is not proessed by Diego.  The total length of the routing information must not exceed 4096 bytes.

#### `egress_rules` [optional]
`egress_rules` are a list of egress firewall rules that are applied to a container running in Diego

##### `protocol` [required]
The protocol of the rule that can be one of the following `tcp`, `udp`,`icmp`, `all`.

##### `destinations` [required]
The destinations of the rule that is a list of either an IP Address (1.2.3.4) or an IP range (1.2.3.4-2.3.4.5) or a CIDR (1.2.3.4/5)

##### `ports` [optional]
A list of destination ports that are integers between 1 and 65535.
> `ports` or `port_range` must be provided for `tcp` and `udp`.
> It is an error when both are provided.

##### `port_range` [optional]
- `start` [required] the start of the range as an integer between 1 and 65535
- `end` [required] the end of the range as an integer between 1 and 65535

> `ports` or `port_range` must be provided for protocol `tcp` and `udp`.
> It is an error when both are provided.

##### `icmp_info` [optional]
- `type` [required] will be an integer between 0 and 255
- `code` [required] will be an integer

> `icmp_info` is required for protocol `icmp`.
> It is an error when provided for other protocols.

##### `log` [optional]
Enable logging of the rule
> `log` is optional for `tcp` and `all`.
> It is an error to provide `log` as true when protocol is `udp` or `icmp`.

> Define all rules with `log` enabled at the end of your `egress_rules` to guarantee logging.

##### Examples
***
`ALL`
```
{
    "protocol": "all",
    "destinations": ["1.2.3.4"],
    "log": true
}
```
***
`TCP`
```
{
    "protocol": "tcp",
    "destinations": ["1.2.3.4-2.3.4.5"],
    "ports": [80, 443],
    "log": true
}
```
***
`UDP`
```
{
    "protocol": "udp",
    "destinations": ["1.2.3.4/4"],
    "port_range": {
        "start": 8000,
        "end": 8085
    }
}
```
***
`ICMP`
```
{
    "protocol": "icmp",
    "destinations": ["1.2.3.4", "2.3.4.5/6"],
    "icmp_info": {
        "type": 1,
        "code": 40
    }
}
```
***
### Logging

Diego uses [loggregator](https://github.com/cloudfoundry/loggregator) to emit logs generated by container processes to the user.

#### `log_guid` [optional]

`log_guid` controls the loggregator guid associated with logs coming from LRP processes.  One typically sets the `log_guid` to the `process_guid` though this is not strictly necessary.

#### `log_source` [optional]

`log_source` is an identifier emitted with each log line.  Individual `RunAction`s can override the `log_source`.  This allows a consumer of the log stream to distinguish between the logs of different processes.

#### `metrics_guid` [optional]

`metrics_guid` controls the loggregator guid associated with metrics coming from LRP processes.

#### Attaching Arbitrary Metadata

#### `annotation` [optional]

Diego allows arbitrary annotations to be attached to a DesiredLRP.  The annotation must not exceed 10 kilobytes in size.
