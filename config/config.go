package config

import (
	"github.com/spf13/viper"
	"log"
	"net/url"
)

const (
	ISOShortDateFormat = "2006-01-02" //  ISO 8601: YYYY-MM-DD
)

func InitConfig() {
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	viper.SetConfigType("yaml")
	viper.SetConfigName("default")

	err := viper.ReadInConfig()
	if err != nil {
		log.Panicf("Could not load default config file: %s\n", err)
	}

	err = viper.BindEnv("GO_ENV")
	if err != nil {
		log.Panicf("Could not bind GO_ENV variable: %s\n", err)
	}
	env := viper.GetString("GO_ENV")
	if env == "" {
		env = "dev"
		viper.Set("GO_ENV", env)
	}

	viper.SetConfigName(env)
	err = viper.MergeInConfig()
	if err != nil {
		log.Panicf("Error in environment [%s] config file: %s\n", env, err)
	}
}

// Create a URL to the client/application site
func CreateUrl(path string, queryString string) string {
	hostname := viper.GetString("application.applicationDomain")
	port := viper.GetString("application.applicationDomainPort")
	domainSecure := viper.GetBool("application.domainSecure")

	scheme := "https"
	if !domainSecure {
		scheme = "http"
	}

	if port != "" {
		hostname += ":" + port
	}

	url := url.URL{
		Scheme:   scheme,
		Host:     hostname,
		Path:     path,
		RawQuery: queryString,
	}

	return url.String()
}
