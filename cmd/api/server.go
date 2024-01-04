package api

import (
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
	"sensor-service/internal/module/sensor"
	"sensor-service/internal/platform/app"
	module "sensor-service/internal/platform/common"
	"sensor-service/internal/platform/httpengine"
	"sensor-service/internal/platform/httpengine/echoserver"
	"sensor-service/internal/platform/messagebroker/mqqt"
	internalMdw "sensor-service/internal/platform/middleware"
	"sensor-service/internal/platform/storage"
	"sensor-service/internal/platform/storage/migration"
)

// Server httpengine service
type Server struct {
	DB        *sqlx.DB
	Router    httpengine.Router
	AppRouter *echo.Echo
}

// NewServer httpengine initialization
func NewServer() (*Server, error) {
	var (
		err        error
		appConfig  app.Config
		server     = new(Server)
		echoRouter = echoserver.NewEchoRouter()
	)

	err = viper.Unmarshal(&appConfig)
	if err != nil {
		return nil, err
	}

	err = server.initInternalDependency(&appConfig)
	if err != nil {
		return nil, err
	}

	server.Router = echoRouter
	server.Run(appConfig)

	return server, nil
}

func (s *Server) Run(config app.Config) {
	s.AppRouter = s.Router.GetRouter()
	s.AppRouter.Use(internalMdw.PanicException)
	s.AppRouter.Use(middleware.RequestID())

	s.initModuleDependency(&config)

	s.Router.SERVE(config.App.Port)
}

func (s *Server) initInternalDependency(appConfig *app.Config) error {
	var err error

	sqlConfig := storage.SQLConfig{
		DriverName:            appConfig.SQL.DriverName,
		ServiceName:           appConfig.App.Name,
		Host:                  appConfig.SQL.Host,
		Port:                  appConfig.SQL.Port,
		Username:              appConfig.SQL.Username,
		Password:              appConfig.SQL.Password,
		Charset:               appConfig.SQL.Charset,
		DBName:                appConfig.SQL.DbName,
		MaxOpenConnection:     appConfig.SQL.MaxOpenConnection,
		MaxIdleConnection:     appConfig.SQL.MaxIdleConnection,
		MaxLifetimeConnection: appConfig.SQL.MaxLifetimeConnection,
	}
	DB, err := storage.NewMysqlClient(&sqlConfig)
	if err != nil {
		return err
	}
	s.DB = DB

	if appConfig.SQL.DbMigrate {
		migration.MigrationRubenv(DB)
	}

	return nil
}

func (s *Server) initModuleDependency(appConfig *app.Config) module.Dependency {
	var (
		dependency module.Dependency
	)
	dependency.MqttClient = mqqt.Connect("sub", appConfig.Mqtt)
	dependency.DB = s.DB
	sensor.StartService(dependency, s.AppRouter)
	return dependency
}
