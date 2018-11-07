package fhslib

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"math/rand"
	"time"
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
  port: 8080
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
		Port     int
		Bindip   string
		Loglevel string
	}
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func GenerateId() string {
	id := rand.Uint32()
	id_str := fmt.Sprintf("%08d", id)
	return id_str
}

func GetConfig() (Config, error) {
	var config Config
	err := yaml.Unmarshal([]byte(defaultConfig), &config)
	return config, err
}
