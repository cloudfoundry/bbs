# Long Running Processes API Reference

This reference does not cover the protobuf payload supplied to each endpoint.

For detailed information on the structs and types listed see [models documentation](https://godoc.org/github.com/cloudfoundry-incubator/bbs/models)

# ActualLRP APIs

## ActualLRPs
Returns all ActualLRPGroups matching the given ActualLRPFilter

### BBS API Endpoint
To fetch all the ActualLRPs: post an ActualLRPGroupsRequest to "/v1/actual_lrp_groups/list"

### Golang Client API
```go
func (c *client) ActualLRPGroups(lager.Logger, models.ActualLRPFilter) ([]*models.ActualLRPGroup, error)
```

#### Input
* `logger lager.Logger`
  * The logging sink
* `filter models.ActualLRPFilter`
  * `Domain string`
  * `CellId string`

#### Output
* `[]*models.ActualLRPGroup`
  * `Instance *ActualLRP`
    * [See ActualLRP Documentation](https://godoc.org/github.com/cloudfoundry-incubator/bbs/models#ActualLRP)
  * `Evacuating *ActualLRP`
    * [See ActualLRP Documentation](https://godoc.org/github.com/cloudfoundry-incubator/bbs/models#ActualLRP)
  * Either or both of the above fields can be filled depending on the state of the ActualLRP
* `error`
  * Non-nil if error occurred

#### Example
```go
client := bbs.NewClient(url)
actualLRPGroups, err := client.ActualLRPGroups(logger, &models.ActualLRPFilter{
    Domain: "some-domain",
    CellId: "some-cell",
    })
if err != nil {
    log.Printf("failed to retrieve actual lrps: " + err.Error())
}
```


## ActualLRPsByProcessGuid
Returns all ActualLRPGroups that have the given process guid

### BBS API Endpoint
To fetch all the ActualLRPs by a processGuid: post an
ActualLRPGroupsByProcessGuidRequest to
"/v1/actual_lrp_groups/list_by_process_guid"

### Golang Client API
```go
func (c *client) ActualLRPGroupsByProcessGuid(logger lager.Logger, processGuid string) ([]*models.ActualLRPGroup, error)
```

#### Input
* `logger lager.Logger`
  * The logging sink
* `processGuid string`
  * The process Guid

#### Output
* `[]*models.ActualLRPGroup`
  * `Instance *ActualLRP`
    * [See ActualLRP Documentation](https://godoc.org/github.com/cloudfoundry-incubator/bbs/models#ActualLRP)
  * `Evacuating *ActualLRP`
    * [See ActualLRP Documentation](https://godoc.org/github.com/cloudfoundry-incubator/bbs/models#ActualLRP)
  * Either or both of the above fields can be filled depending on the state of the ActualLRP
* `error`
  * Non-nil if error occurred

#### Example
```go
client := bbs.NewClient(url)
actualLRPGroups, err := client.ActualLRPGroupsByProcessGuid(logger, "my-guid")
if err != nil {
    log.Printf("failed to retrieve actual lrps: " + err.Error())
}
```

## ActualLRPGroupByProcessGuidAndIndex
Returns the ActualLRPGroup with the given process guid and instance index

### BBS API Endpoint
Post an ActualLRPGroupByProcessGuidAndIndexRequest to
"/v1/actual_lrp_groups/get_by_process_guid_and_index"

### Golang Client API
```go
func (c *client) ActualLRPGroupByProcessGuidAndIndex(logger lager.Logger, processGuid string, index int) (*models.ActualLRPGroup, error)
```

#### Input
* `logger lager.Logger`
  * The logging sink
* `processGuid string`
  * The process Guid
* `index int`
  * The instance index

#### Output
* `*models.ActualLRPGroup`
  * `Instance *ActualLRP`
    * [See ActualLRP Documentation](https://godoc.org/github.com/cloudfoundry-incubator/bbs/models#ActualLRP)
  * `Evacuating *ActualLRP`
    * [See ActualLRP Documentation](https://godoc.org/github.com/cloudfoundry-incubator/bbs/models#ActualLRP)
  * Either or both of the above fields can be filled depending on the state of the ActualLRP
* `error`
  * Non-nil if error occurred

#### Example
```go
client := bbs.NewClient(url)
actualLRPGroup, err := client.ActualLRPGroupByProcessGuidAndIndex(logger, "my-guid", 0)
if err != nil {
    log.Printf("failed to retrieve actual lrps: " + err.Error())
}
```


## RetireActualLRP
Shuts down the ActualLRP matching the given ActualLRPKey, but does not modify the desired state

### BBS API Endpoint
Post a RetireActualLRPRequest to "/v1/actual_lrps/retire"

### Golang Client API
```go
func (c *client) RetireActualLRP(logger lager.Logger, key *models.ActualLRPKey) error
```

#### Input
* `logger lager.Logger`
  * The logging sink
* `key *models.ActualLRPKey`
  * `ProcessGuid string`
    * The process Guid
  * `Index int32`
    * The instance index
  * `Domain string`
    * The domain

#### Output
* `error`
  * Non-nil if error occurred

#### Example
```go
client := bbs.NewClient(url)
err := client.RetireActualLRP(logger, &models.ActualLRPKey{
    ProcessGuid: "some-process-guid",
    Index: 0,
    Domain: "cf-apps",
})
if err != nil {
    log.Printf("failed to retire actual lrps: " + err.Error())
}
```
# DesiredLRP APIs

## DesiredLRPs
Lists all DesiredLRPs that match the given DesiredLRPFilter

### BBS API Endpoint
Post a DesiredLRPsRequest to "/v1/desired_lrps/list.r1"

DEPRECATED: Post a DesiredLRPsRequest to "/v1/desired_lrps/list"

### Golang Client API
```go
func (c *client) DesiredLRPs(logger lager.Logger, filter models.DesiredLRPFilter) ([]*models.DesiredLRP, error)
```

#### Input
* `logger lager.Logger`
  * The logging sink
* `filter models.DesiredLRPFilter`
  * `Domain string`
    * The domain (optional)

#### Output
* `[]*models.DesiredLRP`
  * [See DesiredLRP Documentation](https://godoc.org/github.com/cloudfoundry-incubator/bbs/models#DesiredLRP)
* `error`
  * Non-nil if error occurred

#### Example
```go
client := bbs.NewClient(url)
desiredLRPS, err := client.DesiredLRPs(logger, &models.DesiredLRPFilter{
    Domain: "cf-apps",
})
if err != nil {
    log.Printf("failed to retrieve desired lrps: " + err.Error())
}
```

## DesiredLRPByProcessGuid
Returns the DesiredLRP with the given process guid

### BBS API Endpoint
Post a DesiredLRPByProcessGuidRequest to "/v1/desired_lrps/get_by_process_guid.r1"

DEPRECATED: Post a DesiredLRPByProcessGuidRequest to "/v1/desired_lrps/get_by_process_guid"

### Golang Client API
```go
func (c *client) DesiredLRPByProcessGuid(logger lager.Logger, processGuid string) (*models.DesiredLRP, error)
```

#### Input
* `logger lager.Logger`
  * The logging sink
* `processGuid string`
  * The process Guid

#### Output
* `*models.DesiredLRP`
  * [See DesiredLRP Documentation](https://godoc.org/github.com/cloudfoundry-incubator/bbs/models#DesiredLRP)
* `error`
  * Non-nil if error occurred

#### Example
```go
client := bbs.NewClient(url)
desiredLRP, err := client.DesiredLRPByProcessGuid(logger, "some-processs-guid")
if err != nil {
    log.Printf("failed to retrieve desired lrp: " + err.Error())
}
```

## DesiredLRPSchedulingInfos
Returns all DesiredLRPSchedulingInfos that match the given DesiredLRPFilter

### BBS API Endpoint
Post a DesiredLRPsRequest to "/v1/desired_lrp_scheduling_infos/list"

### Golang Client API
```go
func (c *client) DesiredLRPSchedulingInfos(logger lager.Logger, filter models.DesiredLRPFilter) ([]*models.DesiredLRPSchedulingInfo, error)
```

#### Input
* `logger lager.Logger`
  * The logging sink
* `filter models.DesiredLRPFilter`
  * `Domain string`
    * The domain (optional)

#### Output
* `[]*models.DesiredLRPSchedulingInfo`
  * [See DesiredLRPSchedulingInfo Documentation](https://godoc.org/github.com/cloudfoundry-incubator/bbs/models#DesiredLRPSchedulingInfo)
* `error`
  * Non-nil if error occurred

#### Example
```go
client := bbs.NewClient(url)
info, err := client.DesiredLRPSchedulingInfos(logger, &models.DesiredLRPFilter{
    Domain: "cf-apps",
})
if err != nil {
    log.Printf("failed to retrieve desired lrp scheduling info: " + err.Error())
}
```

## DesireLRP
Creates the given DesiredLRP and its corresponding ActualLRPs

### BBS API Endpoint
Post a DesireLRPRequest to "/v1/desired_lrp/desire.r1"

DEPRECATED: Post a DesireLRPRequest to "/v1/desired_lrp/desire"

### Golang Client API
```go
func (c *client) DesireLRP(logger lager.Logger, desiredLRP *models.DesiredLRP) error
```

#### Input
* `logger lager.Logger`
  * The logging sink
* `desiredLRP *models.DesiredLRP`
  * See the [LRP Examples page](lrp-examples.md)
    for how to create a desired LRP

#### Output
* `error`
  * Non-nil if error occurred

#### Example
See the [LRP Examples page](lrp-examples.md)
for example

## UpdateDesiredLRP
Updates the DesiredLRP matching the given process guid

### BBS API Endpoint
Post a UpdateDesiredLRPRequest to "/v1/desired_lrp/update"

### Golang Client API
```go
func (c *client) UpdateDesiredLRP(logger lager.Logger, processGuid string, update *models.DesiredLRPUpdate) error
```

#### Input
* `logger lager.Logger`
  * The logging sink
* `processGuid string`
  * The process Guid
* `update *models.DesiredLRPUpdate`
  * `Instances *int32`
    * The number of instances
  * `Routes *Routes`
    * Raw route information
  * `Annotation *string`
    * A string annotation

#### Output
* `error`
  * Non-nil if error occurred

#### Example
```go
client := bbs.NewClient(url)
instances := 4
annotation := "My annotation"
err := client.UpdateDesiredLRP(logger, "some-process-guid", &models.DesiredLRPUpdate{
    Instances: &instances,
    Routes: &models.Routes{},
    Annotation: &annotation,
})
if err != nil {
    log.Printf("failed to update desired lrp: " + err.Error())
}
```

## RemoveDesiredLRP
Removes the DesiredLRP matching the given process guid

### BBS API Endpoint
Post a RemoveDesiredLRPRequest to "/v1/desired_lrp/remove"

### Golang Client API
```go
func (c *client) RemoveDesiredLRP(logger lager.Logger, processGuid string) error
```

#### Input
* `logger lager.Logger`
  * The logging sink
* `processGuid string`
  * The process Guid

#### Output
* `error`
  * Non-nil if error occurred

#### Example
```go
client := bbs.NewClient(url)
err := client.RemoveDesiredLRP(logger, "some-process-guid")
if err != nil {
    log.Printf("failed to remove desired lrp: " + err.Error())
}
```
