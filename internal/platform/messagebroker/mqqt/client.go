package mqqt

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"time"
)

func Connect(clientId string, uri struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Host     string `mapstructure:"Host"`
	Port     string `mapstructure:"port"`
}) mqtt.Client {
	opts := CreateClientOptions(clientId, uri)
	client := mqtt.NewClient(opts)
	token := client.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}
	log.Println(uri.Host)
	if err := token.Error(); err != nil {
		log.Println("errr ? ", err)
	}
	return client
}

func CreateClientOptions(clientId string, uri struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Host     string `mapstructure:"Host"`
	Port     string `mapstructure:"port"`
}) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s", fmt.Sprintf(`%s:%s`, uri.Host, uri.Port)))
	//opts.SetUsername(uri.User.Username())
	//password, _ := uri.User.Password()
	//opts.SetPassword(password)
	opts.SetClientID(clientId)
	return opts
}
