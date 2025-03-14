package models

func (request *ProtoDesireTaskRequest) Validate() error {
	return request.FromProto().Validate()
}

func (req *DesireTaskRequest) Validate() error {
	var validationError ValidationError

	if !taskGuidPattern.MatchString(req.TaskGuid) {
		validationError = validationError.Append(ErrInvalidField{"task_guid"})
	}

	if req.Domain == "" {
		validationError = validationError.Append(ErrInvalidField{"domain"})
	}

	if req.TaskDefinition == nil {
		validationError = validationError.Append(ErrInvalidField{"task_definition"})
	} else if defErr := req.TaskDefinition.Validate(); defErr != nil {
		validationError = validationError.Append(defErr)
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (request *ProtoStartTaskRequest) Validate() error {
	return request.FromProto().Validate()
}

func (req *StartTaskRequest) Validate() error {
	var validationError ValidationError

	if !taskGuidPattern.MatchString(req.TaskGuid) {
		validationError = validationError.Append(ErrInvalidField{"task_guid"})
	}
	if req.CellId == "" {
		validationError = validationError.Append(ErrInvalidField{"cell_id"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (request *ProtoCompleteTaskRequest) Validate() error {
	return request.FromProto().Validate()
}

func (req *CompleteTaskRequest) Validate() error {
	var validationError ValidationError

	if !taskGuidPattern.MatchString(req.TaskGuid) {
		validationError = validationError.Append(ErrInvalidField{"task_guid"})
	}
	if req.CellId == "" {
		validationError = validationError.Append(ErrInvalidField{"cell_id"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (request *ProtoFailTaskRequest) Validate() error {
	return request.FromProto().Validate()
}

func (req *FailTaskRequest) Validate() error {
	var validationError ValidationError

	if !taskGuidPattern.MatchString(req.TaskGuid) {
		validationError = validationError.Append(ErrInvalidField{"task_guid"})
	}
	if req.FailureReason == "" {
		validationError = validationError.Append(ErrInvalidField{"failure_reason"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (request *ProtoRejectTaskRequest) Validate() error {
	return request.FromProto().Validate()
}

func (req *RejectTaskRequest) Validate() error {
	var validationError ValidationError

	if !taskGuidPattern.MatchString(req.TaskGuid) {
		validationError = validationError.Append(ErrInvalidField{"task_guid"})
	}
	if req.RejectionReason == "" {
		validationError = validationError.Append(ErrInvalidField{"failure_reason"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (request *ProtoTasksRequest) Validate() error {
	return request.FromProto().Validate()
}

func (req *TasksRequest) Validate() error {
	return nil
}

func (request *ProtoTaskByGuidRequest) Validate() error {
	return request.FromProto().Validate()
}

func (request *TaskByGuidRequest) Validate() error {
	var validationError ValidationError

	if request.TaskGuid == "" {
		validationError = validationError.Append(ErrInvalidField{"task_guid"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (request *ProtoTaskGuidRequest) Validate() error {
	return request.FromProto().Validate()
}

func (request *TaskGuidRequest) Validate() error {
	var validationError ValidationError

	if request.TaskGuid == "" {
		validationError = validationError.Append(ErrInvalidField{"task_guid"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}
