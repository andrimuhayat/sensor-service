package storage

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// SQLConfig represent configuration for SQL driver
type SQLConfig struct {
	DriverName            string
	ServiceName           string
	Host                  string
	Port                  int
	Username              string
	Password              string
	Charset               string
	DBName                string
	MaxOpenConnection     int
	MaxIdleConnection     int
	MaxLifetimeConnection time.Duration
}

func NewMysqlClient(config *SQLConfig) (*sqlx.DB, error) {
	stringCon := []string{config.Username, ":", config.Password, "@tcp(", config.Host, ":", strconv.Itoa(config.Port), ")/", config.DBName, "?charset=", config.Charset, "&parseTime=true&loc=Local"}
	connS := strings.Join(stringCon, "")

	log.Println(connS, config.DriverName)

	db, err := sqlx.Open(config.DriverName, connS)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		return nil, err
	}

	db.SetMaxOpenConns(config.MaxOpenConnection)
	db.SetMaxIdleConns(config.MaxIdleConnection)
	db.SetConnMaxLifetime(config.MaxLifetimeConnection)

	return db, nil
}
