# Cells API Reference

This reference does not cover the protobuf payload supplied to each endpoint.

For detailed information on the structs and types listed see [models documentation](https://godoc.org/code.cloudfoundry.org/bbs/models)

# Cells APIs

## Cells

### BBS API Endpoint
Do an HTTP Get to "/v1/cells/list.r1"

### Golang Client API
```go
func (c *client) Cells(logger lager.Logger)([]*models.CellPresence, error)
```

#### Input
* `logger lager.Logger`
  * The logging sink

#### Output
* `[]*models.CellPresence`
  * [See CellPresence Documentation](https://godoc.org/code.cloudfoundry.org/bbs/models#CellPresence)
* `error`
  * Non-nil if error occurred

#### Example

```go
client := bbs.NewClient(url)
cells, err := client.Cells(logger)
if err != nil {
    log.Printf("failed to retrieve cells: " + err.Error())
}
```
