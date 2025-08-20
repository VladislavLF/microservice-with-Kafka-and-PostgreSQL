package config

import (
	"time"

	"github.com/spf13/viper"
)

type Server struct {
	Addr         string        `mapstructure:"addr"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

type DB struct {
	DSN             string        `mapstructure:"dsn"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type Kafka struct {
	Brokers        []string      `mapstructure:"brokers"`
	Topic          string        `mapstructure:"topic"`
	GroupID        string        `mapstructure:"group_id"`
	MinBytes       int           `mapstructure:"min_bytes"`
	MaxBytes       int           `mapstructure:"max_bytes"`
	CommitInterval time.Duration `mapstructure:"commit_interval"`
}

type Cache struct {
	Capacity int           `mapstructure:"capacity"`
	TTL      time.Duration `mapstructure:"ttl"`
}

type UI struct {
	Enable    bool   `mapstructure:"enable"`
	StaticDir string `mapstructure:"static_dir"`
}

type Config struct {
	Server Server `mapstructure:"server"`
	DB     DB     `mapstructure:"db"`
	Kafka  Kafka  `mapstructure:"kafka"`
	Cache  Cache  `mapstructure:"cache"`
	UI     UI     `mapstructure:"ui"`
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AutomaticEnv()
	v.SetEnvPrefix("ORDERS")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, err
	}
	return &c, nil
}
