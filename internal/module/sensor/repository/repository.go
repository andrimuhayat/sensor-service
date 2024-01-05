package repository

import (
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"sensor-service/internal/module/sensor/dto"
	"sensor-service/internal/module/sensor/entity"
	"sensor-service/internal/platform/helper"
	"strconv"
	"strings"
)

type IRepository interface {
	GetAllSensor(dto.SensorQueryParam) ([]*entity.Sensor, error)
}

type Repository struct {
	DB *sqlx.DB
}

func (r Repository) GetAllSensor(params dto.SensorQueryParam) (sensors []*entity.Sensor, err error) {
	selectData := fmt.Sprintf(`%s,%s`, `count(*) over () total_data`, strings.Join(helper.PrintFields(entity.Sensor{}), `,`))
	queryBuilder := sq.Select(selectData).From(helper.GetTableName(entity.Sensor{}))

	if params.CombinationIds != "" {
		ids1, ids2, err := stringToStruct(params.CombinationIds)
		if err != nil {
			return nil, err
		}
		queryBuilder = queryBuilder.Where(sq.Expr(fmt.Sprintf(`ID1 IN (%s) AND ID2 IN(%s)`, ids1, ids2)))
	}

	if params.HourFrom != "" && params.HourTo != "" {
		hfrom := helper.Convert12HourTo24Hour(params.HourFrom)
		hto := helper.Convert12HourTo24Hour(params.HourTo)
		queryBuilder = queryBuilder.Where(
			sq.Expr(fmt.Sprintf(`TIME(timestamp) BETWEEN '%s' AND '%s'`, hfrom, hto)))
	}

	if params.DateFrom != "" && params.DateTo != "" {
		queryBuilder = queryBuilder.Where(
			sq.Expr(fmt.Sprintf(`DATE(timestamp) BETWEEN '%s' AND '%s'`, params.DateFrom, params.DateTo)))
	}

	limit, page := helper.LimitOffset(params.Limit, params.Page)
	queryBuilder = queryBuilder.Limit(uint64(limit)).Offset(uint64(page))

	queryString, args, err := queryBuilder.PlaceholderFormat(sq.Question).ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.DB.Queryx(queryString, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var data entity.Sensor
		if err := rows.StructScan(&data); err != nil {
			//log.Println(err)
			return nil, err
		}
		sensors = append(sensors, &data)
	}
	return sensors, nil
}

func NewRepository(db *sqlx.DB) IRepository {
	return Repository{
		DB: db,
	}
}

func stringToStruct(input string) (string, string, error) {
	var ids1, ids2 string
	// Remove brackets and split by comma to separate key-value pairs
	innerArrays := strings.Split(strings.Trim(input, "[]"), "], [")

	for _, innerArray := range innerArrays {
		// Remove inner brackets and split by comma to separate key-value pairs
		pairs := strings.Split(innerArray, ",")
		for _, pair := range pairs {
			// Split key and value by "="
			parts := strings.Split(pair, "=")
			if len(parts) != 2 {
				return "", "", fmt.Errorf("invalid format")
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Check key and assign value to struct fields accordingly
			if key == "ID1" {
				if ids1 == "" {
					ids1 = fmt.Sprintf(`'%s'`, value)
				} else {
					ids1 += fmt.Sprintf(`,'%s'`, value)
				}
			} else if key == "ID2" {
				_, err := strconv.Atoi(value)
				if err != nil {
					return "", "", fmt.Errorf("ID2 must be an integer")
				}
				if ids2 == "" {
					ids2 = fmt.Sprintf(`'%s'`, value)
				} else {
					ids2 += fmt.Sprintf(`,'%s'`, value)
				}
			} else {
				return "", "", fmt.Errorf("unknown key %s", key)
			}
		}
	}

	return ids1, ids2, nil
}
