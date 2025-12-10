package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type SupportedDatabaseType string

const (
	DatabaseTypePostgres   SupportedDatabaseType = "postgres"
	DatabaseTypeClickhouse SupportedDatabaseType = "clickhouse"
)

type EnvironmentType string

const (
	EnvironmentTypeDev  EnvironmentType = "dev"
	EnvironmentTypeProd EnvironmentType = "prod"
)

type AppConfig struct {
	EnvironmentType EnvironmentType       `mapstructure:"environment_type" yaml:"environment_type"`
	Port            string                `mapstructure:"port" yaml:"port"`
	DatabaseType    SupportedDatabaseType `mapstructure:"database_type" yaml:"database_type"`
	PostgresSQLUrl  string                `mapstructure:"postgres_url" yaml:"postgres_url"`
	ClickhouseUrl   string                `mapstructure:"clickhouse_url" yaml:"clickhouse_url"`
}

func Read() *AppConfig {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$PWD/config") // call multiple times to add many search paths
	viper.AddConfigPath(".")
	viper.AddConfigPath("/config")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	var appConfig AppConfig
	if err := viper.Unmarshal(&appConfig); err != nil {
		panic(fmt.Errorf("fatal error unmarshalling config: %w", err))
	}

	return &appConfig
}
