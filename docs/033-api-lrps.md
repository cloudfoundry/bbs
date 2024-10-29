---
title: LRP API Reference
expires_at : never
tags: [diego-release, bbs]
---
# Long Running Processes API Reference

This reference does not cover the protobuf payload supplied to each endpoint.
Instead, it illustrates calls to the API via the Golang `bbs.Client` interface.
Each method on that `Client` interface takes a `lager.Logger` as the first argument to log errors generated within the client.
This first `Logger` argument will not be duplicated on the descriptions of the method arguments.

For detailed information on the types referred to below, see the [godoc documentation for the BBS models](https://godoc.org/code.cloudfoundry.org/bbs/models).

# ActualLRP APIs

## ActualLRPs

Returns all [ActualLRPs](https://godoc.org/code.cloudfoundry.org/bbs/models#ActualLRP) matching the given [ActualLRPFilter](https://godoc.org/code.cloudfoundry.org/bbs/models#ActualLRPFilter).

### BBS API Endpoint

POST an [ActualLRPsRequest](https://godoc.org/code.cloudfoundry.org/bbs/models#ActualLRPsRequest)
to `/v1/actual_lrps/list`
and receive an [ActualLRPsResponse](https://godoc.org/code.cloudfoundry.org/bbs/models#ActualLRPsResponse).

### Golang Client API

```go
ActualLRPs(lager.Logger, models.ActualLRPFilter) ([]*models.ActualLRP, error)
```

#### Inputs

* `models.ActualLRPFilter`:
  * `Domain string`: If non-empty, filter to only ActualLRPs in this domain.
  * `CellId string`: If non-empty, filter to only ActualLRPs with this cell ID.
  * `ProcessGuid string`: If non-empty, filter to only ActualLRPs with this process GUID.
  * `Index *int32`: If non-nil, filter to only ActualLRPs with this instance index.

#### Output

* `[]*models.ActualLRP`: Slice of [`*models.ActualLRP`](https://godoc.org/code.cloudfoundry.org/bbs/models#ActualLRP).
* `error`:  Non-nil if an error occurred.


#### Example

```go
client := bbs.NewClient(url)
actualLRPs, err := client.ActualLRPs(logger, &models.ActualLRPFilter{
    Domain:      "some-domain",
    CellId:      "some-cell",
    ProcessGuid: "some-process-guid",
    Index:       &someIndex,
    })
if err != nil {
    log.Printf("failed to retrieve actual lrps: " + err.Error())
}
```


## ActualLRPGroups

**Deprecated** in favor of [ActualLRPs](#actuallrps).

Returns all [ActualLRPGroups](https://godoc.org/code.cloudfoundry.org/bbs/models#ActualLRPGroup) matching the given [ActualLRPFilter](https://godoc.org/code.cloudfoundry.org/bbs/models#ActualLRPFilter).

### BBS API Endpoint

POST an [ActualLRPGroupsRequest](https://godoc.org/code.cloudfoundry.org/bbs/models#ActualLRPGroupsRequest)
to `/v1/actual_lrp_groups/list`
and receive an [ActualLRPGroupsResponse](https://godoc.org/code.cloudfoundry.org/bbs/models#ActualLRPGroupsResponse).

### Golang Client API

```go
ActualLRPGroups(lager.Logger, models.ActualLRPFilter) ([]*models.ActualLRPGroup, error)
```

#### Inputs

* `models.ActualLRPFilter`:
  * `Domain string`: If non-empty, filter to only ActualLRPGroups in this domain.
  * `CellId string`: If non-empty, filter to only ActualLRPs with this cell ID.

#### Output

* `[]*models.ActualLRPGroup`: Slice of ActualLRPGroups. Either the `Instance` or the `Evacuating` [`*models.ActualLRP`](https://godoc.org/code.cloudfoundry.org/bbs/models#ActualLRP) may be present depending on the state of the LRP instances.
* `error`:  Non-nil if an error occurred.


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

**Deprecated** in favor of [ActualLRPs](#actuallrps) with an [ActualLRPFilter](https://godoc.org/code.cloudfoundry.org/bbs/models#ActualLRPFilter).

Returns all [ActualLRPGroups](https://godoc.org/code.cloudfoundry.org/bbs/models#ActualLRPGroup) for the given process guid.

### BBS API Endpoint


POST an [ActualLRPGroupsByProcessGuidRequest](https://godoc.org/code.cloudfoundry.org/bbs/models#ActualLRPGroupsByProcessGuidRequest)
to `/v1/actual_lrp_groups/list_by_process_guid`
and receive an [ActualLRPGroupsResponse](https://godoc.org/code.cloudfoundry.org/bbs/models#ActualLRPGroupsResponse).

### Golang Client API

```go
ActualLRPGroupsByProcessGuid(logger lager.Logger, processGuid string) ([]*models.ActualLRPGroup, error)
```

#### Inputs

* `processGuid string`: The process guid of the LRP.

#### Output

* `[]*models.ActualLRPGroup`: Slice of ActualLRPGroups. Either the `Instance` or the `Evacuating` [`*models.ActualLRP`](https://godoc.org/code.cloudfoundry.org/bbs/models#ActualLRP) may be present depending on the state of the LRP instances.
* `error`:  Non-nil if an error occurred.


#### Example
```go
client := bbs.NewClient(url)
actualLRPGroups, err := client.ActualLRPGroupsByProcessGuid(logger, "my-guid")
if err != nil {
    log.Printf("failed to retrieve actual lrps: " + err.Error())
}
```

## ActualLRPGroupByProcessGuidAndIndex

**Deprecated** in favor of [ActualLRPs](#actuallrps) with an [ActualLRPFilter](https://godoc.org/code.cloudfoundry.org/bbs/models#ActualLRPFilter).

Returns the [ActualLRPGroup](https://godoc.org/code.cloudfoundry.org/bbs/models#ActualLRPGroup) with the given process guid and instance index.

### BBS API Endpoint

POST an [ActualLRPGroupByProcessGuidAndIndexRequest](https://godoc.org/code.cloudfoundry.org/bbs/models#ActualLRPGroupByProcessGuidAndIndexRequest)
to
`/v1/actual_lrp_groups/get_by_process_guid_and_index`
and receive an [ActualLRPGroupResponse](https://godoc.org/code.cloudfoundry.org/bbs/models#ActualLRPGroupResponse).

### Golang Client API

```go
ActualLRPGroupByProcessGuidAndIndex(logger lager.Logger, processGuid string, index int) (*models.ActualLRPGroup, error)
```

#### Inputs

* `processGuid string`: The process guid to retrieve.
* `index int`: The instance index to retrieve.

#### Output

* `*models.ActualLRPGroup`: ActualLRPGroup for this LRP at this index. Either the `Instance` or the `Evacuating` [`*models.ActualLRP`](https://godoc.org/code.cloudfoundry.org/bbs/models#ActualLRP) may be present depending on the state of the LRP instances.
* `error`:  Non-nil if an error occurred.


#### Example
```go
client := bbs.NewClient(url)
actualLRPGroup, err := client.ActualLRPGroupByProcessGuidAndIndex(logger, "my-guid", 0)
if err != nil {
    log.Printf("failed to retrieve actual lrps: " + err.Error())
}
```


## RetireActualLRP

Stops the ActualLRP matching the given [ActualLRPKey](https://godoc.org/code.cloudfoundry.org/bbs/models#ActualLRPKey), but does not modify the desired state.

### BBS API Endpoint

POST a [RetireActualLRPRequest](https://godoc.org/code.cloudfoundry.org/bbs/models#RetireActualLRPRequest)
to `/v1/actual_lrps/retire`
and receive an [ActualLRPLifecycleResponse](https://godoc.org/code.cloudfoundry.org/bbs/models#ActualLRPLifecycleResponse).

### Golang Client API

```go
RetireActualLRP(logger lager.Logger, key *models.ActualLRPKey) error
```

#### Inputs

* `key *models.ActualLRPKey`: ActualLRPKey for the instance. Includes the LRP process guid, index, and LRP domain.

#### Output

* `error`:  Non-nil if an error occurred.


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

Lists all [DesiredLRPs](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRP) that match the given [DesiredLRPFilter](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPFilter).

### BBS API Endpoint

POST a [DesiredLRPsRequest](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPsRequest) to `/v1/desired_lrps/list.r3` and receive a [DesiredLRPsResponse](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPsResponse) with V3 [DesiredLRPs](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRP).

#### Deprecated Endpoints

* POST a [DesiredLRPsRequest](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPsRequest) to `/v1/desired_lrps/list.r2` and receive a [DesiredLRPsResponse](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPsResponse) with V2 [DesiredLRPs](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRP).
* POST a [DesiredLRPsRequest](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPsRequest) to `/v1/desired_lrps/list.r1` and receive a [DesiredLRPsResponse](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPsResponse) with V1 [DesiredLRPs](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRP).
* POST a [DesiredLRPsRequest](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPsRequest) to `/v1/desired_lrps/list` and receive a [DesiredLRPsResponse](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPsResponse) with V0 [DesiredLRPs](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRP).

### Golang Client API

```go
DesiredLRPs(logger lager.Logger, filter models.DesiredLRPFilter) ([]*models.DesiredLRP, error)
```

#### Inputs

* `filter models.DesiredLRPFilter`: [DesiredLRPFilter](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPFilter) to restrict the DesiredLRPs returned.
  * `Domain string`: If non-empty, filter to only DesiredLRPs in this domain.
  * `ProcessGuids []string`: If non-empty, filter to only DesiredLRPs with ProcessGuid in the given slice.

#### Output

* `[]*models.DesiredLRP`: List of [DesiredLRPs](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRP).
* `error`:  Non-nil if an error occurred.


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

Returns the DesiredLRP with the given process guid.

### BBS API Endpoint

POST a [DesiredLRPByProcessGuidRequest](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPByProcessGuidRequest) to `/v1/desired_lrps/get_by_process_guid.r3` and receive a [DesiredLRPResponse](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPResponse) with a V3 [DesiredLRP](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRP).

#### Deprecated Endpoints

* POST a [DesiredLRPByProcessGuidRequest](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPByProcessGuidRequest) to `/v1/desired_lrps/get_by_process_guid.r2` and receive a [DesiredLRPResponse](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPResponse) with a V2 [DesiredLRP](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRP).
* POST a [DesiredLRPByProcessGuidRequest](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPByProcessGuidRequest) to `/v1/desired_lrps/get_by_process_guid.r1` and receive a [DesiredLRPResponse](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPResponse) with a V1 [DesiredLRP](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRP).
* POST a [DesiredLRPByProcessGuidRequest](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPByProcessGuidRequest) to `/v1/desired_lrps/get_by_process_guid` and receive an [DesiredLRPResponse](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPResponse) with a V0 [DesiredLRP](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRP).

### Golang Client API

```go
DesiredLRPByProcessGuid(logger lager.Logger, processGuid string) (*models.DesiredLRP, error)
```

#### Inputs

* `processGuid string`: The GUID for the DesiredLRP.

#### Output

* `*models.DesiredLRP`: The requested [DesiredLRP](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRP).
* `error`:  Non-nil if an error occurred.


#### Example

```go
client := bbs.NewClient(url)
desiredLRP, err := client.DesiredLRPByProcessGuid(logger, "some-processs-guid")
if err != nil {
    log.Printf("failed to retrieve desired lrp: " + err.Error())
}
```

## DesiredLRPSchedulingInfos

Returns all DesiredLRPSchedulingInfos that match the given DesiredLRPFilter.

### BBS API Endpoint

POST a [DesiredLRPsRequest](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPsRequest)
to `/v1/desired_lrp_scheduling_infos/list`
and receive a [DesiredLRPSchedulingInfosResponse](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPSchedulingInfosResponse).

### Golang Client API

```go
DesiredLRPSchedulingInfos(logger lager.Logger, filter models.DesiredLRPFilter) ([]*models.DesiredLRPSchedulingInfo, error)
```

#### Inputs

* `filter models.DesiredLRPFilter`: [DesiredLRPFilter](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPFilter) to restrict the DesiredLRPs returned.
  * `Domain string`: If non-empty, filter to only DesiredLRPs in this domain.
  * `ProcessGuids []string`: If non-empty, filter to only DesiredLRPs with ProcessGuid in the given slice.

#### Output

* `[]*models.DesiredLRPSchedulingInfo`: List of [DesiredLRPSchedulingInfo](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPSchedulingInfo) records.
* `error`:  Non-nil if an error occurred.


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

## DesiredLRPSchedulingInfoByProcessGuid

Returns the DesiredLRPSchedulingInfo with the given process guid.

### BBS API Endpoint

POST a [DesiredLRPByProcessGuidRequest](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPByProcessGuidRequest) to `/v1/desired_lrp_scheduling_infos/get_by_process_guid` and receive a [DesiredLRPSchedulingInfoByProcessGuidResponse](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPSchedulingInfoByProcessGuidResponse) with a [DesiredLRPSchedulingInfo](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPSchedulingInfo).

### Golang Client API

```go
DesiredLRPSchedulingInfoByProcessGuid(ctx context.Context, logger lager.Logger, processGuid string) (*models.DesiredLRPSchedulingInfo, error)
```

#### Inputs

* `processGuid string`: The GUID for the DesiredLRPSchedulingInfo.

#### Output

* `*models.DesiredLRPSchedulingInfo`: The requested [DesiredLRPSchedulingInfo](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPSchedulingInfo).
* `error`:  Non-nil if an error occurred.


#### Example

```go
client := bbs.NewClient(url)
schedInfo, err := client.DesiredLRPBySchedulingInfoProcessGuid(logger, "some-processs-guid")
if err != nil {
    log.Printf("failed to retrieve desired lrp scheduling info: " + err.Error())
}
```

## DesiredLRPRoutingInfos

Returns all DesiredLRPs with only the routing info, that matches the given DesiredLRPFilter.

### BBS API Endpoint 

POST a [DesiredLRPsRequest](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPsRequest)
to `/v1/desired_lrp_routing_infos/list`
and receive a [DesiredLRPsResponse](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPsResponse).

### Golang Client API
```go
DesiredLRPRoutingInfos(ctx context.Context, logger lager.Logger, filter models.DesiredLRPFilter) ([]*models.DesiredLRP, error)
```

#### Inputs 

* `filter models.DesiredLRPFilter`: [DesiredLRPFilter](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPFilter) to restrict the DesiredLRPs returned.
  * `Domain string`: If non-empty, filter to only DesiredLRPs in this domain.
  * `ProcessGuids []string`: If non-empty, filter to only DesiredLRPs with ProcessGuid in the given slice.

#### Output 

* `[]*models.DesiredLRP`: List of [DesiredLRPS](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRP) records.
* Returned fields: 
```go 
".process_guid",
".domain",
".log_guid",
".instances",
".routes",
".modification_tag_epoch",
".modification_tag_index",
".run_info",
```
* `error`:  Non-nil if an error occurred.

#### Example

```go
client := bbs.NewClient(url)
routingInfos, err := client.DesiredLRPRoutingInfos(logger, &models.DesiredLRPFilter{
    Domain: "cf-apps",
})
if err != nil {
    log.Printf("failed to retrieve desired lrp routing info: " + err.Error())
}
```

## DesireLRP

Create a DesiredLRP and its corresponding associated ActualLRPs.

### BBS API Endpoint

POST a [DesireLRPRequest](https://godoc.org/code.cloudfoundry.org/bbs/models#DesireLRPRequest)
to `/v1/desired_lrp/desire.r1`
and receive a [DesiredLRPLifecycleResponse](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPLifecycleResponse).

#### Deprecated Endpoints

* POST a [DesireLRPRequest](https://godoc.org/code.cloudfoundry.org/bbs/models#DesireLRPRequest) with a V0 or V1 [DesiredLRP](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRP) to `/v1/desired_lrp/desire` and receive a [DesiredLRPLifecycleResponse](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPLifecycleResponse).


### Golang Client API

```go
DesireLRP(logger lager.Logger, desiredLRP *models.DesiredLRP) error
```

#### Inputs

* `desiredLRP *models.DesiredLRP`: [DesiredLRP](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRP) to create.

#### Output

* `error`:  Non-nil if an error occurred.

#### Example

See the [LRP Examples page](032-lrp-examples.md).


## UpdateDesiredLRP

Updates the [DesiredLRP](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRP) with the given process GUID.

### BBS API Endpoint

POST a [UpdateDesiredLRPRequest](https://godoc.org/code.cloudfoundry.org/bbs/models#UpdateDesiredLRPRequest)
to `/v1/desired_lrp/update`
and receive a [DesiredLRPLifecycleResponse](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPLifecycleResponse).

### Golang Client API

```go
UpdateDesiredLRP(logger lager.Logger, processGuid string, update *models.DesiredLRPUpdate) error
```

#### Inputs

* `processGuid string`: The GUID for the [DesiredLRP](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRP) to update.
* `update *models.DesiredLRPUpdate`: [DesiredLRPUpdate](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPUpdate) struct containing fields to update, if any.
  * `Instances *int32`: Optional. The number of instances.
  * `MetricTags map[string]*MetricTagValue`: Optional. Map of metric tags.
  * `Routes *Routes`: Optional. Map of routing information.
  * `Annotation *string`: Optional. The annotation string on the DesiredLRP.

#### Output

* `error`:  Non-nil if an error occurred.


#### Example

```go
client := bbs.NewClient(url)
instances := 4
annotation := "My annotation"
err := client.UpdateDesiredLRP(logger, "some-process-guid", &models.DesiredLRPUpdate{
    Instances: &instances,
    MetricTags: map[string]*models.MetricTagValue{"source_id": {Static: "some-guid"}},
    Routes: &models.Routes{},
    Annotation: &annotation,
})
if err != nil {
    log.Printf("failed to update desired lrp: " + err.Error())
}
```

## RemoveDesiredLRP

Removes the [DesiredLRP](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRP) with the given process GUID.

### BBS API Endpoint

POST a [RemoveDesiredLRPRequest](https://godoc.org/code.cloudfoundry.org/bbs/models#RemoveDesiredLRPRequest)
to `/v1/desired_lrp/remove`
and receive a [DesiredLRPLifecycleResponse](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRPLifecycleResponse).

### Golang Client API

```go
RemoveDesiredLRP(logger lager.Logger, processGuid string) error
```

#### Inputs

* `processGuid string`: The GUID for the [DesiredLRP](https://godoc.org/code.cloudfoundry.org/bbs/models#DesiredLRP) to remove.

#### Output

* `error`:  Non-nil if an error occurred.

#### Example

```go
client := bbs.NewClient(url)
err := client.RemoveDesiredLRP(logger, "some-process-guid")
if err != nil {
    log.Printf("failed to remove desired lrp: " + err.Error())
}
```
