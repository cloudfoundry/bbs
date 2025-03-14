package models

import "encoding/json"

func (request *ProtoActualLRPsRequest) Validate() error {
	return request.FromProto().Validate()
}

func (request *ActualLRPsRequest) Validate() error {
	return nil
}

type internalActualLRPsRequest struct {
	Domain      string `json:"domain"`
	CellId      string `json:"cell_id"`
	ProcessGuid string `json:"process_guid"`
	Index       *int32 `json:"index,omitempty"`
}

func (request *ActualLRPsRequest) UnmarshalJSON(data []byte) error {
	var internalRequest internalActualLRPsRequest
	if err := json.Unmarshal(data, &internalRequest); err != nil {
		return err
	}

	request.Domain = internalRequest.Domain
	request.CellId = internalRequest.CellId
	request.ProcessGuid = internalRequest.ProcessGuid
	if internalRequest.Index != nil {
		request.SetIndex(internalRequest.Index)
	}

	return nil
}

func (request ActualLRPsRequest) MarshalJSON() ([]byte, error) {
	internalRequest := internalActualLRPsRequest{
		Domain:      request.Domain,
		CellId:      request.CellId,
		ProcessGuid: request.ProcessGuid,
	}

	i := request.GetIndex()
	internalRequest.Index = i
	return json.Marshal(internalRequest)
}

func (request *ProtoActualLRPGroupsRequest) Validate() error {
	return request.FromProto().Validate()
}

// Deprecated: use the ActualLRPInstances API instead
func (request *ActualLRPGroupsRequest) Validate() error {
	return nil
}

func (request *ProtoActualLRPGroupsByProcessGuidRequest) Validate() error {
	return request.FromProto().Validate()
}

// Deprecated: use the ActualLRPInstances API instead
func (request *ActualLRPGroupsByProcessGuidRequest) Validate() error {
	var validationError ValidationError

	if request.ProcessGuid == "" {
		validationError = validationError.Append(ErrInvalidField{"process_guid"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (request *ProtoActualLRPGroupByProcessGuidAndIndexRequest) Validate() error {
	return request.FromProto().Validate()
}

// Deprecated: use the ActualLRPInstances API instead
func (request *ActualLRPGroupByProcessGuidAndIndexRequest) Validate() error {
	var validationError ValidationError

	if request.ProcessGuid == "" {
		validationError = validationError.Append(ErrInvalidField{"process_guid"})
	}

	if request.Index < 0 {
		validationError = validationError.Append(ErrInvalidField{"index"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (request *ProtoRemoveActualLRPRequest) Validate() error {
	return request.FromProto().Validate()
}

func (request *RemoveActualLRPRequest) Validate() error {
	var validationError ValidationError

	if request.ProcessGuid == "" {
		validationError = validationError.Append(ErrInvalidField{"process_guid"})
	}

	if request.Index < 0 {
		validationError = validationError.Append(ErrInvalidField{"index"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (request *ProtoClaimActualLRPRequest) Validate() error {
	return request.FromProto().Validate()
}

func (request *ClaimActualLRPRequest) Validate() error {
	var validationError ValidationError

	if request.ProcessGuid == "" {
		validationError = validationError.Append(ErrInvalidField{"process_guid"})
	}

	if request.ActualLrpInstanceKey == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_instance_key"})
	} else if err := request.ActualLrpInstanceKey.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (request *ProtoStartActualLRPRequest) Validate() error {
	return request.FromProto().Validate()
}

func (request *StartActualLRPRequest) Validate() error {
	var validationError ValidationError

	if request.ActualLrpKey == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_key"})
	} else if err := request.ActualLrpKey.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if request.ActualLrpInstanceKey == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_instance_key"})
	} else if err := request.ActualLrpInstanceKey.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if request.ActualLrpNetInfo == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_net_info"})
	} else if err := request.ActualLrpNetInfo.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (request *ProtoCrashActualLRPRequest) Validate() error {
	return request.FromProto().Validate()
}

func (request *CrashActualLRPRequest) Validate() error {
	var validationError ValidationError

	if request.ActualLrpKey == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_key"})
	} else if err := request.ActualLrpKey.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if request.ActualLrpInstanceKey == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_instance_key"})
	} else if err := request.ActualLrpInstanceKey.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (request *ProtoFailActualLRPRequest) Validate() error {
	return request.FromProto().Validate()
}

func (request *FailActualLRPRequest) Validate() error {
	var validationError ValidationError

	if request.ActualLrpKey == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_key"})
	} else if err := request.ActualLrpKey.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if request.ErrorMessage == "" {
		validationError = validationError.Append(ErrInvalidField{"error_message"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (request *ProtoRetireActualLRPRequest) Validate() error {
	return request.FromProto().Validate()
}

func (request *RetireActualLRPRequest) Validate() error {
	var validationError ValidationError

	if request.ActualLrpKey == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_key"})
	} else if err := request.ActualLrpKey.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (request *ProtoRemoveEvacuatingActualLRPRequest) Validate() error {
	return request.FromProto().Validate()
}

func (request *RemoveEvacuatingActualLRPRequest) Validate() error {
	var validationError ValidationError

	if request.ActualLrpKey == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_key"})
	} else if err := request.ActualLrpKey.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if request.ActualLrpInstanceKey == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_instance_key"})
	} else if err := request.ActualLrpInstanceKey.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (request *ProtoEvacuateClaimedActualLRPRequest) Validate() error {
	return request.FromProto().Validate()
}

func (request *EvacuateClaimedActualLRPRequest) Validate() error {
	var validationError ValidationError

	if request.ActualLrpKey == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_key"})
	} else if err := request.ActualLrpKey.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if request.ActualLrpInstanceKey == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_instance_key"})
	} else if err := request.ActualLrpInstanceKey.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (request *ProtoEvacuateCrashedActualLRPRequest) Validate() error {
	return request.FromProto().Validate()
}

func (request *EvacuateCrashedActualLRPRequest) Validate() error {
	var validationError ValidationError

	if request.ActualLrpKey == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_key"})
	} else if err := request.ActualLrpKey.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if request.ActualLrpInstanceKey == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_instance_key"})
	} else if err := request.ActualLrpInstanceKey.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if request.ErrorMessage == "" {
		validationError = validationError.Append(ErrInvalidField{"error_message"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (request *ProtoEvacuateStoppedActualLRPRequest) Validate() error {
	return request.FromProto().Validate()
}

func (request *EvacuateStoppedActualLRPRequest) Validate() error {
	var validationError ValidationError

	if request.ActualLrpKey == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_key"})
	} else if err := request.ActualLrpKey.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if request.ActualLrpInstanceKey == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_instance_key"})
	} else if err := request.ActualLrpInstanceKey.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (request *ProtoEvacuateRunningActualLRPRequest) Validate() error {
	return request.FromProto().Validate()
}

func (request *EvacuateRunningActualLRPRequest) Validate() error {
	var validationError ValidationError

	if request.ActualLrpKey == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_key"})
	} else if err := request.ActualLrpKey.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if request.ActualLrpInstanceKey == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_instance_key"})
	} else if err := request.ActualLrpInstanceKey.Validate(); err != nil {
		validationError = validationError.Append(err)
	}

	if request.ActualLrpNetInfo == nil {
		validationError = validationError.Append(ErrInvalidField{"actual_lrp_net_info"})
	} else if err := request.ActualLrpNetInfo.Validate(); err != nil {
		validationError = validationError.Append(err)
	}
	if !validationError.Empty() {
		return validationError
	}

	return nil
}
