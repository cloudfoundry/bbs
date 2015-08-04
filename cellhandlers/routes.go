package cellhandlers

import "github.com/tedsuo/rata"

const (
	StopLRPInstanceRoute = "StopLRPInstance"
	CancelTaskRoute      = "CancelTask"
)

var Routes = rata.Routes{
	{Path: "/v1/lrps/:process_guid/instances/:instance_guid/stop", Method: "POST", Name: StopLRPInstanceRoute},
	{Path: "/v1/tasks/:task_guid/cancel", Method: "POST", Name: CancelTaskRoute},
}
