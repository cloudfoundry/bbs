# Task Examples

## Desiring a Task

```go
client := bbs.NewClient("http://10.244.16.2:8889")
err = client.DesireTask(
  "some-guid",
  "some-domain",
  &models.TaskDefinition{
    RootFs: "docker:///busybox",
    Action: models.WrapAction(&models.RunAction{
      User:           "root",
      Path:           "sh",
      Args:           []string{"-c", "echo hello world > result-file.txt"},
      ResourceLimits: &models.ResourceLimits{},
    }),
    CompletionCallbackUrl: "http://10.244.16.6:6660",
    ResultFile:            "result-file.txt",
  },
)
```

## Polling for Task Info

```go
for {
  task, err := client.TaskByGuid("some-guid")
  if err != nil {
    log.Printf("failed to fetch task!")
    panic(err)
  }
  if task.State == models.Task_Resolving {
    log.Printf("here's the result from you polled task:\n %s\n\n", task.Result)
    break
  }
  time.Sleep(time.Second)
}
```

## Recieving a TaskCallbackResponse

To recieve the TaskCallbackResponse, we're going to start up a classic http server.
```go
func taskCallbackHandler(w http.ResponseWriter, r *http.Request) {
	var taskResponse models.TaskCallbackResponse
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("failed to read task response")
		panic(err)
	}
	err = json.Unmarshal(data, &taskResponse)
	if err != nil {
		log.Printf("failed to unmarshal json")
		panic(err)
	}
	log.Printf("here's the result from your TaskCallbackResponse:\n %s\n\n", taskResponse.Result)
}

http.HandleFunc("/", taskCallbackHandler)
go http.ListenAndServe("8080", nil)
```
With this running, if the above task was desired, it would run on Diego, echo 'hello world' to it's ResultFile and complete. Diego would then populate the Result field of the TaskCallbackResponse with 'hello world' and post it as json to the CompletionCallbackUrl. The server we created here would then read that and print 'hello world' as part of the task response.

## Cancelling a Task

```go
client := bbs.NewClient(url)
err := client.CancelTask("some-guid")
```
