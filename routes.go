package bbs

import "github.com/tedsuo/rata"

const (
	// Domains
	DomainsRoute      = "Domains"
	UpsertDomainRoute = "UpsertDomain"

	// Actual LRPs
	ActualLRPGroupsRoute                     = "ActualLRPGroups"
	ActualLRPGroupsByProcessGuidRoute        = "ActualLRPGroupsByProcessGuid"
	ActualLRPGroupByProcessGuidAndIndexRoute = "ActualLRPGroupsByProcessGuidAndIndex"

	// Actual LRP Lifecycle
	ClaimActualLRPRoute  = "ClaimActualLRP"
	StartActualLRPRoute  = "StartActualLRP"
	CrashActualLRPRoute  = "CrashActualLRP"
	FailActualLRPRoute   = "FailActualLRP"
	RemoveActualLRPRoute = "RemoveActualLRP"
	RetireActualLRPRoute = "RetireActualLRP"

	// Desired LRPs
	DesiredLRPsRoute             = "DesiredLRPs"
	DesiredLRPByProcessGuidRoute = "DesiredLRPByProcessGuid"

	// Desired LRPs
	TasksRoute      = "Tasks"
	TaskByGuidRoute = "TaskByGuid"

	// Event Streaming
	EventStreamRoute = "EventStream"
)

var Routes = rata.Routes{
	// Domains
	{Path: "/v1/domains", Method: "GET", Name: DomainsRoute},
	{Path: "/v1/domains/:domain", Method: "PUT", Name: UpsertDomainRoute},

	// Actual LRPs
	{Path: "/v1/actual_lrp_groups", Method: "GET", Name: ActualLRPGroupsRoute},
	{Path: "/v1/actual_lrp_groups/:process_guid", Method: "GET", Name: ActualLRPGroupsByProcessGuidRoute},
	{Path: "/v1/actual_lrp_groups/:process_guid/index/:index", Method: "GET", Name: ActualLRPGroupByProcessGuidAndIndexRoute},

	// Actual LRP Lifecycle
	{Path: "/v1/actual_lrps/claim", Method: "POST", Name: ClaimActualLRPRoute},
	{Path: "/v1/actual_lrps/start", Method: "POST", Name: StartActualLRPRoute},
	{Path: "/v1/actual_lrps/crash", Method: "POST", Name: CrashActualLRPRoute},
	{Path: "/v1/actual_lrps/fail", Method: "POST", Name: FailActualLRPRoute},
	{Path: "/v1/actual_lrps/:process_guid/index/:index", Method: "DELETE", Name: RemoveActualLRPRoute},
	{Path: "/v1/actual_lrps/retire", Method: "POST", Name: RetireActualLRPRoute},

	// Desired LRPs
	{Path: "/v1/desired_lrps", Method: "GET", Name: DesiredLRPsRoute},
	{Path: "/v1/desired_lrps/:process_guid", Method: "GET", Name: DesiredLRPByProcessGuidRoute},

	// Tasks
	{Path: "/v1/tasks", Method: "GET", Name: TasksRoute},
	{Path: "/v1/tasks/:task_guid", Method: "GET", Name: TaskByGuidRoute},

	// Event Streaming
	{Path: "/v1/events", Method: "GET", Name: EventStreamRoute},
}
