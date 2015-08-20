package models

func (request *DesiredLRPsRequest) Validate() error {
	return nil
}

func (request *DesiredLRPByProcessGuidRequest) Validate() error {
	var validationError ValidationError

	if request.ProcessGuid == "" {
		validationError = validationError.Append(ErrInvalidField{"process_guid"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}
