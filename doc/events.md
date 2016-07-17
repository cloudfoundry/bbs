# Events

The BBS emits events when a DesiredLRP or ActualLRP is created,
updated, or deleted. The following sections provide details on how to subscribe
to those events as well as the type of events supported by the BBS.

## Subscribing to events

You can use the `SubscribeToEvents(logger lager.Logger) (events.EventSource,
error)` client method to subscribe to events. For example:

``` go
client := bbs.NewClient(url)
eventSource, err := client.SubscribeToEvents(logger)
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

The following types of events are emitted:

## DesiredLRP events

### `DesiredLRPCreatedEvent`

When a new DesiredLRP is created, a
[DesiredLRPCreatedEvent](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPCreatedEvent)
is emitted. The value of the `DesiredLrp` field contains information about the
DesiredLRP that was just created.

### `DesiredLRPChangedEvent`

When a DesiredLRP changes, a
[DesiredLRPChangedEvent](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPChangedEvent)
is emitted. The value of the `Before` and `After` fields have information about the
DesiredLRP before and after the change.

### `DesiredLRPRemovedEvent`

When a DesiredLRP is deleted, a
[DesiredLRPRemovedEvent](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPRemovedEvent)
is emitted. The field value of `DesiredLrp` will have information about the
DesiredLRP that was just removed.

## ActualLRP events

### `ActualLRPCreatedEvent`

When a new ActualLRP is created, a
[ActualLRPCreatedEvent](https://godoc.org/code.cloudfoundry.org/bbs/models#ActualLRPCreatedEvent)
is emitted. The value of the `ActualLrpGroup` field contains more information
about the ActualLRP.


### `ActualLRPChangedEvent`

When a ActualLRP changes, a
[ActualLRPChangedEvent](https://godoc.org/code.cloudfoundry.org/bbs/models#ActualLRPChangedEvent)
is emitted. The value of the `Before` and `After` fields contains information about the
ActualLRP state before and after the change.

### `ActualLRPRemovedEvent`

When a ActualLRP is removed, a
[ActualLRPRemovedEvent](https://godoc.org/code.cloudfoundry.org/bbs/models#ActualLRPRemovedEvent)
is emitted. The value of the `ActualLrp` field contains information about the
ActualLRP that was just removed.

### `ActualLRPCrashedEvent`

When a ActualLRP crashes a
[ActualLRPCrashedEvent](https://godoc.org/code.cloudfoundry.org/bbs/models#ActualLRPCrashedEvent)
is emitted. The event will have the following field values:

1. `ActualLRPKey`: The LRP key of the ActualLRP.
1. `ActualLRPInstanceKey`: The instance key of the ActualLRP.
1. `CrashCount`: The number of times the ActualLRP has crashed, including this latest crash.
1. `CrashReason`: The last error that caused the ActualLRP to crash.
1. `Since`: The timestamp when the ActualLRP last crashed, in nanoseconds in the Unix epoch.
