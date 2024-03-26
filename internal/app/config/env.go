package config

import (
	"fmt"
	"os"
	"strconv"
)

func envOverride(name string, param *string) {
	if param == nil {
		return
	}
	if val, ok := os.LookupEnv(name); ok {
		*param = val
	}
}

func envOverrideBool(name string, param *bool) error {
	if param == nil {
		return nil
	}
	val, ok := os.LookupEnv(name)
	if !ok {
		return nil
	}
	bVal, err := strconv.ParseBool(val)
	if err != nil {
		return fmt.Errorf("cannot parse bool value in env %s: %w", name, err)
	}
	*param = bVal
	return nil
}
