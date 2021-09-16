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
	_attachmentRepo "Refractor/internal/attachment/repos/postgres"
	_attachmentService "Refractor/internal/attachment/service"
	_authRepo "Refractor/internal/auth/repos/kratos"
	_authService "Refractor/internal/auth/service"
	_authorizer "Refractor/internal/authorizer"
	_gameService "Refractor/internal/game/service"
	_groupHandler "Refractor/internal/group/delivery/http"
	_groupRepo "Refractor/internal/group/repos/postgres"
	_groupService "Refractor/internal/group/service"
	_infractionHandler "Refractor/internal/infraction/delivery/http"
	_infractionRepo "Refractor/internal/infraction/repos/postgres"
	_infractionService "Refractor/internal/infraction/service"
	"Refractor/internal/mail/service"
	_playerHandler "Refractor/internal/player/delivery/http"
	_playerRepo "Refractor/internal/player/repos/postgres/player"
	_playerNameRepo "Refractor/internal/player/repos/postgres/playername"
	_playerService "Refractor/internal/player/service"
	_rconService "Refractor/internal/rcon/service"
	_searchHandler "Refractor/internal/search/delivery/http"
	_searchService "Refractor/internal/search/service"
	_serverHandler "Refractor/internal/server/delivery/http"
	_postgresServerRepo "Refractor/internal/server/repos/postgres"
	_serverService "Refractor/internal/server/service"
	_userHandler "Refractor/internal/user/delivery/http"
	_userRepo "Refractor/internal/user/repos/postgres"
	_userService "Refractor/internal/user/service"
	"Refractor/internal/watchdog"
	_websocketHandler "Refractor/internal/websocket/delivery/http"
	_websocketService "Refractor/internal/websocket/service"
	"Refractor/pkg/api"
	"Refractor/pkg/api/middleware"
	"Refractor/pkg/conf"
	"Refractor/pkg/perms"
	"Refractor/pkg/tmpl"
	"Refractor/platforms/playfab"
	"context"
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
	"net/url"
	"strings"
	"time"
)

func main() {
	config, err := conf.LoadConfig()
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

	kratosClient := setupKratos(config)

	// Set up application components
	userMetaRepo := _userRepo.NewUserRepo(db, logger)

	authServer, err := setupEchoPages(logger, kratosClient, config, userMetaRepo)
	if err != nil {
		log.Fatalf("Could not set up auth webserver. Error: %v", err)
	}

	mailService, err := service.NewMailService(config)
	if err != nil {
		log.Fatalf("Could not set up mail service. Error: %v", err)
	}

	authRepo := _authRepo.NewAuthRepo(config)
	authService := _authService.NewAuthService(authRepo, userMetaRepo, mailService, time.Second*2, logger)

	groupRepo, err := _groupRepo.NewGroupRepo(db, logger)
	if err != nil {
		log.Fatalf("Could not set up group repository. Error: %v", err)
	}

	users, err := authRepo.GetAllUsers(context.TODO())
	if err != nil {
		log.Fatalf("Could not check if a user currently exists. Error: %v", err)
	}

	// If no users exist, we create one from the initial user config variables.
	if len(users) < 1 {
		if err := SetupInitialUser(authService, groupRepo, config); err != nil {
			log.Fatalf("Could not create initial user. Error: %v", err)
		}

		log.Printf("Initial superadmin user (%s) has been created!", config.InitialUserUsername)
	}

	middlewareBundle := domain.Middleware{
		ProtectMiddleware:    middleware.NewAPIProtectMiddleware(config),
		ActivationMiddleware: middleware.NewActivationMiddleware(userMetaRepo),
	}

	authorizer := _authorizer.NewAuthorizer(groupRepo, logger)

	groupService := _groupService.NewGroupService(groupRepo, authorizer, time.Second*2, logger)
	_groupHandler.ApplyGroupHandler(apiGroup, groupService, authorizer, middlewareBundle, logger)

	gameService := _gameService.NewGameService()
	gameService.AddGame(mordhau.NewMordhauGame(playfab.NewPlayfabPlatform()))

	playerNameRepo := _playerNameRepo.NewPlayerNameRepo(db, logger)
	playerRepo := _playerRepo.NewPlayerRepo(db, playerNameRepo, logger)

	serverRepo := _postgresServerRepo.NewServerRepo(db, logger, config)
	serverService := _serverService.NewServerService(serverRepo, playerRepo, authorizer, time.Second*2, logger)
	_serverHandler.ApplyServerHandler(apiGroup, serverService, authorizer, middlewareBundle, logger)

	userService := _userService.NewUserService(userMetaRepo, authRepo, groupRepo, authorizer, time.Second*2, logger)
	_userHandler.ApplyUserHandler(apiGroup, userService, authService, authorizer, middlewareBundle, logger)

	rconService := _rconService.NewRCONService(logger, gameService)

	websocketService := _websocketService.NewWebsocketService(playerRepo, authorizer, time.Second*2, logger)
	go websocketService.StartPool()
	_websocketHandler.ApplyWebsocketHandler(apiServer, websocketService, middlewareBundle, logger)

	playerService := _playerService.NewPlayerService(playerRepo, playerNameRepo, time.Second*2, logger)
	_playerHandler.ApplyPlayerHandler(apiGroup, playerService, authorizer, middlewareBundle, logger)

	infractionRepo := _infractionRepo.NewInfractionRepo(db, logger)

	attachmentRepo := _attachmentRepo.NewAttachmentRepo(db, logger)
	attachmentService := _attachmentService.NewAttachmentService(attachmentRepo, infractionRepo, authorizer, time.Second*2, logger)

	infractionService := _infractionService.NewInfractionService(infractionRepo, playerRepo, serverRepo,
		attachmentRepo, userMetaRepo, authorizer, time.Second*2, logger)
	_infractionHandler.ApplyInfractionHandler(apiGroup, infractionService, attachmentService, authorizer, middlewareBundle, logger)

	searchService := _searchService.NewSearchService(playerRepo, time.Second*2, logger)
	_searchHandler.ApplySearchHandler(apiGroup, searchService, authorizer, middlewareBundle, logger)

	// Subscribe to events
	rconService.SubscribeJoin(playerService.HandlePlayerJoin)
	rconService.SubscribeQuit(playerService.HandlePlayerQuit)
	rconService.SubscribeJoin(websocketService.HandlePlayerJoin)
	rconService.SubscribeQuit(websocketService.HandlePlayerQuit)
	rconService.SubscribeJoin(serverService.HandlePlayerJoin)
	rconService.SubscribeQuit(serverService.HandlePlayerQuit)
	rconService.SubscribeServerStatus(serverService.HandleServerStatusChange)
	rconService.SubscribeServerStatus(websocketService.HandleServerStatusChange)

	// Connect RCON clients for all existing servers
	if err := SetupServerClients(rconService, serverService, logger); err != nil {
		log.Fatalf("Could not set up RCON server clients. Error: %v", err)
	}

	// Start watchdog
	go func() {
		err := watchdog.StartRCONServerWatchdog(rconService, serverService, logger)
		if err != nil {
			log.Fatalf("Could not start RCON server watchdog. Error: %v", err)
		}
	}()

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

func setupKratos(config *conf.Config) *kratos.APIClient {
	kratosConf := kratos.NewConfiguration()

	uri, err := url.Parse(config.KratosPublic)
	if err != nil {
		log.Fatalf("Invalid kratos public URI provided. Error: %v", err)
	}

	if config.Mode == "dev" {
		kratosConf.Scheme = "http"
	} else {
		kratosConf.Scheme = "https"
	}

	kratosConf.Host = uri.Host
	kratosConf.Debug = true

	kratosClient := kratos.NewAPIClient(kratosConf)

	return kratosClient
}

func setupEchoAPI(logger *zap.Logger, config *conf.Config) (*echo.Echo, error) {
	e := echo.New()
	e.HTTPErrorHandler = api.GetEchoErrorHandler(logger)

	e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins:     []string{"*"}, // TODO: make this dynamically switchable
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowMethods:     []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.PATCH},
		AllowCredentials: true,
	}))

	return e, nil
}

func setupEchoPages(logger *zap.Logger, client *kratos.APIClient, config *conf.Config, metaRepo domain.UserMetaRepo) (*echo.Echo, error) {
	e := echo.New()
	e.HTTPErrorHandler = api.GetEchoErrorHandler(logger)

	// Set up rendering of server side pages
	e.Renderer = tmpl.NewRenderer("./auth/templates/*.html", true)

	protect := middleware.NewBrowserProtectMiddleware(config, metaRepo)

	pagesHandler := auth.NewPublicHandlers(client, config)

	// Serve css stylesheet
	e.File("/k/style.css", "./auth/static/style.css")

	echo.NotFoundHandler = pagesHandler.RootHandler
	kratosGroup := e.Group("/k")
	kratosGroup.GET("/login", pagesHandler.LoginHandler)
	kratosGroup.GET("/verify", pagesHandler.VerificationHandler)
	kratosGroup.GET("/recovery", pagesHandler.RecoveryHandler)
	kratosGroup.GET("/settings", pagesHandler.SettingsHandler, protect)
	kratosGroup.GET("/activated", pagesHandler.SetupCompleteHandler, protect)

	return e, nil
}

func SetupInitialUser(authService domain.AuthService, groupRepo domain.GroupRepo, config *conf.Config) error {
	user, err := authService.CreateUser(context.TODO(), &domain.Traits{
		Email:    config.InitialUserEmail,
		Username: config.InitialUserUsername,
	}, "RefractorSys")
	if err != nil {
		return err
	}

	// Set super admin flag on the user override
	if err := groupRepo.SetUserOverrides(context.TODO(), user.Identity.Id, &domain.Overrides{
		AllowOverrides: perms.GetFlag(perms.FlagSuperAdmin).String(),
		DenyOverrides:  "0",
	}); err != nil {
		return err
	}

	return nil
}

func SetupServerClients(rconService domain.RCONService, serverService domain.ServerService, log *zap.Logger) error {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*10)
	defer cancel()

	allServers, err := serverService.GetAll(ctx)
	if err != nil {
		log.Error("Could not get all servers", zap.Error(err))
		return err
	}

	for _, server := range allServers {
		if err := serverService.CreateServerData(server.ID); err != nil {
			log.Error("Could not create server data", zap.Int64("Server", server.ID), zap.Error(err))
			continue
		}

		if err := rconService.CreateClient(server); err != nil {
			log.Warn("Could not connect RCON client", zap.Int64("Server", server.ID), zap.Error(err))
			continue
		}

		log.Info("RCON client connected", zap.Int64("Server", server.ID))
	}

	return nil
}
