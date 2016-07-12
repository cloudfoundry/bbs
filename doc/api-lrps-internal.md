# Long Running Processes Internal API Reference

This reference does not cover the protobuf payload supplied to each endpoint.
Instead, it illustrates calls to the API via the Golang `bbs.InternalClient` interface.
Each method on that `InternalClient` interface takes a `lager.Logger` as the first argument to log errors generated within the client.
This first `Logger` argument will not be duplicated on the descriptions of the method arguments.

For detailed information on the types referred to below, see the [godoc documentation for the BBS models](https://godoc.org/code.cloudfoundry.org/bbs/models).

# ActualLRP APIs

## ClaimActualLRP

Called to claim an actual LRP and associate it will the cell.

### BBS API Endpoint

POST an [ClaimActualLRPRequest](https://godoc.org/code.cloudfoundry.org/bbs/models#ClaimActualLRPRequest) to `/v1/actual_lrps/claim`, and receive an [ActualLRPLifecycleResponse](https://godoc.org/code.cloudfoundry.org/bbs/models#ActualLRPLifecycleResponse).

### Golang Client API

```go
func (c *client) ClaimActualLRP(logger lager.Logger, processGuid string, index int, instanceKey *models.ActualLRPInstanceKey)
```

#### Inputs

* `processGUID`: The process GUID
* `index int`: Index of the LRP.
* `instanceKey *modles.ActualLRPInstanceKey`: ActualLRP InstanceKey
  * `InstanceGuid string`
    * The instance GUID
  * `CellID string`
    * Unique identifier for the Cell

#### Output

* `error`:  Non-nil if an error occurred.


#### Example

```go
client := bbs.NewClient(url)
err := client.ClaimActualLRP(logger, "some-guid", 0, &models.ActualLRPInstanceKey{
	InstanceGuid: "some-instance-guid",
	CellId: "some-cellID",
)
if err != nil {
    log.Printf("failed to claim actual lrp: " + err.Error())
}
```

## StartActualLRP

Called to start an ActualLRP

### BBS API Endpoint

POST an [StartActualLRPRequest](https://godoc.org/code.cloudfoundry.org/bbs/models#StartActualLRPRequest) to `/v1/actual_lrps/start`, and receive an [ActualLRPLifecycleResponse](https://godoc.org/code.cloudfoundry.org/bbs/models#ActualLRPLifecycleResponse).

### Golang Client API

```go
func (c *client) StartActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, netInfo *models.ActualLRPNetInfo) error
```

#### Inputs

* `key *modles.ActualLRPKey`: ActualLRP InstanceKey
  * `ProcessGuid string`
    * The process GUID
  * `Index int32`
    * Process Index
  * `Domain string`
    * The Domain
* `instanceKey *modles.ActualLRPInstanceKey`: ActualLRP InstanceKey
  * `InstanceGuid string`
    * The instance GUID
  * `CellID string`
    * Unique identifier for the Cell
* `netInfo *models.ActualLRPNetInfo`: Networking information for the Actual LRP
  * `Address string`
    * IP Address
  * `Ports []*models.PortMapping`
    * `ContainerPort uint32`
      * The port on the container
    * `HostPort uint32`
      * The port on the host

#### Output

* `error`:  Non-nil if an error occurred.


#### Example

```go
client := bbs.NewClient(url)
err := client.StartActualLRP(logger, &models.ActualLRPKey{
	   ProcessGuid: "some-guid",
	   Index: 0,
	   Domain: "some-domain",
	},
	&models.ActualLRPInstanceKey{
	    InstanceGuid: "some-instance-guid",
	    CellId: "some-cellID",
	},
	&models.ActualLRPNetInfo{
	    Address: "1.2.3.4",
	    models.NewPortMapping(10,20),
	},
    )
)
if err != nil {
    log.Printf("failed to start actual lrp: " + err.Error())
}
```

## CrashActualLRP

Called to mark an ActualLRP instance as crashed

### BBS API Endpoint

POST an [CrashActualLRPRequest](https://godoc.org/code.cloudfoundry.org/bbs/models#CrashActualLRPRequest) to `/v1/actual_lrps/crash`, and receive an [ActualLRPLifecycleResponse](https://godoc.org/code.cloudfoundry.org/bbs/models#ActualLRPLifecycleResponse).

### Golang Client API

```go
func (c *client) CrashActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, errorMessage string) error
```

#### Inputs

* `key *modles.ActualLRPKey`: ActualLRP
  * `ProcessGuid string`
    * The process GUID
  * `Index int32`
    * Process Index
  * `Domain string`
    * The Domain
* `instanceKey *modles.ActualLRPInstanceKey`: ActualLRP InstanceKey
  * `InstanceGuid string`
    * The instance GUID
  * `CellID string`
    * Unique identifier for the Cell
* errorMessage string: The Error message describing the crash reason

#### Output

* `error`:  Non-nil if an error occurred.


#### Example

```go
client := bbs.NewClient(url)
err := client.CrashActualLRP(logger, &models.ActualLRPKey{
	   ProcessGuid: "some-guid",
	   Index: 0,
	   Domain: "some-domain",
	},
	&models.ActualLRPInstanceKey{
	    InstanceGuid: "some-instance-guid",
	    CellId: "some-cellID",
	},
	"Crashed Reason",
    )
)
if err != nil {
    log.Printf("failed to crash actual lrp: " + err.Error())
}
```

## FailActualLRP

Called to mark an ActualLRP as failed

### BBS API Endpoint

POST an [FailActualLRPRequest](https://godoc.org/code.cloudfoundry.org/bbs/models#FailActualLRPRequest) to `/v1/actual_lrps/fail`, and receive an [ActualLRPLifecycleResponse](https://godoc.org/code.cloudfoundry.org/bbs/models#ActualLRPLifecycleResponse).

### Golang Client API

```go
func (c *client) FailActualLRP(logger lager.Logger, key *models.ActualLRPKey, errorMessage string) error
```

#### Inputs

* `key *modles.ActualLRPKey`: ActualLRP
  * `ProcessGuid string`
    * The process GUID
  * `Index int32`
    * Process Index
  * `Domain string`
    * The Domain
* errorMessage string: The Error message desribing the crash reason

#### Output

* `error`:  Non-nil if an error occurred.

#### Example

```go
client := bbs.NewClient(url)
err := client.FailActualLRP(logger, &models.ActualLRPKey{
	   ProcessGuid: "some-guid",
	   Index: 0,
	   Domain: "some-domain",
	},
	"Failure Reason",
    )
)
if err != nil {
    log.Printf("failed to fail actual lrp: " + err.Error())
}
```

## RemoveActualLRP

Called to remove an ActualLRP

### BBS API Endpoint

POST an [RemoveActualLRPRequest](https://godoc.org/code.cloudfoundry.org/bbs/models#RemoveActualLRPRequest) to `/v1/actual_lrps/remove`, and receive an [ActualLRPLifecycleResponse](https://godoc.org/code.cloudfoundry.org/bbs/models#ActualLRPLifecycleResponse).

### Golang Client API

```go
func (c *client) RemoveActualLRP(logger lager.Logger, processGuid string, index int, instanceKey *models.ActualLRPInstanceKey) error
```

#### Inputs

* `ProcessGuid string`
  * The process GUID
* `Index int32`
  * Process Index
* `instanceKey *modles.ActualLRPInstanceKey`: ActualLRP InstanceKey
  * `InstanceGuid string`
    * The instance GUID
  * `CellID string`
    * Unique identifier for the Cell

#### Output

* `error`:  Non-nil if an error occurred.

#### Example

```go
client := bbs.NewClient(url)
err := client.RemoveActualLRP(logger, "some-guid", 0, &models.ActualLRPInstanceKey{
	InstanceGuid: "some-instance-guid",
	CellId: "some-cellID",
)
)
if err != nil {
    log.Printf("failed to remove an actual lrp: " + err.Error())
}
```

## EvacuateClaimedActualLRP

Called to evacuate a claimed actual LRP.

### BBS API Endpoint

POST an [EvacuateClaimedActualLRPRequest](https://godoc.org/code.cloudfoundry.org/bbs/models#EvacuateClaimedActualLRPRequest) to `/v1/actual_lrps/evacuate_claimed`, and receive an [EvacuationResponse](https://godoc.org/code.cloudfoundry.org/bbs/models#EvacuationResponse).

### Golang Client API

```go
func (c *client) EvacuateClaimedActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey) (bool, error)
```

#### Inputs

* `key *modles.ActualLRPKey`: ActualLRP InstanceKey
  * `ProcessGuid string`
    * The process GUID
  * `Index int32`
    * Process Index
  * `Domain string`
    * The Domain
* `instanceKey *modles.ActualLRPInstanceKey`: ActualLRP InstanceKey
  * `InstanceGuid string`
    * The instance GUID
  * `CellID string`
    * Unique identifier for the Cell

#### Output

* `bool`: Flag to indicate if the container should be kept
* `error`:  Non-nil if an error occurred.

#### Example

```go
client := bbs.NewClient(url)
keepContainer, err := client.EvacuateClaimedActualLRP(logger, &models.ActualLRPKey{
	       ProcessGuid: "some-guid",
	       Index: 0,
	       Domain: "some-domain",
	},
	&models.ActualLRPInstanceKey{
	    InstanceGuid: "some-instance-guid",
	    CellId: "some-cellID",
)
if err != nil {
    log.Printf("failed to evacuate claimed actual lrp: " + err.Error())
}
```

## EvacuateCrashedActualLRP

Called to evacuate a crashed actual LRP.

### BBS API Endpoint

POST an [EvacuateCrashedActualLRPRequest](https://godoc.org/code.cloudfoundry.org/bbs/models#EvacuateCrashedActualLRPRequest) to `/v1/actual_lrps/evacuate_crashed`, and receive an [EvacuationResponse](https://godoc.org/code.cloudfoundry.org/bbs/models#EvacuationResponse).

### Golang Client API

```go
func (c *client) EvacuateCrashedActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey, errorMessage string) (bool, error)
```

#### Inputs

* `key *modles.ActualLRPKey`: ActualLRP InstanceKey
  * `ProcessGuid string`
    * The process GUID
  * `Index int32`
    * Process Index
  * `Domain string`
    * The Domain
* `instanceKey *modles.ActualLRPInstanceKey`: ActualLRP InstanceKey
  * `InstanceGuid string`
    * The instance GUID
  * `CellID string`
    * Unique identifier for the Cell
* errorMessage string: The crashing error message

#### Output

* `bool`: Flag to indicate if the container should be kept
* `error`:  Non-nil if an error occurred.

#### Example

```go
client := bbs.NewClient(url)
keepContainer, err := client.EvacuateCrashedActualLRP(logger, &models.ActualLRPKey{
	       ProcessGuid: "some-guid",
	       Index: 0,
	       Domain: "some-domain",
	},
	&models.ActualLRPInstanceKey{
	    InstanceGuid: "some-instance-guid",
	    CellId: "some-cellID",
	"some error message",
)
if err != nil {
    log.Printf("failed to evacuate crashed actual lrp: " + err.Error())
}
```

## EvacuateStoppedActualLRP

Called to evacuate a stopped actual LRP.

### BBS API Endpoint

POST an [EvacuateStoppedActualLRPRequest](https://godoc.org/code.cloudfoundry.org/bbs/models#EvacuateStoppedActualLRPRequest) to `/v1/actual_lrps/evacuate_stopped`, and receive an [EvacuationResponse](https://godoc.org/code.cloudfoundry.org/bbs/models#EvacuationResponse).

### Golang Client API

```go
func (c *client) EvacuateStoppedActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey) (bool, error)
```

#### Inputs

* `key *modles.ActualLRPKey`: ActualLRP InstanceKey
  * `ProcessGuid string`
    * The process GUID
  * `Index int32`
    * Process Index
  * `Domain string`
    * The Domain
* `instanceKey *modles.ActualLRPInstanceKey`: ActualLRP InstanceKey
  * `InstanceGuid string`
    * The instance GUID
  * `CellID string`
    * Unique identifier for the Cell

#### Output

* `bool`: Flag to indicate if the container should be kept
* `error`:  Non-nil if an error occurred.

#### Example

```go
client := bbs.NewClient(url)
keepContainer, err := client.EvacuateStoppedActualLRP(logger, &models.ActualLRPKey{
	       ProcessGuid: "some-guid",
	       Index: 0,
	       Domain: "some-domain",
	},
	&models.ActualLRPInstanceKey{
	    InstanceGuid: "some-instance-guid",
	    CellId: "some-cellID",
	"some error message",
)
if err != nil {
    log.Printf("failed to evacuate stopped actual lrp: " + err.Error())
}
```

## EvacuateRunningActualLRP

Called to evacuate a running actual LRP.

### BBS API Endpoint

POST an [EvacuateRunningActualLRPRequest](https://godoc.org/code.cloudfoundry.org/bbs/models#EvacuateRunningActualLRPRequest) to `/v1/actual_lrps/evacuate_running`, and receive an [EvacuationResponse](https://godoc.org/code.cloudfoundry.org/bbs/models#EvacuationResponse).

### Golang Client API

```go
func (c *client) EvacuateRunningActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey) (bool, error)
```

#### Inputs

* `key *modles.ActualLRPKey`: ActualLRP InstanceKey
  * `ProcessGuid string`
    * The process GUID
  * `Index int32`
    * Process Index
  * `Domain string`
    * The Domain
* `instanceKey *modles.ActualLRPInstanceKey`: ActualLRP InstanceKey
  * `InstanceGuid string`
    * The instance GUID
  * `CellID string`
    * Unique identifier for the Cell

#### Output

* `bool`: Flag to indicate if the container should be kept
* `error`:  Non-nil if an error occurred.

#### Example

```go
client := bbs.NewClient(url)
keepContainer, err := client.EvacuateRunningActualLRP(logger, &models.ActualLRPKey{
	       ProcessGuid: "some-guid",
	       Index: 0,
	       Domain: "some-domain",
	},
	&models.ActualLRPInstanceKey{
	    InstanceGuid: "some-instance-guid",
	    CellId: "some-cellID",
	"some error message",
)
if err != nil {
    log.Printf("failed to evacuate running actual lrp: " + err.Error())
}
```

## RemoveEvacuatingActualLRP

Called to remove an evacuating LRP.

### BBS API Endpoint

POST an [RemoveEvacuatingActualLRPRequest](https://godoc.org/code.cloudfoundry.org/bbs/models#RemoveEvacuatingActualLRPRequest) to `/v1/actual_lrps/remove_evacuating`, and receive an [RemoveEvacuatingActualLRPResponse](https://godoc.org/code.cloudfoundry.org/bbs/models#RemoveEvacuatingActualLRPResponse).

### Golang Client API

```go
func (c *client) RemoveEvacuatingActualLRP(logger lager.Logger, key *models.ActualLRPKey, instanceKey *models.ActualLRPInstanceKey) error
```

#### Inputs

* `key *modles.ActualLRPKey`: ActualLRP InstanceKey
  * `ProcessGuid string`
    * The process GUID
  * `Index int32`
    * Process Index
  * `Domain string`
    * The Domain
* `instanceKey *modles.ActualLRPInstanceKey`: ActualLRP InstanceKey
  * `InstanceGuid string`
    * The instance GUID
  * `CellID string`
    * Unique identifier for the Cell

#### Output

* `error`:  Non-nil if an error occurred.

#### Example

```go
client := bbs.NewClient(url)
err := client.RemoveEvacuatingActualLRP(logger, &models.ActualLRPKey{
	       ProcessGuid: "some-guid",
	       Index: 0,
	       Domain: "some-domain",
	},
	&models.ActualLRPInstanceKey{
	    InstanceGuid: "some-instance-guid",
	    CellId: "some-cellID",
	"some error message",
)
if err != nil {
    log.Printf("failed to remove evacuating actual lrp: " + err.Error())
}
```
