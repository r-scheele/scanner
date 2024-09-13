package config

import (
	"os"
	"strings"
)

const (
	devEnv = "development"
)

// AppCfg - the application configuration
var AppCfg map[string]string

// AppEnv - the application env (development|production)
var AppEnv string

// IsDevEnv - return true if running in development environment
// i.e. if DEPLOY_ENV is set to 'development'
func IsDevEnv() bool {
	return AppEnv == devEnv
}

// LoadAppConfig - Loads app config based on DEPLOY_ENV
func LoadAppConfig() error {
	var appEnvOk bool
	AppEnv, appEnvOk = os.LookupEnv("DEPLOY_ENV")
	if !appEnvOk {
		AppEnv = devEnv
	}

	// Initialize AppCfg from system environment variables
	AppCfg = make(map[string]string)
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) == 2 {
			AppCfg[pair[0]] = pair[1]
		}
	}

	return nil
}

// GetConfigString - return string value of given config key
// defaultVal is returned if key not present in config
func GetConfigString(key string, defaultVal string) string {
	if val, ok := AppCfg[key]; ok && val != "" {
		return val
	}
	return defaultVal
}

// GetConfigStringSlice - return a string slice value of
// given config key. An empty slice is returned if key not
// present in config
func GetConfigStringSlice(key string) []string {
	if val, ok := AppCfg[key]; ok && val != "" {
		return strings.Split(val, ",")
	}
	return []string{}
}
