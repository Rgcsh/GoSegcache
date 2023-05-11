package config

import (
	"GoSegcache/pkg/glog"
	"GoSegcache/utils"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

type Core struct {
	UserName           string `yaml:"UserName"`
	PassWord           string `yaml:"PassWord"`
	ServerPort         int    `yaml:"ServerPort"`
	LFUDecayTime       uint16 `yaml:"LFUDecayTime"`
	LFULogFactor       uint16 `yaml:"LFULogFactor"`
	GOMemLimit         string `yaml:"GOMemLimit"`
	GOMemLimitVal      int64
	LFUVisitCountLimit uint8  `yaml:"LFUVisitCountLimit"`
	LFUEnable          int    `yaml:"LFUEnable"`
	LFUMemLimit        string `yaml:"LFUMemLimit"`
	LFUMemLimitVal     int64
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

	//对 配置值进行检测及转换
	//修改LFUMemLimit 为byte长度
	size, unit, err := utils.ExtractStoreUnit(Conf.Core.LFUMemLimit)
	if err != nil {
		e := fmt.Sprintf("LFUMemLimit error:%s,should be '3K','3G','3T','3M'...", err)
		glog.Log.Error(e)
		panic(e)
	}
	Conf.Core.LFUMemLimitVal = utils.ToBytes(size, unit)

	//修改GOMemLimit 为byte长度
	size, unit, err = utils.ExtractStoreUnit(Conf.Core.GOMemLimit)
	if err != nil {
		e := fmt.Sprintf("GOMemLimit error:%s,should be '3K','3G','3T','3M'...", err)
		glog.Log.Error(e)
		panic(e)
	}
	Conf.Core.GOMemLimitVal = utils.ToBytes(size, unit)

}
