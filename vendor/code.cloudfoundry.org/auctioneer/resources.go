package auctioneer

// LRPStartRequest and TaskStartRequest have moved to code.cloudfoundry.org/bbs/models.
// These aliases are maintained for backward compatibility; vendored bbs code references them.

import bbsmodels "code.cloudfoundry.org/bbs/models"

type LRPStartRequest  = bbsmodels.LRPStartRequest
type TaskStartRequest = bbsmodels.TaskStartRequest

var NewLRPStartRequestFromSchedulingInfo = bbsmodels.NewLRPStartRequestFromSchedulingInfo
var NewTaskStartRequestFromModel         = bbsmodels.NewTaskStartRequestFromModel
