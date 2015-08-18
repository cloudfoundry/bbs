package models

func (request *UpsertDomainRequest) Validate() error {
	var validationError ValidationError

	if request.Domain == "" {
		return validationError.Append(ErrInvalidField{"domain"})
	}

	return nil
}
