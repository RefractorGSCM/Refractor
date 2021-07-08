package conf

import "github.com/spf13/viper"

// Config stores all configuration of Refractor.
// The values are read in by Viper from a config file or from environment variables.
type Config struct {
	DBDriver     string `mapstructure:"DB_DRIVER"`
	DBSource     string `mapstructure:"DB_SOURCE"`
	APIBind      string `mapstructure:"API_BIND"`
	KratosPublic string `mapstructure:"KRATOS_PUBLIC_ROOT"`
	KratosAdmin  string `mapstructure:"KRATOS_ADMIN_ROOT"`
}

// LoadConfig reads configuration from a file or environment variables.
func LoadConfig(path string) (*Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	// Tells viper to automatically override values read from a config file with values in env variables if they exist.
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	config := &Config{}
	err := viper.Unmarshal(config)

	return config, err
}
