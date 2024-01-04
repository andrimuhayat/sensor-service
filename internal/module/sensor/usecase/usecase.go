package usecase

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/mitchellh/mapstructure"
	"log"
	"math"
	"net/http"
	"sensor-service/config"
	"sensor-service/internal/module/sensor/dto"
	"sensor-service/internal/module/sensor/entity"
	sensor "sensor-service/internal/module/sensor/repository"
	"sensor-service/internal/platform/helper"
	"sensor-service/internal/platform/httpengine/httpresponse"
	"strconv"
)

type IUseCase interface {
	ListenStreamingData()
	GetAllSensor(config.HTTPRequest) (*httpresponse.Pagination, *httpresponse.HTTPError)
	UpdateSensor(config.HTTPRequest) *httpresponse.HTTPError
	DeleteSensor(config.HTTPRequest) *httpresponse.HTTPError
}

type UseCase struct {
	SensorRepository  sensor.IRepository
	GenericRepository sensor.IGenericRepository
	MqttClient        mqtt.Client
}

func (u UseCase) GetAllSensor(request config.HTTPRequest) (*httpresponse.Pagination, *httpresponse.HTTPError) {
	var err error
	httpError := httpresponse.HTTPError{}
	var queryParamReq dto.SensorQueryParam
	var totalData int

	config := helper.DecoderConfig(&queryParamReq)
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		httpError.Code = http.StatusInternalServerError
		httpError.Message = httpresponse.ErrorInternalServerError.Message
		return nil, &httpError
	}
	if err = decoder.Decode(request.Queries); err != nil {
		log.Println("{SensorDataGenerateRequest}{Decode}{Error} : ", err)
	}

	limit, _ := strconv.Atoi(queryParamReq.Limit)
	page, _ := strconv.Atoi(queryParamReq.Page)

	if limit == 0 {
		queryParamReq.Limit = "10"
	}

	if page == 0 {
		queryParamReq.Page = "1"
	}

	sensors, err := u.SensorRepository.GetAllSensor(queryParamReq)
	if err != nil {
		log.Println(err)
		httpError.Code = http.StatusInternalServerError
		httpError.Message = httpresponse.ErrorInternalServerError.Message
		return nil, &httpError
	}

	if len(sensors) > 0 && sensors != nil {
		totalData = helper.GetIntPtrValue(sensors[0].TotalData)
	}

	limit, _ = strconv.Atoi(queryParamReq.Limit)
	page, _ = strconv.Atoi(queryParamReq.Page)

	return &httpresponse.Pagination{
		Data:         sensors,
		TotalData:    totalData,
		TotalPage:    int(math.Ceil(float64(totalData) / float64(limit))),
		CurrentPage:  page,
		TotalPerPage: len(sensors),
	}, nil
}

func (u UseCase) UpdateSensor(request config.HTTPRequest) *httpresponse.HTTPError {
	var err error
	httpError := httpresponse.HTTPError{}
	var queryParamReq dto.SensorQueryParam

	config := helper.DecoderConfig(&queryParamReq)
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		httpError.Code = http.StatusInternalServerError
		httpError.Message = httpresponse.ErrorInternalServerError.Message
		return &httpError
	}
	if err = decoder.Decode(request.Queries); err != nil {
		log.Println("{SensorDataGenerateRequest}{Decode}{Error} : ", err)
		//helper.ResponseWithError(w, http.StatusInternalServerError, httpresponse.ErrorInternalServerError.Message)
	}

	var sensorDataGenerateRequest dto.SensorDataGenerateRequest

	config = helper.DecoderConfig(&sensorDataGenerateRequest)
	decoder, err = mapstructure.NewDecoder(config)
	if err != nil {
		httpError.Code = http.StatusInternalServerError
		httpError.Message = httpresponse.ErrorInternalServerError.Message
		return &httpError
	}
	if err = decoder.Decode(request.Body); err != nil {
		log.Println("{SensorDataGenerateRequest}{Decode}{Error} : ", err)
	}

	allSensor, err := u.SensorRepository.GetAllSensor(queryParamReq)
	if err != nil {
		httpError.Code = http.StatusInternalServerError
		httpError.Message = httpresponse.ErrorInternalServerError.Message
		return &httpError
	}

	for _, v := range allSensor {
		err = u.GenericRepository.Update(*v)
		if err != nil {
			httpError.Code = http.StatusInternalServerError
			httpError.Message = httpresponse.ErrorInternalServerError.Message
			return &httpError
		}
	}

	return nil
}

func (u UseCase) DeleteSensor(request config.HTTPRequest) *httpresponse.HTTPError {
	var err error
	httpError := httpresponse.HTTPError{}
	var queryParamReq dto.SensorQueryParam

	config := helper.DecoderConfig(&queryParamReq)
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		httpError.Code = http.StatusInternalServerError
		httpError.Message = httpresponse.ErrorInternalServerError.Message
		return &httpError
	}
	if err = decoder.Decode(request.Queries); err != nil {
		log.Println("{SensorDataGenerateRequest}{Decode}{Error} : ", err)
	}

	allSensor, err := u.SensorRepository.GetAllSensor(queryParamReq)
	if err != nil {
		httpError.Code = http.StatusInternalServerError
		httpError.Message = httpresponse.ErrorInternalServerError.Message
		return &httpError
	}

	for _, v := range allSensor {
		err = u.GenericRepository.DeleteByID(*v, v.ID)
		if err != nil {
			httpError.Code = http.StatusInternalServerError
			httpError.Message = httpresponse.ErrorInternalServerError.Message
			return &httpError
		}
	}
	return nil
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
