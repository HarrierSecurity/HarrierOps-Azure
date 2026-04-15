package render

import "encoding/json"

func JSON(payload any) (string, error) {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data) + "\n", nil
}
