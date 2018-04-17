package fhslib

import (
	"gopkg.in/yaml.v2"
)

var defaultConfig = `
common:
  password: hi
client:
  ip: localhost
  port: 1080
  loglevel: debug
server:
  ip: 127.0.0.1
  bindip: 127.0.0.1
  loglevel: debug
`

type Config struct {
	Common struct {
		Password string
	}
	Client struct {
		Ip       string
		Port     int
		Loglevel string
	}
	Server struct {
		Ip       string
		Bindip   string
		Loglevel string
	}
}

func GetConfig() (Config, error) {
	var config Config
	err := yaml.Unmarshal([]byte(defaultConfig), &config)
	return config, err
}
