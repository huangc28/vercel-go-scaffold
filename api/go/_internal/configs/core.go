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

	AWS struct {
		AccessKey        string `mapstructure:"access_key_id"`
		SecretKey        string `mapstructure:"secret_access_key"`
		S3BucketRegion   string `mapstructure:"s3_bucket_region"`
		S3SnapshotBucket string `mapstructure:"s3_snapshot_bucket"`
	} `mapstructure:"aws"`

	Clerk struct {
		SecretKey string `mapstructure:"secret_key"`
	} `mapstructure:"clerk"`

	Starburst struct {
		Host     string `mapstructure:"host"`
		Port     string `mapstructure:"port"`
		Catalog  string `mapstructure:"catalog"`
		Schema   string `mapstructure:"schema"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
	} `mapstructure:"starburst"`

	Inngest struct {
		EventKey string `mapstructure:"event_key"`
		AppID    string `mapstructure:"app_id"`
	} `mapstructure:"inngest"`
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

	vp.SetDefault("aws.access_key_id", "")
	vp.SetDefault("aws.secret_access_key", "")
	vp.SetDefault("aws.s3_bucket_region", "")
	vp.SetDefault("aws.s3_snapshot_bucket", "")

	vp.SetDefault("clerk.secret_key", "")

	vp.SetDefault("starburst.host", "")
	vp.SetDefault("starburst.port", "")
	vp.SetDefault("starburst.catalog", "")
	vp.SetDefault("starburst.schema", "")
	vp.SetDefault("starburst.user", "")
	vp.SetDefault("starburst.password", "")

	vp.SetDefault("inngest.event_key", "")
	vp.SetDefault("inngest.app_id", "")

	return vp
}
