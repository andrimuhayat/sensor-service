package usecase

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"sensor-service/internal/module/sensor/entity"
	sensor "sensor-service/internal/module/sensor/repository"
	"sensor-service/internal/platform/helper"
)

type IUseCase interface {
	ListenStreamingData()
}

type UseCase struct {
	SensorRepository  sensor.IRepository
	GenericRepository sensor.IGenericRepository
	MqttClient        mqtt.Client
}

func (u UseCase) ListenStreamingData() {
	u.MqttClient.Subscribe(helper.STREAMSENSOR, 0, func(client mqtt.Client, msg mqtt.Message) {
		var dataSensor entity.Sensor
		err := json.Unmarshal(msg.Payload(), &dataSensor)
		if err != nil {
			log.Println("{ListenStreamingData}{Unmarshal}{Error} : ", err)
		}
		fmt.Printf("* [%s] %s\n", msg.Topic(), string(msg.Payload()))
		err = u.GenericRepository.Create(dataSensor)
		if err != nil {
			log.Println("{ListenStreamingData}{Create}{Error} : ", err)
		}
	})
}

func NewUseCase(repository sensor.IRepository, client mqtt.Client, genericRepository sensor.IGenericRepository) IUseCase {
	return UseCase{
		SensorRepository:  repository,
		MqttClient:        client,
		GenericRepository: genericRepository,
	}
}
