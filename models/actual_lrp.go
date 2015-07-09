package models

import (
	"errors"
	"strings"
	"time"
)

const (
	ActualLRPStateUnclaimed = "UNCLAIMED"
	ActualLRPStateClaimed   = "CLAIMED"
	ActualLRPStateRunning   = "RUNNING"
	ActualLRPStateCrashed   = "CRASHED"
)

var ActualLRPStates = []string{
	ActualLRPStateUnclaimed,
	ActualLRPStateClaimed,
	ActualLRPStateRunning,
	ActualLRPStateCrashed,
}

type ActualLRPFilter struct {
	Domain string
	CellID string
}

func NewActualLRPKey(processGuid string, index int32, domain string) ActualLRPKey {
	return ActualLRPKey{&processGuid, &index, &domain}
}

func NewActualLRPInstanceKey(instanceGuid string, cellId string) ActualLRPInstanceKey {
	return ActualLRPInstanceKey{&instanceGuid, &cellId}
}

func NewActualLRPNetInfo(address string, ports []*PortMapping) ActualLRPNetInfo {
	return ActualLRPNetInfo{&address, ports}
}

func EmptyActualLRPNetInfo() ActualLRPNetInfo {
	return NewActualLRPNetInfo("", []*PortMapping{})
}

func (info ActualLRPNetInfo) Empty() bool {
	return info.GetAddress() == "" && len(info.GetPorts()) == 0
}

func (key ActualLRPInstanceKey) Empty() bool {
	return key.GetInstanceGuid() == "" && key.GetCellId() == ""
}

func (actual ActualLRP) ShouldRestartCrash(now time.Time, calc RestartCalculator) bool {
	if actual.GetState() != ActualLRPStateCrashed {
		return false
	}

	return calc.ShouldRestart(now.UnixNano(), actual.GetSince(), actual.GetCrashCount())
}

func (before ActualLRP) AllowsTransitionTo(lrpKey ActualLRPKey, instanceKey ActualLRPInstanceKey, newState string) bool {
	if !before.ActualLRPKey.Equal(&lrpKey) {
		return false
	}

	if before.GetState() == ActualLRPStateClaimed && newState == ActualLRPStateRunning {
		return true
	}

	if (before.GetState() == ActualLRPStateClaimed || before.GetState() == ActualLRPStateRunning) &&
		(newState == ActualLRPStateClaimed || newState == ActualLRPStateRunning) &&
		(!before.ActualLRPInstanceKey.Equal(&instanceKey)) {
		return false
	}

	return true
}

func (group ActualLRPGroup) Resolve() (*ActualLRP, bool, error) {
	if group.Instance == nil && group.Evacuating == nil {
		return nil, false, ErrActualLRPGroupInvalid
	}

	if group.Instance == nil {
		return group.Evacuating, true, nil
	}

	if group.Evacuating == nil {
		return group.Instance, false, nil
	}

	if group.Instance.GetState() == ActualLRPStateRunning || group.Instance.GetState() == ActualLRPStateCrashed {
		return group.Instance, false, nil
	} else {
		return group.Evacuating, true, nil
	}
}

func (actual ActualLRP) Validate() error {
	var validationError ValidationError

	err := actual.ActualLRPKey.Validate()
	if err != nil {
		validationError = validationError.Append(err)
	}

	if actual.GetSince() == 0 {
		validationError = validationError.Append(ErrInvalidField{"since"})
	}

	switch actual.GetState() {
	case ActualLRPStateUnclaimed:
		if !actual.ActualLRPInstanceKey.Empty() {
			validationError = validationError.Append(errors.New("instance key cannot be set when state is unclaimed"))
		}
		if !actual.ActualLRPNetInfo.Empty() {
			validationError = validationError.Append(errors.New("net info cannot be set when state is unclaimed"))
		}

	case ActualLRPStateClaimed:
		if err := actual.ActualLRPInstanceKey.Validate(); err != nil {
			validationError = validationError.Append(err)
		}
		if !actual.ActualLRPNetInfo.Empty() {
			validationError = validationError.Append(errors.New("net info cannot be set when state is claimed"))
		}
		if strings.TrimSpace(actual.GetPlacementError()) != "" {
			validationError = validationError.Append(errors.New("placement error cannot be set when state is claimed"))
		}

	case ActualLRPStateRunning:
		if err := actual.ActualLRPInstanceKey.Validate(); err != nil {
			validationError = validationError.Append(err)
		}
		if err := actual.ActualLRPNetInfo.Validate(); err != nil {
			validationError = validationError.Append(err)
		}
		if strings.TrimSpace(actual.GetPlacementError()) != "" {
			validationError = validationError.Append(errors.New("placement error cannot be set when state is running"))
		}

	case ActualLRPStateCrashed:
		if !actual.ActualLRPInstanceKey.Empty() {
			validationError = validationError.Append(errors.New("instance key cannot be set when state is crashed"))
		}
		if !actual.ActualLRPNetInfo.Empty() {
			validationError = validationError.Append(errors.New("net info cannot be set when state is crashed"))
		}
		if strings.TrimSpace(actual.GetPlacementError()) != "" {
			validationError = validationError.Append(errors.New("placement error cannot be set when state is crashed"))
		}

	default:
		validationError = validationError.Append(ErrInvalidField{"state"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (key *ActualLRPKey) Validate() error {
	var validationError ValidationError

	if key.GetProcessGuid() == "" {
		validationError = validationError.Append(ErrInvalidField{"process_guid"})
	}

	if key.GetIndex() < 0 {
		validationError = validationError.Append(ErrInvalidField{"index"})
	}

	if key.GetDomain() == "" {
		validationError = validationError.Append(ErrInvalidField{"domain"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (key *ActualLRPNetInfo) Validate() error {
	var validationError ValidationError

	if key.GetAddress() == "" {
		return validationError.Append(ErrInvalidField{"address"})
	}

	return nil
}

func (key *ActualLRPInstanceKey) Validate() error {
	var validationError ValidationError

	if key.GetCellId() == "" {
		validationError = validationError.Append(ErrInvalidField{"cell_id"})
	}

	if key.GetInstanceGuid() == "" {
		validationError = validationError.Append(ErrInvalidField{"instance_guid"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}
