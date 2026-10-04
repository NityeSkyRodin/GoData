package exporter

import (
	"encoding/json"
	"os"
)

func ParseToJSONFile(data any, filename string) error {
	fileData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, fileData, 0644)
}