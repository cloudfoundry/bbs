##### `EnvironmentVariables` [optional]

Diego supports the notion of container-level environment variables.  All processes that run in the container will inherit these environment variables.

For more details on the environment variables provided to processes in the container, read [Container Runtime Environment](environment.md).

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

##### `VolumeMounts` [optional]

TODO

##### `SecurityGroupRule`

List of firewall rules applied to the Task container. If traffic originating inside the container has a destination matching one of the rules, it is allowed egress. For example,

Defining security networking rules uses the SecurityGroupRule model and can define the protocol, destinations IPs, and ports.

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
