package config

import (
	"os"

	"github.com/spf13/viper"
)

// LoadProvider returns a configured viper instance
func LoadProvider(appName string) *viper.Viper {
	return readViperConfig(appName)
}

func readViperConfig(appName string) *viper.Viper {
	v := viper.New()
	v.SetEnvPrefix(appName)
	v.AutomaticEnv()

	// global defaults

	var logLevelEnv = os.Getenv("BITCOIN_KEYCHAIN_LOG_LEVEL")
	var logLevel = "info"
	if logLevelEnv != "" {
		logLevel = logLevelEnv
	}

	var jsonLogsEnv = os.Getenv("BITCOIN_KEYCHAIN_JSON_LOGS")
	var jsonLogs = true
	if jsonLogsEnv == "false" {
		jsonLogs = false
	}

	v.SetDefault("json_logs", jsonLogs)
	v.SetDefault("loglevel", logLevel)

	return v
}
