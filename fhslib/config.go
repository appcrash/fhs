package fhslib

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"math/rand"
	"time"
)

const default_config = `
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

var fhslib_config *Config

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
	ensureLogger()
	rand.Seed(time.Now().UTC().UnixNano())
}

func GenerateId() string {
	id := rand.Uint32()
	id_str := fmt.Sprintf("%08d", id)
	return id_str
}

func GetConfig() Config {
	if fhslib_config != nil {
		return *fhslib_config
	}

	fhslib_config = &Config{}
	yaml.Unmarshal([]byte(default_config), fhslib_config)

	if data, err := ioutil.ReadFile("config.yml"); err != nil {
		Log.Info("fhslib config file not found, use default config")
	} else {
		// override default config
		err := yaml.Unmarshal(data, fhslib_config)
		if err != nil {
			Log.Error("fhslib config file format error")
			panic(err)
		}
	}

	Log.Infof("config is %+v", fhslib_config)

	return *fhslib_config
}
