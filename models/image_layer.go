package models

import (
	"strings"
)

func (l *ImageLayer) Validate() error {
	var validationError ValidationError

	if l.GetUrl() == "" {
		validationError = validationError.Append(ErrInvalidField{"url"})
	}

	if l.GetDestinationPath() == "" {
		validationError = validationError.Append(ErrInvalidField{"destination_path"})
	}

	if l.GetMediaType() == "" {
		validationError = validationError.Append(ErrInvalidField{"media_type"})
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

func convertImageLayersToDownloadActionsAndCachedDependencies(layers []*ImageLayer, legacyDownloadUser string, existingCachedDependencies []*CachedDependency, existingAction *Action) ([]*CachedDependency, *Action) {
	cachedDependencies := []*CachedDependency{}
	downloadActions := []ActionInterface{}

	for _, layer := range layers {
		if layer.LayerType == ImageLayer_Shared {
			c := &CachedDependency{
				Name:              layer.Name,
				From:              layer.Url,
				To:                layer.DestinationPath,
				ChecksumAlgorithm: layer.ChecksumAlgorithm,
				ChecksumValue:     layer.ChecksumValue,
			}

			if layer.ChecksumValue == "" {
				c.CacheKey = layer.Url
			} else {
				c.CacheKey = layer.ChecksumAlgorithm + ":" + layer.ChecksumValue
			}

			cachedDependencies = append(cachedDependencies, c)
		}

		if layer.LayerType == ImageLayer_Exclusive {
			downloadActions = append(downloadActions, &DownloadAction{
				Artifact:          layer.Name,
				From:              layer.Url,
				To:                layer.DestinationPath,
				CacheKey:          layer.ChecksumAlgorithm + ":" + layer.ChecksumValue, // checksum required for exclusive layers
				User:              legacyDownloadUser,
				ChecksumAlgorithm: layer.ChecksumAlgorithm,
				ChecksumValue:     layer.ChecksumValue,
			})
		}
	}

	cachedDependencies = append(cachedDependencies, existingCachedDependencies...)

	var action *Action
	if len(downloadActions) > 0 {
		parallelDownloadActions := Parallel(downloadActions...)
		if existingAction != nil {
			action = WrapAction(Serial(parallelDownloadActions, UnwrapAction(existingAction)))
		} else {
			action = WrapAction(Serial(parallelDownloadActions))
		}
	} else {
		action = existingAction
	}

	return cachedDependencies, action
}
