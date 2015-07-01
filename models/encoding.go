package models

import "encoding/json"

func FromJSON(payload []byte, v Validator) error {
	err := json.Unmarshal(payload, v)
	if err != nil {
		return err
	}
	return v.Validate()
}
