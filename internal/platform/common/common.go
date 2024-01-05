package module

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/jmoiron/sqlx"
	authRepo "sensor-service/internal/module/auth/repository"
	authUseCase "sensor-service/internal/module/auth/usecase"
	sensorRepo "sensor-service/internal/module/sensor/repository"
	sensorUseCase "sensor-service/internal/module/sensor/usecase"
)

type Dependency struct {
	DB *sqlx.DB
	//repository
	SensorRepository        sensorRepo.IRepository
	GenericRepositorySensor sensorRepo.IGenericRepository
	GenericRepositoryAuth   authRepo.IGenericRepository
	//usecase
	SensorUseCase sensorUseCase.IUseCase
	AuthUseCase   authUseCase.IUseCase
	//message
	MqttClient mqtt.Client
}
