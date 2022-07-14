package config

import (
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Debug           bool   `mapstructure:"debug" yaml:"debug"`
	Token           string `mapstructure:"token" yaml:"token"`
	FetchingTimeout int    `mapstructure:"fetching_timeout" yaml:"fetching_timeout"`
	DBPath          string `mapstructure:"db_path" yaml:"db_path"`
}

type DB struct {
	Path string `yaml:"path"`
}

func Load() (*Config, error) {
	bindEnvs(Config{})

	cfg := Config{
		FetchingTimeout: 60,
		DBPath:          "var/pidor.db",
	}
	err := viper.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func bindEnvs(iface interface{}, parts ...string) {
	ifv := reflect.ValueOf(iface)
	ift := reflect.TypeOf(iface)
	for i := 0; i < ift.NumField(); i++ {
		v := ifv.Field(i)
		t := ift.Field(i)
		tv, ok := t.Tag.Lookup("mapstructure")
		if !ok {
			continue
		}
		switch v.Kind() {
		case reflect.Struct:
			bindEnvs(v.Interface(), append(parts, tv)...)
		default:
			viper.BindEnv(strings.Join(append(parts, tv), "."))
		}
	}
}
