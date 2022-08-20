package cfg

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

const (
	// env var names
	EnvVarNameTest    string = "TEST_ENV"
	EnvVarNameRuntime string = "RUNTIME_ENV"

	// env var test values
	EnvValTestDefault             string = "N/A"
	EnvValTestUnit                string = "UNIT"
	EnvValTestPostgresE2E         string = "PostgresE2E"
	EnvValTestPostgresIntegration string = "PostgresIntegration"
	EnvValTestSqliteIntegration   string = "SqliteIntegration"

	// env var runtime values
	EnvValRuntimeDev     string = "DEV"
	EnvValRuntimeStaging string = "STAGING"
	EnvValRuntimeProd    string = "PROD"
)

var (
	// Test environemnt types. By default it is unit test environment
	testEnv = map[string]struct{}{
		EnvValTestDefault:             {},
		EnvValTestUnit:                {},
		EnvValTestPostgresE2E:         {},
		EnvValTestPostgresIntegration: {},
		EnvValTestSqliteIntegration:   {},
	}

	// Deployment environments. By default it is developer environment
	runtimeEnv = map[string]struct{}{
		EnvValRuntimeDev:     {},
		EnvValRuntimeStaging: {},
		EnvValRuntimeProd:    {},
	}
)

func ValidateEnv() error {
	testVal := os.Getenv(EnvVarNameTest)
	if testVal == "" {
		os.Setenv(EnvVarNameTest, EnvValTestDefault)
	} else if _, contains := testEnv[testVal]; !contains {
		return fmt.Errorf("env variable %s assigned unexpected value: %s", EnvVarNameTest, testVal)
	}
	runtimeVal := os.Getenv(EnvVarNameRuntime)
	if runtimeVal == "" {
		os.Setenv(EnvVarNameRuntime, EnvValRuntimeDev)
		runtimeVal = EnvValRuntimeDev
	} else if _, contains := runtimeEnv[runtimeVal]; !contains {
		return fmt.Errorf("env variable %s assigned unexpected value: %s", EnvVarNameRuntime, runtimeVal)
	}
	log.Infof("Running in %s environment", runtimeVal)
	return nil
}
