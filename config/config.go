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
	DB              DB     `mapstructure:"db" yaml:"db"`
}

type DB struct {
	Path string `yaml:"path"`
}

func Load() (*Config, error) {
	viper.SetDefault("fetching_timeout", 60)
	viper.SetDefault("db", DB{Path: "./var/pidor.db"})
	bindEnvs(Config{})

	var cfg Config
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
		tv, ok := t.Tag.Lookup("yaml")
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
