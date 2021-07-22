/*
 * This file is part of Refractor.
 *
 * Refractor is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package conf

import "github.com/spf13/viper"

// Config stores all configuration of Refractor.
// The values are read in by Viper from a config file or from environment variables.
type Config struct {
	DBDriver            string `mapstructure:"DB_DRIVER"`
	DBSource            string `mapstructure:"DB_SOURCE"`
	APIBind             string `mapstructure:"API_BIND"`
	KratosPublic        string `mapstructure:"KRATOS_PUBLIC_ROOT"`
	KratosAdmin         string `mapstructure:"KRATOS_ADMIN_ROOT"`
	Mode                string `mapstructure:"MODE"`
	InitialUserEmail    string `mapstructure:"INITIAL_USER_EMAIL"`
	InitialUserUsername string `mapstructure:"INITIAL_USER_USERNAME"`
	SmtpConnectionUri   string `mapstructure:"SMTP_CONNECTION_URI"`
}

// LoadConfig reads configuration from a file or environment variables.
func LoadConfig(path string) (*Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.SetDefault("MODE", "production")

	// Tells viper to automatically override values read from a config file with values in env variables if they exist.
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	config := &Config{}
	err := viper.Unmarshal(config)

	return config, err
}
