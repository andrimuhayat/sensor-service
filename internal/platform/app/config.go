package app

import "time"

// Config represent the entire service configurations.
type Config struct {
	App struct {
		Port string `mapstructure:"port"`
		Name string `mapstructure:"name"`
	} `mapstructure:"app"`
	SQL struct {
		DriverName            string        `mapstructure:"driver_name"`
		Host                  string        `mapstructure:"host"`
		Port                  int           `mapstructure:"port"`
		Username              string        `mapstructure:"username"`
		Password              string        `mapstructure:"password"`
		Charset               string        `mapstructure:"charset"`
		DbName                string        `mapstructure:"db_name"`
		MaxOpenConnection     int           `mapstructure:"max_open_connection"`
		MaxIdleConnection     int           `mapstructure:"max_idle_connection"`
		MaxLifetimeConnection time.Duration `mapstructure:"max_lifetime_connection"`
		DbMigrate             bool          `mapstructure:"db_migrate"`
	} `mapstructure:"mysql"`
	Kafka struct {
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		Host     string `mapstructure:"url"`
		Port     string `mapstructure:"port"`
	} `mapstructure:"kafka"`
	Mqtt struct {
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		Host     string `mapstructure:"Host"`
		Port     string `mapstructure:"port"`
	} `mapstructure:"mqtt"`
}
