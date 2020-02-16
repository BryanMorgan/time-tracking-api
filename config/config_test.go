package config

import (
	"testing"

	"github.com/spf13/viper"
)

func TestConfig(t *testing.T) {
	InitConfig()

	hostName := viper.GetString("application.hostname")
	if hostName == "" {
		t.Errorf("application.hostname not found in config")
	}

	missing := viper.GetString("application.this.does.not.exist")
	if missing != "" || viper.IsSet("this.does.not.exist") {
		t.Errorf("invalid configuration found")
	}

}
