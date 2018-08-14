package models

import (
	"strings"

	"code.cloudfoundry.org/bbs/format"
)

func (l *ImageLayer) Validate() error {
	var validationError ValidationError

	if l.GetUrl() == "" {
		validationError = validationError.Append(ErrInvalidField{"url"})
	}

	if l.GetDestinationPath() == "" {
		validationError = validationError.Append(ErrInvalidField{"destination_path"})
	}

	if l.GetContentType() == "" {
		validationError = validationError.Append(ErrInvalidField{"content_type"})
	}

	if l.GetChecksumValue() != "" && l.GetChecksumAlgorithm() == "" {
		validationError = validationError.Append(ErrInvalidField{"checksum algorithm"})
	}

	if l.GetChecksumValue() == "" && l.GetChecksumAlgorithm() != "" {
		validationError = validationError.Append(ErrInvalidField{"checksum value"})
	}

	if l.GetChecksumValue() != "" && l.GetChecksumAlgorithm() != "" {
		if !contains([]string{"md5", "sha1", "sha256"}, strings.ToLower(l.GetChecksumAlgorithm())) {
			validationError = validationError.Append(ErrInvalidField{"invalid algorithm"})
		}
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func validateImageLayers(layers []*ImageLayer) ValidationError {
	var validationError ValidationError

	if len(layers) > 0 {
		for _, layer := range layers {
			err := layer.Validate()
			if err != nil {
				validationError = validationError.Append(ErrInvalidField{"image_layer"})
				validationError = validationError.Append(err)
			}
		}
	}

	return validationError
}

func (l *ImageLayer) Version() format.Version {
	return format.V0
}
