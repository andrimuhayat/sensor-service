package sensor

import (
	"github.com/labstack/echo/v4"
	"sensor-service/internal/module/sensor/handler"
	"sensor-service/internal/module/sensor/repository"
	"sensor-service/internal/module/sensor/usecase"
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

func StartService(dependency module.Dependency, router *echo.Echo) {
	//init repo
	dependency.SensorRepository = repository.NewRepository(dependency.DB)
	dependency.GenericRepository = repository.NewGenericRepository(dependency.DB)
	//init usecase
	dependency.SensorUseCase = usecase.NewUseCase(dependency.SensorRepository, dependency.MqttClient, dependency.GenericRepository)
	// define handler
	sensorHandler := handler.NewHandler(dependency.SensorUseCase)
	//init route
	versionRoute := router.Group("/test")
	//run consumer mqtt
	RunConsumer(&sync.WaitGroup{}, dependency.SensorUseCase.ListenStreamingData)

	handler.NewSensorRoute(sensorHandler, versionRoute)
}
