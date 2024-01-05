package sensor

import (
	"github.com/labstack/echo/v4"
	"sensor-service/internal/module/sensor/handler"
	"sensor-service/internal/module/sensor/repository"
	"sensor-service/internal/module/sensor/usecase"
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
	dependency.SensorRepository = repository.NewRepository(dependency.DB)
	dependency.GenericRepositorySensor = repository.NewGenericRepository(dependency.DB)
	//init usecase
	dependency.SensorUseCase = usecase.NewUseCase(dependency.SensorRepository, dependency.MqttClient, dependency.GenericRepositorySensor)
	// define handler
	sensorHandler := handler.NewHandler(dependency.SensorUseCase)
	//init route
	versionRoute := router.Group("/api")
	//run consumer mqtt
	RunConsumer(&sync.WaitGroup{}, dependency.SensorUseCase.ListenStreamingData)

	handler.NewSensorRoute(sensorHandler, versionRoute, app)
}
