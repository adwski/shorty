package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func newFromFile(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open config file: %w", err)
	}
	cfgBytes, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("cannot read config file: %w", err)
	}
	var cfg Config
	if err = json.Unmarshal(cfgBytes, &cfg); err != nil {
		return nil, fmt.Errorf("cannot unmarshal config: %w", err)
	}
	return &cfg, nil
}
