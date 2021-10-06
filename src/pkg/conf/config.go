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

import (
	"Refractor/pkg/env"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
)

// Config stores all configuration of Refractor.
// The values are read in by Viper from a config file or from environment variables.
type Config struct {
	DBDriver            string `mapstructure:"DB_DRIVER"`
	DBSource            string `mapstructure:"DB_SOURCE"`
	KratosPublic        string `mapstructure:"KRATOS_PUBLIC_ROOT"`
	KratosAdmin         string `mapstructure:"KRATOS_ADMIN_ROOT"`
	FrontendRoot        string `mapstructure:"FRONTEND_ROOT"`
	Mode                string `mapstructure:"MODE"`
	InitialUserEmail    string `mapstructure:"INITIAL_USER_EMAIL"`
	InitialUserUsername string `mapstructure:"INITIAL_USER_USERNAME"`
	SmtpConnectionUri   string `mapstructure:"SMTP_CONNECTION_URI"`
	EncryptionKey       string `mapstructure:"ENCRYPTION_KEY"`
}

// LoadConfig reads configuration from a file or environment variables.
func LoadConfig() (*Config, error) {
	if err := godotenv.Load("../app.env"); err == nil {
		log.Println("Environment variables loaded from app.env file")
	}

	err := env.RequireEnv("DB_DRIVER").
		RequireEnv("DB_SOURCE").
		RequireEnv("KRATOS_PUBLIC_ROOT").
		RequireEnv("KRATOS_ADMIN_ROOT").
		RequireEnv("FRONTEND_ROOT").
		RequireEnv("SMTP_CONNECTION_URI").
		RequireEnv("ENCRYPTION_KEY").
		RequireEnv("INITIAL_USER_EMAIL").
		RequireEnv("INITIAL_USER_USERNAME").
		GetError()
	if err != nil {
		return nil, err
	}

	config := &Config{
		DBDriver:            os.Getenv("DB_DRIVER"),
		DBSource:            os.Getenv("DB_SOURCE"),
		KratosPublic:        os.Getenv("KRATOS_PUBLIC_ROOT"),
		KratosAdmin:         os.Getenv("KRATOS_ADMIN_ROOT"),
		FrontendRoot:        os.Getenv("FRONTEND_ROOT"),
		InitialUserEmail:    os.Getenv("INITIAL_USER_EMAIL"),
		InitialUserUsername: os.Getenv("INITIAL_USER_USERNAME"),
		SmtpConnectionUri:   os.Getenv("SMTP_CONNECTION_URI"),
		EncryptionKey:       os.Getenv("ENCRYPTION_KEY"),
	}

	if len(config.EncryptionKey) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes")
	}

	if os.Getenv("MODE") == "dev" {
		config.Mode = "dev"
	} else {
		config.Mode = "prod"
	}

	return config, err
}
