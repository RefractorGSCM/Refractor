package main

import (
	"Refractor/auth"
	"Refractor/domain"
	"Refractor/games/mordhau"
	_authRepo "Refractor/internal/auth/repos/kratos"
	_gameService "Refractor/internal/game/service"
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
	sqladapter "github.com/Blank-Xu/sql-adapter"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v4"
	"github.com/labstack/echo/v4"
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

	db, driverName, err := setupDatabase(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatalf("Could not set up database. Error: %v", err)
	}

	enforcer, err := setupCasbin(db, driverName)
	if err != nil {
		log.Fatalf("Could not set up casbin. Error: %v", err)
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

	if len(users) < 1 {
		if err := SetupInitialUser(authRepo, config); err != nil {
			log.Fatalf("Could not create initial user. Error: %v", err)
		}

		log.Printf("Initial superadmin user (%s) has been created!", config.InitialUserUsername)
	}

	protectMiddleware := middleware.NewAPIProtectMiddleware(config)

	gameService := _gameService.NewGameService()
	gameService.AddGame(mordhau.NewMordhauGame(playfab.NewPlayfabPlatform()))

	serverRepo := _postgresServerRepo.NewServerRepo(db, logger)
	serverService := _serverService.NewServerService(serverRepo, time.Second*2)
	_serverHandler.ApplyServerHandler(apiGroup, serverService, protectMiddleware)

	// Setup complete. Begin serving requests.
	logger.Info("Setup complete!")
	enforcer.Enforce()

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

//go:embed rbac_model.conf
var rbacModel string

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

func setupCasbin(db *sql.DB, dbType string) (*casbin.Enforcer, error) {
	// Init casbin adapter
	adapter, err := sqladapter.NewAdapter(db, dbType, "Casbin")
	if err != nil {
		return nil, errors.Wrap(err, "Could not create casbin sql adapter")
	}

	casbinModel, err := model.NewModelFromString(rbacModel)
	if err != nil {
		return nil, errors.Wrap(err, "Could not get casbin model")
	}

	enforcer, err := casbin.NewEnforcer(casbinModel, adapter)
	if err != nil {
		return nil, errors.Wrap(err, "Could not create casbin enforcer")
	}

	return enforcer, nil
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

	return e, nil
}

func setupEchoPages(logger *zap.Logger, client *kratos.APIClient, config *conf.Config) (*echo.Echo, error) {
	e := echo.New()
	e.HTTPErrorHandler = api.GetEchoErrorHandler(logger)

	// Set up rendering of server side pages
	e.Renderer = tmpl.NewRenderer("./auth/templates/*.html", true)

	protect := middleware.NewBrowserProtectMiddleware(config)

	pagesHandler := auth.NewPublicHandlers(client, config)

	echo.NotFoundHandler = pagesHandler.RootHandler
	kratosGroup := e.Group("/k")
	kratosGroup.GET("/login", pagesHandler.LoginHandler)
	kratosGroup.GET("/recovery", pagesHandler.RecoveryHandler)
	kratosGroup.GET("/settings", pagesHandler.SettingsHandler, protect)

	return e, nil
}

func SetupInitialUser(authRepo domain.AuthRepo, config *conf.Config) error {
	if _, err := authRepo.CreateUser(&domain.Traits{
		Email:    config.InitialUserEmail,
		Username: config.InitialUserUsername,
	}); err != nil {
		return err
	}

	return nil
}
