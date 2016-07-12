# Internal Tasks API Reference

This reference does not cover the protobuf payload supplied to each endpoint.

For detailed information on the structs and types listed see [models documentation](https://godoc.org/code.cloudfoundry.org/bbs/models)

# Internal Tasks APIs

## StartTask
```go
{Path: "/v1/tasks/start", Method: "POST", Name: StartTaskRoute},
```

### BBS API Endpoint
Post a StartTaskRequest to "/v1/tasks/start"

### Golang Client API
```go
func (c *client) StartTask(logger lager.Logger, taskGuid string, cellID string) (bool, error)
```

#### Input
* `logger lager.Logger`
  * The logging sink
* `taskGuid string`
  * The task guid
* `cellID string`
  * Cell ID in which the tasks is started

#### Output
* `bool`
  * `true` if task should be started, `false` if not
* `error`
  * Non-nil if error occurred

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
```go
{Path: "/v1/tasks/fail", Method: "POST", Name: FailTaskRoute},
```

### BBS API Endpoint
Post a FailTaskRequest to "/v1/tasks/fail"

### Golang Client API
```go
func (c *client) FailTask(logger lager.Logger, taskGuid, failureReason string) error
```

#### Input
* `logger lager.Logger`
  * The logging sink
* `taskGuid string`
  * The task guid
* `failureReason string`
  * Reason for which the task has failed

#### Output
* `error`
  * Non-nil if error occurred

#### Example
```go
client := bbs.NewClient(url)
err := client.FailTask(logger, "task-guid", "not enough resources")
if err != nil {
    log.Printf("could not fail task: " + err.Error())
}
```

## CompleteTask
```go
{Path: "/v1/tasks/complete", Method: "POST", Name: CompleteTaskRoute},
```

### BBS API Endpoint
Post a CompleteTaskRequest to "/v1/tasks/fail"

### Golang Client API
```go
func (c *client) CompleteTask(logger lager.Logger, taskGuid, cellId string, failed bool, failureReason, result string) error
```

#### Input
* `logger lager.Logger`
  * The logging sink
* `taskGuid string`
  * The task guid
* `cellId string`
  * Cell ID in which the task ran
* `failed bool`
  * Whether the task failed or not
* `failureReason string`
  * Reason for which the task has failed
* `result string`
  * Task result in text format

#### Output
* `error`
  * Non-nil if error occurred

#### Example
```go
client := bbs.NewClient(url)
err = client.CompleteTask(logger, "task-guid", "cell-1", false, "", "result")
if err != nil {
    log.Printf("could not complete task: " + err.Error())
}
```

