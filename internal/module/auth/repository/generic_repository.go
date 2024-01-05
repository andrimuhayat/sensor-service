package repository

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"log"
	"reflect"
	"sensor-service/internal/platform/helper"
)

type IGenericRepository interface {
	Create(T any) error
	Update(T any) error
	FindByID(T any, id int) (d *any, err error)
	FindBy(T any, R any) (d *any, err error)
	FindAll(T any) ([]*any, error)
	FindAllBy(T any, R any) ([]*any, error)
	DeleteByID(T any, id int) error
	FindByEmail(T any, email string) (d *any, err error)
}

type GenericRepository struct {
	DB *sqlx.DB
}

func (r GenericRepository) FindByEmail(T any, email string) (d *any, err error) {
	userColumn := helper.PrintFields(T)
	queryBuilder := sq.Select(userColumn...).
		Where(sq.Eq{"`email`": email}).
		From(helper.GetTableName(T))

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
		t := reflect.TypeOf(T)
		val := reflect.New(t).Interface()
		if err = rows.Scan(helper.StrutForScan(val)...); err != nil {
			return nil, err
		}
		d = &val
	}
	return d, nil
}

func (r GenericRepository) FindAllBy(T any, R any) ([]*any, error) {
	//TODO implement me
	panic("implement me")
}

func (r GenericRepository) FindBy(T any, R any) (d *any, err error) {
	//TODO implement me
	panic("implement me")
}

func (r GenericRepository) Create(T any) error {
	values := helper.GetInsertValues(T)
	columns := helper.PrintInsertFields(T)

	queryBuilder := sq.Insert(helper.GetTableName(T)).
		Columns(columns...).
		Values(values...)

	queryString, args, err := queryBuilder.PlaceholderFormat(sq.Question).ToSql()
	if err != nil {
		return err
	}
	_, err = r.DB.Exec(queryString, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r GenericRepository) Update(T any) error {
	columnValues := helper.StructMap(T)
	queryBuilder := sq.Update(helper.GetTableName(T)).SetMap(columnValues).Where(sq.Eq{`id`: columnValues["id"].(int)})
	queryString, args, err := queryBuilder.PlaceholderFormat(sq.Question).ToSql()
	if err != nil {
		return err
	}
	_, err = r.DB.Exec(queryString, args...)
	if err != nil {
		log.Println(err)
		return err
	}

	if err != nil {
		return err
	}

	return nil
}

func (r GenericRepository) FindByID(T any, id int) (d *any, err error) {
	userColumn := helper.PrintFields(T)
	queryBuilder := sq.Select(userColumn...).
		Where(sq.Eq{"id": id}).
		From(helper.GetTableName(T))

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
		t := reflect.TypeOf(T)
		val := reflect.New(t).Interface()
		if err = rows.Scan(helper.StrutForScan(val)...); err != nil {
			return nil, err
		}
		d = &val
	}
	return d, nil
}

func (r GenericRepository) FindAll(T any) (datas []*any, err error) {
	columns := helper.PrintFields(T)
	queryBuilder := sq.Select(columns...).
		From(helper.GetTableName(T))

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
		t := reflect.TypeOf(T)
		val := reflect.New(t).Interface()
		if err = rows.Scan(helper.StrutForScan(val)...); err != nil {
			return nil, err
		}
		datas = append(datas, &val)
	}

	return datas, nil
}

func (r GenericRepository) DeleteByID(T any, id int) error {
	queryBuilder := sq.Delete(helper.GetTableName(T)).Where(sq.Eq{`id`: id})
	queryString, args, err := queryBuilder.PlaceholderFormat(sq.Question).ToSql()
	if err != nil {
		return err
	}
	_, err = r.DB.Exec(queryString, args...)
	if err != nil {
		log.Println(err)
		return err
	}

	if err != nil {
		return err
	}

	return nil
}

func NewGenericRepository(conn *sqlx.DB) IGenericRepository {
	return GenericRepository{
		DB: conn,
	}
}
