package entity

import (
	"sensor-service/internal/platform/helper"
)

type Sensor struct {
	ID          int                    `json:"id" db:"id"`
	SensorValue float32                `json:"sensor_value" db:"sensor_value"`
	SensorType  string                 `json:"sensor_type" db:"sensor_type"`
	ID1         string                 `json:"ID1" db:"ID1"`
	ID2         int                    `json:"ID2" db:"ID2"`
	Timestamp   *helper.DateTimeString `json:"timestamp" db:"timestamp"`
}
