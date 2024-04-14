package config

import (
	"github.com/BelyaevEI/backend-trainee-assignment-2024/internal/models"
	"github.com/spf13/viper"
)

type Config struct {
	DSN            string `mapstructure:"DSN"`            // DSN for postgreSQL
	Host           string `mapstructure:"Host"`           // Server host
	Port           string `mapstructure:"Port"`           // Server port
	RedisAddr      string `mapstructure:"RedisAddr"`      // Addres for redis
	RedisPassword  string `mapstructure:"RedisPassword"`  // Password for redis
	AdminToken     string `mapstructure:"AdminToken"`     // Admin token for authorisation
	UserToken      string `mapstructure:"UserToken"`      // User token for authorisation
	AdminSecretKey string `mapstructure:"AdminSecretKey"` // Secret key for validation admin token
	UserSecretKey  string `mapstructure:"UserSecretKey"`  // Secret key for validation user token
	Rabbit         string `mapstructure:"Rabbit"`         // DSN for RabbitMQ
}

// Reading config file for setting application
func LoadConfig(path string) (Config, error) {

	conf := Config{}

	viper.AddConfigPath(path)
	viper.SetConfigName(models.ConfigName)
	viper.SetConfigType(models.ConfigType)
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return conf, err
	}

	err = viper.Unmarshal(&conf)
	if err != nil {
		return conf, err
	}

	return conf, nil
}
