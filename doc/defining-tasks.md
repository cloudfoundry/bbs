## Defining Tasks

This document explains the fields available when defining a new Task. For a higher-level overview of the Diego Task API, see the [Tasks Overview](tasks.md).

```go
client := bbs.NewClient(url)
err := client.DesireTask(
  "task-guid", // 'guid' parameter
  "domain",    // 'domain' parameter
  &models.TaskDefinition{
    RootFs: "docker:///docker.com/docker",
    EnvironmentVariables: []*models.EnvironmentVariable{
      {
        Name:  "FOO",
        Value: "BAR",
      },
    },
    CachedDependencies: []*models.CachedDependency{
      {
        Name: "app bits",
        From: "https://blobstore.com/bits/app-bits",
        To: "/usr/local/app",
        CacheKey: "cache-key",
        LogSource: "log-source",
        ChecksumAlgorithm: "md5",
        ChecksumValue: "the-checksum",
      },
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
    CompletionCallbackUrl: "http://36.195.164.128:8080",
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

Diego clients must provide each Task with a unique Task identifier. Use this identifier to refer to the Task later.

- It is an error to create a Task with a guid matching that of an existing Task.
- The `guid` must include only the characters `a-z`, `A-Z`, `0-9`, `_` and `-`.
- The `guid` must not be empty.


#### `domain` [required]

Diego clients must label their Tasks with a domain. These domains partition the Tasks into logical groups, which clients may retrieve via the BBS API. Task domains are purely organizational (for example, for enabling multiple clients to use Diego without accidentally interfering with each other) and do not affect the Task's placement or lifecycle.

- It is an error to provide an empty `domain`.


### Task Definition Fields

#### Container Contents and Environment

##### `RootFs` [required]

The `RootFs` field specifies the root filesystem to use inside the container. One class of root filesystems are the `preloaded` root filesystems, which are directories colocated on the Diego Cells and registered with their cell reps. Clients specify a preloaded root filesystem in the form:

```go
RootFs: "preloaded:ROOTFS-NAME"
```

Cloud Foundry buildpack-based apps use the `cflinuxfs2` preloaded filesystem, built to work with Cloud Foundry buildpacks:

```go
RootFs: "preloaded:cflinuxfs2"
```

Clients may also provide a root filesystem based on a Docker image:

```go
RootFs: "docker:///docker-org/docker-image#docker-tag"
```

To pull the image from a different registry than Docker Hub, specify it as the host in the URI string, e.g.:

```go
RootFs: "docker://index.myregistry.gov/docker-org/docker-image#docker-tag"
```

##### `EnvironmentVariables` [optional]

Clients may define environment variables at the container level, which all processes running in the container will receive. For example:

```go
EnvironmentVariables: []*models.EnvironmentVariable{
  {
    Name:  "FOO",
    Value: "BAR",
  },
  {
    Name:  "LANG",
    Value: "en_US.UTF-8",
  },
}
```

For more details on the environment variables provided to processes in the container, see the section on the [Container Runtime Environment](environment.md)


##### `CachedDependencies` [optional]

List of dependencies to cache on the Diego Cell and then to bind-mount into the container at the specified location. For example:

```go
CachedDependencies: []*models.CachedDependency{
  {
    Name: "app bits",
    From: "https://blobstore.com/bits/app-bits",
    To: "/usr/local/app",
    CacheKey: "cache-key",
    LogSource: "log-source",
    ChecksumAlgorithm: "md5",
    ChecksumValue: "the-checksum",
  },
},
```

The `ChecksumAlgorithm` and `ChecksumValue` are optional and used to validate the downloaded binary.  They must be used together.

##### `TrustedSystemCertificatesPath` [optional]

An absolute path inside the container's filesystem where trusted system certificates will be provided if an operator has specified them.



#### Container Limits

##### `CpuWeight` [optional]

To control the CPU shares provided to a container, set `CpuWeight`. This must be a positive number in the range `1-100`. The `CpuWeight` enforces a relative fair share of the CPU among containers per unit time. To explain, suppose that conatainer A and container B each runs a busy process that attempts to consume as much CPU as possible.

- If A and B each has `CpuWeight: 100`, their processes will receive approximately equal amounts of CPU time.
- If A has `CpuWeight: 25` and B has `CpuWeight: 75`, A's process will receive about one quarter of the CPU time, and B's process will receive about three quarters of it.


##### `DiskMb` [optional]

A disk quota in mebibytes applied to the container. Data written on top of the container's root filesystem counts against this quota. If it is exceeeded, writes will fail, but the container runtime will not kill processes in the container directly.

- The `DiskMb` value must be an integer greater than or equal to 0.
- If set to 0, no disk quota is applied to the container.


##### `MemoryMb` [optional]

A memory limit in mebibytes applied to the container.  If the total memory consumption by all processs running in the container exceeds this value, the container will be destroyed.

- The `MemoryMb` value must be an integer greater than or equal to 0.
- If set to 0, no memory quota is applied to the container.


##### `Privileged` [optional]

- If false, Diego will create a container that is in a user namespace.  Processes that run as root will actually be root only within the user namespace and will not have administrative privileges on the host system.
- If true, Diego creates a container without a user namespace, so that container root corresponds to root on the host system.


#### Actions

##### `Action` [required]

Encodes the action to execute when running the Task.  For more details, see the section on [Actions](actions.md).


#### Task Completion and Output

When the `Action` on a Task finishes, the Task is marked as `COMPLETED`.

##### `ResultFile` [optional]

If specified on a Task, Diego retrieves the contents of this file from the container when the Task completes successfully. The retrieved contents are made available in the `Result` field of the `TaskResponse` (see [below](#retrieving-tasks)).

- Diego only returns the first 10 kilobytes of the `ResultFile`.  If you need to communicate back larger datasets, consider using an `UploadAction` to upload the result file to another service.


##### `CompletionCallbackUrl` [optional]

Diego clients have several ways to learn that a Task has `COMPLETED`: they can poll the Task, subscribe to the Task event stream, or register a callback.

If a `CompletionCallbackUrl` is provided, Diego will send a `POST` request to the provided URL when the Task completes.  The body of the `POST` will include the `TaskResponse` (see [below](#retrieving-tasks)).

- Almost any response from the callback will resolve the Task, thereby removing it from the BBS.
- If the callback responds with status code '503 Service Unavailable' or '504 Gateway Timeout', however, Diego will immediately retry the callback up to 3 times.

- If these status codes persist, if the callback times out, or if a connection cannot be established, Diego will try again after a short period of time, typically 30 seconds.
- After about 2 minutes without a successful response from the callback URL, Diego will give up on the task and delete it.

#### Networking

By default network access for any container is limited but some tasks may need specific network access and that can be setup using `egress_rules` field.


##### `EgressRules` [optional]

List of firewall rules applied to the Task container. If traffic originating inside the container has a destination matching one of the rules, it is allowed egress. For example,

```go
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
}
```

This list of rules allows all outgoing TCP traffic bound for ports 1 though 1024 and UDP traffic to subnet 8.8.0.0/16 on port 53. Syslog messages are emitted for new connections matching the TCP rule.

###### `Protocol` [required]

The protocol type of the rule can be one of the following values: `tcp`, `udp`,`icmp`, or `all`.

###### `Destinations` [required]

List of string representing a single IPv4 address (`1.2.3.4`), a range of IPv4 addresses (`1.2.3.4-2.3.4.5`), or an IPv4 subnet in CIDR notation (`1.2.3.4/24`).


###### `Ports` and `PortRange` [optional]

The `Ports` field is a list of integers between 1 and 65535 that correspond to destination ports.
The `PortRange` field is a struct with a `Start` field and an `End` field, both integers between 1 and 65535. These values are required and signify the start and end of the port range, inclusive.

- Either `Ports` or `PortRange` must be provided for protocol `tcp` and `udp`.
- It is an error to provide both.

###### `IcmpInfo` [optional]

The `IcmpInfo` field stores two fields with parameters that pertain to ICMP traffic:

- `Type` [required]: integer between 0 and 255
- `Code` [required]: integer

```go
rule := &SecurityGroupRule{
  Protocol: "icmp",
  IcmpInfo: &ICMPInfo{Type: 8, Code: 0}
}
```

- `IcmpInfo` is required for protocol `icmp`.
- It is an error to provide for other protocols.

###### `Log` [optional]

If true, the system will log new outgoing connections that match the rule.

- `Log` is optional for `tcp` and `all`.
- It is an error to set `Log` to true when the protocol is `udp` or `icmp`.
- To ensure that they apply first, put all rules with `Log` set to true at the **end** of the rule list.

##### Examples of Egress rules

---

Protocol `all`:

```go
all := &SecurityGroupRule{
    Protocol: "all",
    Destinations: []string{"1.2.3.4"},
    Log: true,
}
```

---

Protocol `tcp`:

```go
tcp := &SecurityGroupRule{
    Protocol: "tcp",
    Destinations: []string{"1.2.3.4-2.3.4.5"},
    Ports: []int[80, 443],
    Log: true,
}
```

---

Protocol `udp`:

```go
udp := &SecurityGroupRule{
    Protocol: "udp",
    Destinations: []string{"1.2.3.4/8"},
    PortRange: {
        Start: 8000,
        End: 8085,
    },
}
```

---

Protocol `icmp`:

```go
icmp := &SecurityGroupRule{
    Protocol: "icmp",
    Destinations: []string{"1.2.3.4", "2.3.4.5/6"},
    IcmpInfo: {
        Type: 1,
        Code: 40,
    },
}
```

---

#### Logging

Diego emits container metrics and logs generated by container processes to the [Loggregator](https://github.com/cloudfoundry/loggregator) system, in the form of [dropsonde log messages](https://github.com/cloudfoundry/dropsonde-protocol/blob/master/events/log.proto) and [container metrics](https://github.com/cloudfoundry/dropsonde-protocol/blob/master/events/metric.proto).

##### `LogGuid` [optional]

The `LogGuid` field sets the `AppId` on log messages coming from the Task.


##### `LogSource` [optional]

The `LogSource` field sets the default `SourceType` on the log messages. Individual Actions on the Task may override this field, so that different actions may be distinguished by the `SourceType` values on log messages.


##### `MetricsGuid` [optional]

The `MetricsGuid` field sets the `ApplicationId` on container metris coming from the Task.


#### Storing Arbitrary Metadata

##### `Annotation` [optional]

Diego allows arbitrary annotations to be attached to a Task.  The annotation may not exceed 10 kilobytes in size.


[back](README.md)
