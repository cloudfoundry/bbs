package models

import "github.com/cloudfoundry-incubator/bbs/format"

func (a *CachedDependency) Validate() error {
	var validationError ValidationError

	if a.GetFrom() == "" {
		validationError = validationError.Append(ErrInvalidField{"from"})
	}

	if a.GetTo() == "" {
		validationError = validationError.Append(ErrInvalidField{"to"})
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func validateCachedDependencies(
	cachedDependencies []*CachedDependency,
	legacyDownloadUser string,
) ValidationError {
	var validationError ValidationError

	if len(cachedDependencies) > 0 {
		if legacyDownloadUser == "" {
			validationError = validationError.Append(ErrInvalidField{"legacy_download_user"})
		}

		for _, cacheDep := range cachedDependencies {
			err := cacheDep.Validate()
			if err != nil {
				validationError = validationError.Append(ErrInvalidField{"cached_dependency"})
				validationError = validationError.Append(err)
			}
		}
	}

	return validationError
}

func (c *CachedDependency) MigrateFromVersion(v format.Version) error {
	return nil
}

func (c *CachedDependency) Version() format.Version {
	return format.V0
}
