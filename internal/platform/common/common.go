package module

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/jmoiron/sqlx"
	"sensor-service/internal/module/sensor/repository"
	"sensor-service/internal/module/sensor/usecase"
)

type Dependency struct {
	DB *sqlx.DB
	//repository
	SensorRepository  repository.IRepository
	GenericRepository repository.IGenericRepository
	//usecase
	SensorUseCase usecase.IUseCase
	//message
	MqttClient mqtt.Client
}
