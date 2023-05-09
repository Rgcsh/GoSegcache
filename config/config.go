package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

type Core struct {
	UserName     string `yaml:"UserName"`
	PassWord     string `yaml:"PassWord"`
	ServerPort   int    `yaml:"ServerPort"`
	LFUDecayTime uint16 `yaml:"LFUDecayTime"`
	LFULogFactor uint16 `yaml:"LFULogFactor"`
	GOMEMLIMIT   string `yaml:"GOMEMLIMIT"`
}

type Config struct {
	Core Core `yaml:"Core"`
}

var Conf = &Config{}

// SetUp
//
//	@Description:设置配置文件
func SetUp() {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		panic("Invalid config path to load,please use 'CONFIG_PATH' set to os environment!")
	}
	f, err := os.Open(path)
	if err != nil {
		panic(fmt.Sprintf("Load config path %s failed:%s,please reconfig", path, err))
	}
	err = yaml.NewDecoder(f).Decode(Conf)
	if err != nil {
		panic(fmt.Sprintf("yaml file decode failed:%s,please check again", err))
	}
}
