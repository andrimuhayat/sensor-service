package dto

import "time"

type SensorQueryParam struct {
	CombinationIds string `json:"combination_ids"`
	HourFrom       string `json:"hour_from"`
	HourTo         string `json:"hour_to"`
	DateFrom       string `json:"date_from"`
	DateTo         string `json:"date_to"`
	Page           string `json:"page"`
	Limit          string `json:"limit"`
}

type SensorDataGenerateRequest struct {
	ID          int       `json:"id"`
	SensorValue float32   `json:"sensor_value"`
	SensorType  string    `json:"sensor_type"`
	ID1         string    `json:"ID1"`
	ID2         int       `json:"ID2"`
	Timestamp   time.Time `json:"timestamp"`
}
