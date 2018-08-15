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

	if (l.GetChecksumValue() != "" || l.GetLayerType() == ImageLayer_Exclusive) && l.GetChecksumAlgorithm() == "" {
		validationError = validationError.Append(ErrInvalidField{"checksum algorithm"})
	}

	if (l.GetChecksumAlgorithm() != "" || l.GetLayerType() == ImageLayer_Exclusive) && l.GetChecksumValue() == "" {
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

func validateImageLayers(layers []*ImageLayer, legacyDownloadUser string) ValidationError {
	var validationError ValidationError

	requiresLegacyDownloadUser := false
	if len(layers) > 0 {
		for _, layer := range layers {
			err := layer.Validate()
			if err != nil {
				validationError = validationError.Append(ErrInvalidField{"image_layer"})
				validationError = validationError.Append(err)
			}

			if layer.LayerType == ImageLayer_Exclusive {
				requiresLegacyDownloadUser = true
			}
		}
	}

	if requiresLegacyDownloadUser && legacyDownloadUser == "" {
		validationError = validationError.Append(ErrInvalidField{"legacy_download_user"})
	}

	return validationError
}

func (l *ImageLayer) Version() format.Version {
	return format.V0
}
