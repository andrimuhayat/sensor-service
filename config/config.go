package config

import (
	"errors"
	"github.com/spf13/viper"
	"path"
	"runtime"
)

var (
	platformConfig map[string]*viper.Viper
)

// SetConfig from configuration file path
func SetConfig(path string) error {
	var err error

	viper.SetConfigName("app")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	err = viper.ReadInConfig()
	if err != nil {
		return err
	}

	return nil
}

// LoadConfig into given configuration structure
func LoadConfig(name string, v interface{}) error {
	var (
		err  error
		conf = viper.New()
	)

	conf.SetConfigName(name)
	conf.SetConfigType("yaml")
	conf.AddConfigPath("./configurations")

	err = conf.ReadInConfig()
	if err != nil {
		return err
	}

	return conf.Unmarshal(v)

}

// GetDefaultConfigPath of the service project
func GetDefaultConfigPath() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if ok == false {
		return "", errors.New("runtime caller not valid")
	}

	filePath := path.Join(path.Dir(filename), "../../configurations/")

	return filePath, nil
}

// ReadPlatformConfig will read platform configuration file and put it into internal memory
func ReadPlatformConfig(name string) error {
	var (
		err  error
		conf = viper.New()
	)

	conf.SetConfigName(name)
	conf.SetConfigType("yaml")
	conf.AddConfigPath("./platform")

	err = conf.ReadInConfig()
	if err != nil {
		return err
	}

	if len(platformConfig) == 0 {
		platformConfig = make(map[string]*viper.Viper)
	}

	platformConfig[name] = conf
	return nil
}

// SetPlatformConfig will set platform configuration into internal memory
func SetPlatformConfig(name string, conf *viper.Viper) error {
	if platformConfig == nil {
		platformConfig = make(map[string]*viper.Viper)
	}

	platformConfig[name] = conf
	return nil
}

// SetModuleConfig will get module configuration from platform configuration file and set it into pointer variable
func SetModuleConfig(platform, product string, v interface{}) error {
	platformConf, ok := platformConfig[platform]
	if !ok {
	}

	return platformConf.UnmarshalKey(product, &v)
}

// GetPlatformConfig will get module configuration from platform configuration file and set it into pointer variable
func GetPlatformConfig(platform string, v interface{}) error {
	platformConf, ok := platformConfig[platform]
	if !ok {

	}

	return platformConf.Unmarshal(&v)
}

// GetAllPlatformConfig returns all registered platform configuration from internal memory
func GetAllPlatformConfig() map[string]*viper.Viper {
	if platformConfig == nil {
		platformConfig = make(map[string]*viper.Viper)
	}

	return platformConfig
}
