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

	if (l.GetDigestValue() != "" || l.GetLayerType() == ImageLayer_Exclusive) && l.GetDigestAlgorithm() == "" {
		validationError = validationError.Append(ErrInvalidField{"checksum algorithm"})
	}

	if (l.GetDigestAlgorithm() != "" || l.GetLayerType() == ImageLayer_Exclusive) && l.GetDigestValue() == "" {
		validationError = validationError.Append(ErrInvalidField{"checksum value"})
	}

	if l.GetDigestValue() != "" && l.GetDigestAlgorithm() != "" {
		if !contains([]string{"md5", "sha1", "sha256"}, strings.ToLower(l.GetDigestAlgorithm())) {
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

type ImageLayers []*ImageLayer

func (layers ImageLayers) FilterByType(layerType ImageLayer_Type) ImageLayers {
	var filtered ImageLayers

	for _, layer := range layers {
		if layer.GetLayerType() == layerType {
			filtered = append(filtered, layer)
		}
	}
	return filtered
}

func (layers ImageLayers) ToDownloadActions(legacyDownloadUser string, existingAction *Action) *Action {
	downloadActions := []ActionInterface{}

	for _, layer := range layers.FilterByType(ImageLayer_Exclusive) {
		downloadActions = append(downloadActions, &DownloadAction{
			Artifact:          layer.Name,
			From:              layer.Url,
			To:                layer.DestinationPath,
			CacheKey:          layer.DigestAlgorithm + ":" + layer.DigestValue, // digest required for exclusive layers
			User:              legacyDownloadUser,
			ChecksumAlgorithm: layer.DigestAlgorithm,
			ChecksumValue:     layer.DigestValue,
		})
	}

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

	return action
}

func (layers ImageLayers) ToCachedDependencies() []*CachedDependency {
	cachedDependencies := []*CachedDependency{}
	for _, layer := range layers.FilterByType(ImageLayer_Shared) {
		c := &CachedDependency{
			Name:              layer.Name,
			From:              layer.Url,
			To:                layer.DestinationPath,
			ChecksumAlgorithm: layer.DigestAlgorithm,
			ChecksumValue:     layer.DigestValue,
		}

		if layer.DigestValue == "" {
			c.CacheKey = layer.Url
		} else {
			c.CacheKey = layer.DigestAlgorithm + ":" + layer.DigestValue
		}

		cachedDependencies = append(cachedDependencies, c)
	}

	return cachedDependencies
}
