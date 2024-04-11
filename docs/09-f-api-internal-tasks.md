---
title: Tasks Internal API
expires_at: never
tags: [diego-release, bbs]
---

# Tasks Internal API

Instead, it illustrates calls to the API via the Golang `bbs.Client` interface.
Each method on that `Client` interface takes a `lager.Logger` as the first argument to log errors generated within the client.
This first `Logger` argument will not be duplicated on the descriptions of the method arguments.

For detailed information on the types referred to below, see the [godoc documentation for the BBS models](https://pkg.go.dev/code.cloudfoundry.org/bbs/models).

# Internal Tasks APIs

## StartTask

### BBS API Endpoint

POST a [StartTaskRequest](https://pkg.go.dev/code.cloudfoundry.org/bbs/models#StartTaskRequest)
to `/v1/tasks/start`
and receive a [StartTaskResponse](https://pkg.go.dev/code.cloudfoundry.org/bbs/models#StartTaskResponse).

### Golang Client API

```go
StartTask(logger lager.Logger, taskGuid string, cellID string) (bool, error)
```

#### Input

* `taskGuid string`: The GUID of the Task to start.
* `cellID string`: ID of the cell intending to start the Task.

#### Output

* `bool`: `true` if the Task should be started, `false` if not.
* `error`: Non-nil if error occurred.

#### Example

```go
client := bbs.NewClient(url)
shouldStart, err := client.StartTask(logger, "task-guid", "cell-1")
if err != nil {
    log.Printf("failed to start task: " + err.Error())
}
if shouldStart {
  log.Print("task should be started")
} else {
  log.Print("task should NOT be started")
}
```

## FailTask

**Deprecated** in favor of [CompleteTask](09-f-api-internal-tasks.md#completetask) and [CancelTask](09-a-api-external-tasks.md#canceltask).

### BBS API Endpoint

POST a [FailTaskRequest](https://pkg.go.dev/code.cloudfoundry.org/bbs/models#FailTaskRequest)
to `/v1/tasks/fail`
and receive a [TaskLifecycleResponse](https://pkg.go.dev/code.cloudfoundry.org/bbs/models#TaskLifecycleResponse).

### Golang Client API

```go
FailTask(logger lager.Logger, taskGuid, failureReason string) error
```

#### Input

* `taskGuid string`: The GUID of the Task to fail.
* `failureReason string`: Reason why the Task failed.

#### Output

* `error`:  Non-nil if an error occurred.

#### Example

```go
client := bbs.NewClient(url)
err := client.FailTask(logger, "task-guid", "not enough resources")
if err != nil {
    log.Printf("could not fail task: " + err.Error())
}
```

## RejectTask

### BBS API Endpoint

POST a [RejectTaskRequest](https://pkg.go.dev/code.cloudfoundry.org/bbs/models#RejectTaskRequest)
to `/v1/tasks/reject`
and receive a [TaskLifecycleResponse](https://pkg.go.dev/code.cloudfoundry.org/bbs/models#TaskLifecycleResponse).

### Golang Client API

```go
RejectTask(logger lager.Logger, taskGuid, rejectionReason string) error
```

#### Input

* `taskGuid string`: The GUID of the Task to fail.
* `rejectReason string`: Reason why the Task was rejected.

#### Output

* `error`:  Non-nil if an error occurred.

#### Example

```go
client := bbs.NewClient(url)
err := client.RejectTask(logger, "task-guid", "not enough resources")
if err != nil {
    log.Printf("could not reject task: " + err.Error())
}
```

## CompleteTask

### BBS API Endpoint
POST a [CompleteTaskRequest](https://pkg.go.dev/code.cloudfoundry.org/bbs/models#CompleteTaskRequest)
to `/v1/tasks/fail`
and receive a [TaskLifecycleResponse](https://pkg.go.dev/code.cloudfoundry.org/bbs/models#TaskLifecycleResponse).

### Golang Client API

```go
CompleteTask(logger lager.Logger, taskGuid, cellId string, failed bool, failureReason, result string) error
```

#### Input

* `taskGuid string`: The GUID of the Task to complete.
* `cellID string`: ID of the cell intending to complete the Task.
* `failed bool`: Whether the Task failed.
* `failureReason string`: If Task failed, the reason why the Task failed.
* `result string`: If Task succeeded, result of the Task.

#### Output

* `error`:  Non-nil if an error occurred.

#### Example

```go
client := bbs.NewClient(url)
err = client.CompleteTask(logger, "task-guid", "cell-1", false, "", "result")
if err != nil {
    log.Printf("could not complete task: " + err.Error())
}
```
