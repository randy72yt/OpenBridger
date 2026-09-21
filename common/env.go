package common

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func GetEnvOrDefault(env string, defaultValue int) int {
	if env == "" || os.Getenv(env) == "" {
		return defaultValue
	}
	num, err := strconv.Atoi(os.Getenv(env))
	if err != nil {
		SysError(fmt.Sprintf("failed to parse %s: %s, using default value: %d", env, err.Error(), defaultValue))
		return defaultValue
	}
	return num
}

func GetEnvOrDefaultString(env string, defaultValue string) string {
	if env == "" || os.Getenv(env) == "" {
		return defaultValue
	}
	return os.Getenv(env)
}

func GetEnvOrDefaultBool(env string, defaultValue bool) bool {
	if env == "" || os.Getenv(env) == "" {
		return defaultValue
	}
	b, err := strconv.ParseBool(os.Getenv(env))
	if err != nil {
		SysError(fmt.Sprintf("failed to parse %s: %s, using default value: %t", env, err.Error(), defaultValue))
		return defaultValue
	}
	return b
}

// OnlinePaymentEnabled preserves existing deployments when unset. An explicit
// value must be "true" to permit new orders; typos fail closed.
func OnlinePaymentEnabled() bool {
	value, configured := os.LookupEnv("OPENBRIDGER_ONLINE_PAYMENT_ENABLED")
	if !configured {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(value), "true")
}
