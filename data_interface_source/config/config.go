package config

import (
	"encoding/json"
	"os"
)

type DynamicMapping struct {
	Database        string            `json:"database"`
	Table           string            `json:"table"`
	Columns         map[string]string `json:"columns"` // optional
	CreateIfMissing bool              `json:"create_if_missing"`
}

type DynamicConfig struct {
	Mappings map[string]DynamicMapping `json:"mappings"`
}

var DynamicMappings DynamicConfig

func LoadDynamicConfig(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(&DynamicMappings)
}
