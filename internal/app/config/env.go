package config

import (
	"fmt"
	"os"
	"strconv"
)

func mergeEnvs(cfg *Config) error {
	envOverride("SERVER_ADDRESS", &cfg.ListenAddr)
	envOverride("PPROF_ADDRESS", &cfg.PprofServerAddr)
	envOverride("BASE_URL", &cfg.BaseURL)
	envOverride("FILE_STORAGE_PATH", &cfg.Storage.FileStoragePath)
	envOverride("DATABASE_DSN", &cfg.Storage.DatabaseDSN)
	envOverride("JWT_SECRET", &cfg.JWTSecret)
	if err := envOverrideBool("ENABLE_HTTPS", &cfg.TLS.Enable); err != nil {
		return err
	}
	return nil
}

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
