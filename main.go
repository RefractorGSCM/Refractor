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

package main

import (
	"Refractor/auth"
	"Refractor/domain"
	"Refractor/games/mordhau"
	_authRepo "Refractor/internal/auth/repos/kratos"
	_authorizer "Refractor/internal/authorizer"
	_gameService "Refractor/internal/game/service"
	_groupHandler "Refractor/internal/group/delivery/http"
	_groupRepo "Refractor/internal/group/repos/postgres"
	_groupService "Refractor/internal/group/service"
	_serverHandler "Refractor/internal/server/delivery/http"
	_postgresServerRepo "Refractor/internal/server/repos/postgres"
	_serverService "Refractor/internal/server/service"
	"Refractor/pkg/api"
	"Refractor/pkg/api/middleware"
	"Refractor/pkg/conf"
	"Refractor/pkg/tmpl"
	"Refractor/platforms/playfab"
	"database/sql"
	"embed"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v4"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	kratos "github.com/ory/kratos-client-go"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"log"
	"strings"
	"time"
)

func main() {
	config, err := conf.LoadConfig(".")
	if err != nil {
		log.Fatalf("Could not load configuration. Error: %v", err)
	}

	logger, err := setupLogger(config.Mode)
	if err != nil {
		log.Fatalf("Could not set up logger. Error: %v", err)
	}

	db, _, err := setupDatabase(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatalf("Could not set up database. Error: %v", err)
	}

	apiServer, err := setupEchoAPI(logger, config)
	if err != nil {
		log.Fatalf("Could not set up API webserver. Error: %v", err)
	}
	apiGroup := apiServer.Group("/api/v1")

	kratosClient := setupKratos()

	authServer, err := setupEchoPages(logger, kratosClient, config)
	if err != nil {
		log.Fatalf("Could not set up auth webserver. Error: %v", err)
	}

	// Set up application components
	authRepo := _authRepo.NewAuthRepo(config)

	users, err := authRepo.GetAllUsers()
	if err != nil {
		log.Fatalf("Could not check if a user currently exists. Error: %v", err)
	}

	// If no users exist, we create one from the initial user config variables.
	if len(users) < 1 {
		if err := SetupInitialUser(authRepo, config); err != nil {
			log.Fatalf("Could not create initial user. Error: %v", err)
		}

		log.Printf("Initial superadmin user (%s) has been created!", config.InitialUserUsername)
	}

	protectMiddleware := middleware.NewAPIProtectMiddleware(config)

	groupRepo, err := _groupRepo.NewGroupRepo(db, logger)
	if err != nil {
		log.Fatalf("Could not set up group repository. Error: %v", err)
	}

	authorizer := _authorizer.NewAuthorizer(groupRepo, logger)

	groupService := _groupService.NewGroupService(groupRepo, time.Second*2)
	_groupHandler.ApplyGroupHandler(apiGroup, groupService, authorizer, protectMiddleware)

	gameService := _gameService.NewGameService()
	gameService.AddGame(mordhau.NewMordhauGame(playfab.NewPlayfabPlatform()))

	serverRepo := _postgresServerRepo.NewServerRepo(db, logger)
	serverService := _serverService.NewServerService(serverRepo, time.Second*2)
	_serverHandler.ApplyServerHandler(apiGroup, serverService, authorizer, protectMiddleware)

	// Setup complete. Begin serving requests.
	logger.Info("Setup complete!")

	go func() {
		log.Fatal(authServer.Start(":4455"))
	}()

	log.Fatal(apiServer.Start(":4000"))
}

func setupLogger(mode string) (*zap.Logger, error) {
	var logger *zap.Logger
	var err error

	if mode == "dev" {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}

	return logger, err
}

func setupDatabase(dbDriver, dbSource string) (*sql.DB, string, error) {
	switch dbDriver {
	case "postgres":
		trimmedURI := strings.Replace(dbSource, "postgres://", "", 1)
		db, err := setupPostgres(trimmedURI)
		return db, "postgres", err
	default:
		return nil, "", fmt.Errorf("unsupported database driver: %s", dbDriver)
	}
}

// Embed migration files into compiled binary for portability
//go:embed migrations/*.sql
var migrationFS embed.FS

func setupPostgres(dbURI string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbURI)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	// Set up migrations
	dfs, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		return nil, errors.Wrap(err, "Could not setup iofs migration source")
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, errors.Wrap(err, "Could not get driver instance")
	}

	m, err := migrate.NewWithInstance("iofs", dfs, "postgres", driver)
	if err != nil {
		return nil, errors.Wrap(err, "Could not create migration with database instance")
	}

	// Run migrations
	err = m.Up()
	if err == nil || err == migrate.ErrNoChange {
		version, _, _ := m.Version()
		log.Printf("Running database schema version %d", version)
	} else {
		return nil, errors.Wrap(err, "Could not run migrations")
	}

	return db, nil
}

func setupKratos() *kratos.APIClient {
	kratosConf := kratos.NewConfiguration()
	kratosConf.Scheme = "http"
	kratosConf.Host = "127.0.0.1:4433"

	kratosClient := kratos.NewAPIClient(kratosConf)

	return kratosClient
}

func setupEchoAPI(logger *zap.Logger, config *conf.Config) (*echo.Echo, error) {
	e := echo.New()
	e.HTTPErrorHandler = api.GetEchoErrorHandler(logger)

	e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowMethods:     []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
		AllowCredentials: true,
	}))

	return e, nil
}

func setupEchoPages(logger *zap.Logger, client *kratos.APIClient, config *conf.Config) (*echo.Echo, error) {
	e := echo.New()
	e.HTTPErrorHandler = api.GetEchoErrorHandler(logger)

	// Set up rendering of server side pages
	e.Renderer = tmpl.NewRenderer("./auth/templates/*.html", true)

	protect := middleware.NewBrowserProtectMiddleware(config)

	pagesHandler := auth.NewPublicHandlers(client, config)

	//echo.NotFoundHandler = pagesHandler.RootHandler
	kratosGroup := e.Group("/k")
	kratosGroup.GET("/login", pagesHandler.LoginHandler)
	kratosGroup.GET("/recovery", pagesHandler.RecoveryHandler)
	kratosGroup.GET("/settings", pagesHandler.SettingsHandler, protect)

	return e, nil
}

func SetupInitialUser(authRepo domain.AuthRepo, config *conf.Config) error {
	_, err := authRepo.CreateUser(&domain.Traits{
		Email:    config.InitialUserEmail,
		Username: config.InitialUserUsername,
	})
	if err != nil {
		return err
	}

	// TODO: Grant permissions

	return nil
}
