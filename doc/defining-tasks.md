## Defining Tasks

This document provides an overview of the fields needed to define a new task. For a higher level overview of the Diego task api, see the [Tasks doc](tasks.md).

```
client := bbs.NewClient(url)
err := client.DesireTask(
          "task-guid",
          "domain",
          &models.TaskDefinition{
              RootFs: "docker:///docker.com/docker",
              EnvironmentVariables: []*models.EnvironmentVariable{
                {
                  Name:  "FOO",
                  Value: "BAR",
                },
              },
              CachedDependencies: []*models.CachedDependency{
                {Name: "app bits", From: "blobstore.com/bits/app-bits", To: "/usr/local/app", CacheKey: "cache-key", LogSource: "log-source"},
              },
              Action: models.WrapAction(&models.RunAction{
                User:           "user",
                Path:           "echo",
                Args:           []string{"hello world"},
                ResourceLimits: &models.ResourceLimits{},
              }),
              MemoryMb:    256,
              DiskMb:      1024,
              CpuWeight:   42,
              Privileged:  true,
              LogGuid:     "123",
              LogSource:   "APP",
              MetricsGuid: "456",
              CompletionCallbackUrl: "36.195.164.128:8080",
              ResultFile:  "some-file.txt",
              EgressRules: []*models.SecurityGroupRule{
                {
                  Protocol:     "tcp",
                  Destinations: []string{"0.0.0.0/0"},
                  PortRange: &models.PortRange{
                    Start: 1,
                    End:   1024,
                  },
                  Log: true,
                },
                {
                  Protocol:     "udp",
                  Destinations: []string{"8.8.0.0/16"},
                  Ports:        []uint32{53},
                },
              },
              Annotation:                    "place any label/note/thing here",
              TrustedSystemCertificatesPath: "/etc/somepath",
          }
)
```

### Task Identifiers

#### `guid` [required]

It is up to the consumer of Diego to provide a *globally unique* task guid.  To subsequently fetch the Task you refer to it by its guid.

- It is an error to attempt to create a Task whose task guid matches that of an existing Task.
- The `guid` must only include the characters `a-z`, `A-Z`, `0-9`, `_` and `-`.
- The `guid` must not be empty

#### `domain` [required]

The consumer of Diego may organize their Tasks into groupings called Domains.  These are purely organizational (e.g. for enabling multiple consumers to use Diego without colliding) and have no implications on the Task's placement or lifecycle.  It is possible to fetch all Tasks in a given Domain.

- It is an error to provide an empty `domain`.

### What's in a Task Definition

#### Container Contents and Environment

##### `RootFs` [required]

The `RootFs` field specifies the root filesystem to mount into the container.  Diego can be configured with a set of *preloaded* RootFses.  These are named root filesystems that are already on the Diego Cells.

Preloaded root filesystems look like:

```
RootFs: "preloaded:ROOTFS-NAME"
```

Diego ships with a root filesystem:
```
RootFs: "preloaded:cflinuxfs2"
```
these are built to work with the Cloud Foundry buildpacks.

It is possible to provide a custom root filesystem by specifying a Docker image for `rootfs`:

```
RootFs: "docker:///docker-org/docker-image#docker-tag"
```

To pull the image from a different registry than the default (Docker Hub), specify it as the host in the URI string, e.g.:

```
RootFs: "docker://index.myregistry.gov/docker-org/docker-image#docker-tag"
```

##### `EnvironmentVariables` [optional]

Diego supports the notion of container-level environment variables.  All processes that run in the container will inherit these environment variables.

EG:
```
environmentVariables := []*models.EnvironmentVariable{
  {
    Name:  "FOO",
    Value: "BAR",
  },
}
```

For more details on the environment variables provided to processes in the container, see [Container Runtime Environment](environment.md)

#### Container Limits

##### `CpuWeight` [optional]

To control the CPU shares provided to a container, set `CpuWeight`.  This must be a positive number in the range `1-100`.  The `CpuWeight` enforces a relative fair share of the CPU among containers.  It's best explained with examples.  Consider the following scenarios (we shall assume that each container is running a busy process that is attempting to consume as many CPU resources as possible):

- Two containers, with equal values of `cpu_weight`: both containers will receive equal shares of CPU time.
- Two containers, one with `CpuWeight=50` the other with `CpuWeight=100`: the later will get (roughly) 2/3 of the CPU time, the former 1/3.

##### `DiskMb` [optional]

A disk quota applied to the entire container.  Any data written on top of the RootFs counts against the Disk Quota.  Processes that attempt to exceed this limit will not be allowed to write to disk.

- `DiskMb` must be an integer >= 0
- If set to 0 no disk constraints are applied to the container
- The units are megabytes

##### `MemoryMb:` [optional]

A memory limit applied to the entire container.  If the aggregate memory consumption by all processs running in the container exceeds this value, the container will be destroyed.

- `MemoryMb:` must be an integer >= 0
- If set to 0 no memory constraints are applied to the container
- The units are megabytes

##### `Privileged` [optional]

If false, Diego will create a container that is in a user namespace.  Processes that succesfully obtain escalated privileges (i.e. root access) will actually only be root within the user namespace and will not be able to maliciously modify the host VM.  If true, Diego creates a container with no user namespace -- escalating to root gives the user *real* root access.

#### Actions

##### `Action` [required]

Encodes the action to run when running the Task.  For more details see [actions](actions.md)

#### Task Completion and Output

When the `Action` on a Task terminates the Task is marked as `COMPLETED`.

##### `ResultFile` [optional]

When a Task completes succesfully Diego can fetch and return the contents of a file in the container.  This is made available in the `result` field of the `TaskResponse` (see [below](#retrieving-tasks)).

To do this, set `ResultFile` to a valid absolute path in the container.

- Diego only returns the first 10KB of the `ResultFile`.  If you need to communicate back larger datasets, consider using an `UploadAction` to upload the result file to a blob store.

##### `CompletionCallbackUrl` [optional]

Consumers of Diego have two options to learn that a Task has `COMPLETED`: they can either poll the action or register a callback.

If a `CompletionCallbackUrl` is provided Diego will `POST` to the provided URL as soon as the Task completes.  The body of the `POST` will include the `TaskResponse` (see [below](#retrieving-tasks)).

- Any response from the callback (be it success or failure) will resolve the Task (removing it from Diego).
- However, if the callback responds with `503` or `504` Diego will immediately retry the callback up to 3 times.  If the `503/504` status persists Diego will try again after a period of time (typically within ~30 seconds).
- If the callback times out or a connection cannot be established, Diego will try again after a period of time (typically within ~30 seconds).
- Diego will eventually (after ~2 minutes) give up on the Task if the callback does not respond succesfully.

#### Networking
By default network access for any container is limited but some tasks might need specific network access and that can be setup using `egress_rules` field.

Rules are evaluated in reverse order of their position, i.e., the last one takes precedence.

##### `EgressRules` [optional]
`EgressRules` are a list of egress firewall rules that are applied to a container running in Diego

```
egressRules := []*models.SecurityGroupRule{
  {
    Protocol:     "tcp",
    Destinations: []string{"0.0.0.0/0"},
    PortRange: &models.PortRange{
      Start: 1,
      End:   1024,
    },
    Log: true,
  },
  {
    Protocol:     "udp",
    Destinations: []string{"8.8.0.0/16"},
    Ports:        []uint32{53},
  },
}
```
###### `Protocol` [required]
The protocol of the rule that can be one of the following `tcp`, `udp`,`icmp`, `all`.

###### `destinations` [required]
The destinations of the rule that is a list of either an IP Address (1.2.3.4) or an IP range (1.2.3.4-2.3.4.5) or a CIDR (1.2.3.4/5)

###### `Ports` [optional]
A list of destination ports that are integers between 1 and 65535.

> `Ports` or `PortRange` must be provided for `tcp` and `udp`.
> It is an error when both are provided.

###### `PortRange` [optional]
- `Start` [required] the start of the range as an integer between 1 and 65535
- `End` [required] the end of the range as an integer between 1 and 65535

> `Ports` or `PortRange` must be provided for protocol `tcp` and `udp`.
> It is an error when both are provided.

###### `IcmpInfo` [optional]
- `Type` [required] will be an integer between 0 and 255
- `Code` [required] will be an integer

```
rule := &SecurityGroupRule{
  Protocol: "icmp",
  IcmpInfo: &ICMPInfo{Type: 8, Code: 0}
}
```

> `IcmpInfo` is required for protocol `icmp`.
> It is an error when provided for other protocols.

###### `Log` [optional]
Enable logging of the rule
> `Log` is optional for `tcp` and `all`.
> It is an error to provide `Log` as true when protocol is `udp` or `icmp`.

> Define all rules with `Log` enabled at the end of your `[]*SecurityGroupRule` to guarantee logging.

##### Examples
***
`ALL`
```
all := &SecurityGroupRule{
    Protocol: "all",
    Destinations: ["1.2.3.4"],rep/conversion_helpers.go
    Log: true,
}
```
***
`TCP`
```
tcp := &SecurityGroupRule{
    Protocol: "tcp",
    Destinations: ["1.2.3.4-2.3.4.5"],
    Ports: [80, 443],
    Log: true,
}
```
***
`UDP`
```
udp := &SecurityGroupRule{
    Protocol: "udp",
    Destinations: ["1.2.3.4/4"],
    PortRange: {
        Start: 8000,
        End: 8085,
    },
}
```
***
`ICMP`
```
icmp := &SecurityGroupRule{
    Protocol: "icmp",
    Destinations: ["1.2.3.4", "2.3.4.5/6"],
    IcmpInfo: {
        Type: 1,
        Code: 40,
    },
}
```
***
#### Logging

Diego uses [doppler](https://github.com/cloudfoundry/loggregator) to emit logs generated by container processes to the user.

##### `LogGuid` [optional]

`LogGuid` controls the doppler guid associated with logs coming from Task processes.  One typically sets the `LogGuid` to the task's `guid` though this is not strictly necessary.

##### `LogSource` [optional]

`LogSource` is an identifier emitted with each log line.  Individual `RunAction`s can override the `LogSource`.  This allows a consumer of the log stream to distinguish between the logs of different processes.

##### `MetricsGuid` [optional]

`LogGuid` controls the doppler guid associated with metrics coming from Task processes.  One typically sets the `MetricsGuid` to the task's `guid` though this is not strictly necessary.


#### Attaching Arbitrary Metadata

##### `Annotation` [optional]

Diego allows arbitrary annotations to be attached to a Task.  The annotation must not exceed 10 kilobytes in size.

##### `TrustedSystemCertificatesPath` [optional]

This is an absolute path inside the container's filesystem where system-wide tls certificates will be installed if an operator has specified them.

[back](README.md)
