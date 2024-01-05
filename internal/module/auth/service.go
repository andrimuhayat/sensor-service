package auth

import (
	"github.com/labstack/echo/v4"
	"sensor-service/internal/module/auth/handler"
	"sensor-service/internal/module/auth/repository"
	"sensor-service/internal/module/auth/usecase"
	"sensor-service/internal/platform/app"
	module "sensor-service/internal/platform/common"
	"sync"
)

func RunConsumer(wg *sync.WaitGroup, f func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		f()
	}()
}

func StartService(dependency module.Dependency, router *echo.Echo, app app.App) {
	//init repo
	dependency.GenericRepositoryAuth = repository.NewGenericRepository(dependency.DB)
	//init usecase
	dependency.AuthUseCase = usecase.NewUseCase(dependency.GenericRepositoryAuth, app)
	// define handler
	sensorHandler := handler.NewHandler(dependency.AuthUseCase)
	//init route
	versionRoute := router.Group("/api")
	handler.NewAuthRoute(sensorHandler, versionRoute)
}
