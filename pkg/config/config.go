package config

import (
	"reflect"

	"github.com/spf13/viper"
)

const EncryptedFlag = "encrypted"

type Config struct {
	// Global config
	LogLevel string `mapstructure:"log_level"`

	// ProxyServer config
	ProxyAddress string `mapstructure:"proxy_address"`
	CertFile     string `mapstructure:"cert_file"`
	KeyFile      string `mapstructure:"key_file"`
}

func New() *Config {
	return &Config{
		LogLevel:     "info",
		ProxyAddress: ":8000",
		CertFile:     "/certs/tls.crt",
		KeyFile:      "/certs/tls.key",
	}
}

func BindEnv(v *viper.Viper, key string) {
	t := reflect.TypeOf(Config{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		v.BindEnv(field.Tag.Get(key))
	}
}
