# Tasks

Diego can run one-off work in the form of Tasks.  When a Task is submitted Diego allocates resources on a Cell, runs the Task, and then reports on the Task's results.  Tasks are guaranteed to run at most once.

## The Task API

We recommend interacting with Diego's task functionality through the BBS' ExternalTaskClient. The RPC calls exposed to consumers are specifically documented [here](https://godoc.org/github.com/cloudfoundry-incubator/bbs#ExternalTaskClient).

## The Task lifecycle

Tasks in Diego undergo a simple lifecycle encoded in the Tasks's state:

- When first created a Task enters the `PENDING` state.
- When succesfully allocated to a Diego Cell the Task will enter the `CLAIMED` state.  At this point the Task's `CellId` will be populated.
- When the Cell begins to create the container and run the Task action, the Task enters the `RUNNING` state.
- When the Task completes, the Cell annotates the `Task` with `Failed`, `FailureReason`, and `Result`, and puts the Task in the `COMPLETED` state.

At this point it is up to the consumer of Diego to acknowledge and resolve the completed Task.  This can either be done via a completion callback (described [above](#completion_callback_url-optional)) or by the Task.  When the Task is being resolved it first enters the `RESOLVING` state and is ultimately removed from Diego.

Diego will automatically prune Tasks that remain unresolved after 2 minutes.

The `RESOLVING` state exists to ensure that the `CompletionCallbackUrl` is initially called at most once per Task.

There are a variety of timeouts associated with the `PENDING` and `CLAIMED` states.  It is possible for a Task to jump directly from `PENDING` or `CLAIMED` to `COMPLETED` (and `Failed`) if any of these timeouts expire.  If you would like to impose a time limit on how long the Task is allowed to run you can use a `TimeoutAction`.

## Defining Tasks

When submitting a task, a valid `guid`, `domain`, and `TaskDefinition` should be provided to [a Client's DesireTask method](https://github.com/cloudfoundry-incubator/bbs/blob/master/client.go#L121). Here's an example of such a call:

```go
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
              Annotation:                    `[{"anything": "you want!"}]... dude`,
              LegacyDownloadUser:            "legacy-jim",
              TrustedSystemCertificatesPath: "/etc/somepath",
          }
)
```

To see what each of these fields do, please reference the doc on [defining tasks](defining-tasks.md)

## Checking up on a Task

The `ExternalTaskClient` can be used to retrieve a task from the bbs. The fields on the returned Task struct that aren't part of the Task's definition represent its status in the system. The returned task has the following additional attributes:

### `State`

State defines where in the [lifecycle](#the_task_lifecycle) the task is.

### `CellId`

`CellId` identifies which of Diego's Cells the task has been delegated to run on.

### `CreatedAt`, `UpdatedAt`, and `FirstCompletedAt`

Timestamps in standard golang time.Time nanoseconds. Perhaps most importantly, the `FirstCompletedAt` timestamp is referenced when a task is pruned 2 minutes after it completes.

## Receiving The Results of a Task

If a `CompletionCallbackUrl` is registered with the original task definition a `TaskCallbackResponse` will be sent back as json to the specified URL when the task is completed.

The `TaskCallbackResponse` should look something like the following:
```go
taskResponse, _ := json.Marshal(&models.TaskCallbackResponse {
  TaskGuid: "some-guid",
  Failed: false,
  FailureReason: "some failure",
  Result: "first 10KB of ResultFile",
  Annotation: "arbitrary",
  CreatedAt: int64(time.Now()),
})
```

Let's describe each of these fields in turn.

#### `Failed`

Once a Task enters the `COMPLETED` state, `Failed` will be a boolean indicating whether the Task completed succesfully or unsuccesfully.

#### `FailureReason`

If `Failed` is `true`, `FailureReason` will be a short string indicating why the Task failed.  Sometimes, in the case of a `RunAction` that has failed this will simply read (e.g.) `exit status 1`.  To debug the Task you will need to fetch the logs from doppler.

#### `Result`

If `ResultFile` was specified and the Task has completed succesfully, `Result` will include the first 10KB of the `ResultFile`.

#### `Annotation`

This is the arbitrary string that was specified in the TaskDefinition.

[back](README.md)
