# Events

BBS can optionally emit events when any desired or actual lrp is created,
updated or removed. The following sections provide details on how to subscribe
to those events as well as the type of events supported by the bbs

## Subscribing to events

You can use the `SubscribeToEvents(logger lager.Logger) (events.EventSource,
error)` client method to subscribe to events. For example

``` go
client := bbs.NewClient(url)
es, err := client.SubscribeToEvents(logger)
if err != nil {
    log.Printf("failed to subscribe to events: " + err.Error())
}
```

You can then loop through the events by calling
[Next](https://godoc.org/code.cloudfoundry.org/bbs/events#EventSource) in a
loop, for example:

``` go
event, err := eventSource.Next()
if err != nil {
    log.Printf("failed to get next event: " + err.Error())
}
log.Printf("received event: %#v", event)
```

To access the event field values, you must convert the event to the right
type. You can use the `EventType` method to determine the type of the event,
for example:

``` go
if event.EventType() == models.EventTypeActualLRPCrashed {
  crashEvent := event.(*models.ActualLRPCrashedEvent)
  log.Printf("lrp has crashed. err: %s", crashEvent.CrashReason)
}
```

Following types of events are emitted when changes to desired LRP and actual LRP are done:

## Desired LRP events

### Desire LRP create event

When a new desired LRP is created a
[DesiredLRPCreatedEvent](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPCreatedEvent)
is emitted. The field value of `DesiredLrp` will have information about the
desired lrp that was just created.

### Desired LRP change event

When a desired LRP changes a
[DesiredLRPChangedEvent](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPChangedEvent)
is emitted. The field value of `Before` and `After` have information about the
desired lrp state before and after the change.

### Desired LRP remove event

When a desired LRP is removed a
[DesiredLRPRemovedEvent](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPRemovedEvent)
is emitted. The field value of `DesiredLrp` will have information about the
desired lrp that was just removed.

## Actual LRP events

### Actual LRP create event

When a new actual LRP is created a
[ActualLRPCreatedEvent](https://godoc.org/code.cloudfoundry.org/bbs/models#ActualLRPRemovedEvent)
is emitted. The field value of `ActualLrpGroup` will contain more information
about the actual lrp


### Actual LRP change event

When a actual LRP changes a
[ActualLRPChangedEvent](https://godoc.org/code.cloudfoundry.org/bbs/models#ActualLRPChangedEvent)
is emitted. The field value of `Before` and `After` have information about the
actual lrp state before and after the change.

### Actual LRP remove event

When a actual LRP is removed a
[ActualLRPRemovedEvent](https://godoc.org/code.cloudfoundry.org/bbs/models#ActualLRPRemovedEvent)
is emitted. The field value of `ActualLrp` will have information about the
actual lrp that was just removed.

### Actual LRP crash event

When a actual LRP crashes a
[ActualLRPCrashedEvent](https://godoc.org/code.cloudfoundry.org/bbs/models#ActualLRPCrashedEvent)
is emitted. The event will have the following field values:

1. `ActualLRPKey`: the lrp key
1. `ActualLRPInstanceKey`: the instance key
1. `CrashCount`: the number of times the actual lrp has crashed
1. `CrashReason`: the last error that caused the actual lrp to crash
1. `Since`: the timestamp when the actual lrp last crashed

[back](README.md)
