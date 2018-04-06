package models

type ActualLRPKeyWithSchedulingInfo struct {
	Key            *ActualLRPKey
	InstanceKey    *ActualLRPInstanceKey
	SchedulingInfo *DesiredLRPSchedulingInfo
}
