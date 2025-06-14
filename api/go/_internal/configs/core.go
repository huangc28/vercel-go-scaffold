package configs

import (
	"strings"

	"github.com/spf13/viper"
)

type ENV string

var (
	Production ENV = "production"
	Preview    ENV = "preview"
	Dev        ENV = "development"
	Test       ENV = "test"
)

type Config struct {
	ENV ENV `mapstructure:"vercel_env"`
	DB  struct {
		Host     string `mapstructure:"host"`
		Port     string `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		Name     string `mapstructure:"name"`
	} `mapstructure:"db"`
}

func NewConfig(vp *viper.Viper) (*Config, error) {
	var cfg Config
	if err := vp.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func NewViper() *viper.Viper {
	vp := viper.New()
	vp.AutomaticEnv()
	vp.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	vp.SetDefault("vercel_env", Dev)

	vp.SetDefault("db.host", "")
	vp.SetDefault("db.port", "")
	vp.SetDefault("db.user", "")
	vp.SetDefault("db.password", "")
	vp.SetDefault("db.name", "")

	return vp
}
