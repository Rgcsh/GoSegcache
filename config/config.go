package config

import (
	"GoSegcache/utils"
	"fmt"
	"github.com/creasty/defaults"
	"gopkg.in/dealancer/validate.v2"
	"gopkg.in/yaml.v2"
	"os"
)

// Core
//
//	@Description:
type Core struct {
	//grpc服务连接时的用户名
	UserName string `yaml:"UserName"`
	//grpc服务连接时的密码
	PassWord string `yaml:"PassWord"`
	//grpc端口号
	ServerPort int `default:"5000" yaml:"ServerPort"`
	// 表示 LFU衰减时,衰减程度,值越大,衰减越慢; 值 >=0,为uint16类型,最大值为65535;默认值为1,表示1分钟访问次数减少1;;默认值为1
	LFUDecayTime uint16 `validate:"gte=0 & lte=65535" default:"1" yaml:"LFUDecayTime"`
	// 表示 LFU增加访问量的概率,数值越大,则概率越小,越不容易新增访问次数; 值 >=0,为uint16类型,最大值为65535;;默认值为10
	LFULogFactor uint16 `validate:"gte=0 & lte=65535" default:"10" yaml:"LFULogFactor"`
	// 设置go程序运行使用的最大内存; 如 1B,2K,3M,4G,5T;默认值为1G
	GOMemLimit string `default:"1G" yaml:"GOMemLimit"`
	// 对GOMemLimit单位换算为Byte
	GOMemLimitVal int64
	// 设置LFU 删除数据时的 访问次数 阈值(最大值为255);小于访问次数的key会被删除
	LFUVisitCountLimit uint8 `validate:"gte=0 & lte=255" default:"10" yaml:"LFUVisitCountLimit"`
	// 设置是否启动LFU算法,当内存快达到限制时,主动删除访问频率低的缓存数据 1:启动 0:不启动;默认值为1
	LFUEnable int `validate:"gte=0 & lte=1" default:"1" yaml:"LFUEnable"`
	// 程序消耗主机 物理内存多少时,触发LFU算法; 如 1B,2K,3M,4G,5T;默认值为1G
	LFUMemLimit string `default:"1G" yaml:"LFUMemLimit"`
	// 对LFUMemLimitVal单位换算为Byte
	LFUMemLimitVal int64
	//单个segment大小;如 1B,2K,3M,4G,5T;默认值为1G
	SegmentSize string `default:"10M" yaml:"SegmentSize"`
	//对SegmentSize单位换算为Byte
	SegmentSizeVal int64
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
	//设置默认值
	if err := defaults.Set(Conf); err != nil {
		panic(fmt.Sprintf("默认值设置失败,错误为:%v", err))
	}

	//参数校验
	if err := validate.Validate(Conf); err != nil {
		panic(fmt.Sprintf("参数校验失败,请检测配置文件是否符合限制,错误为:%v", err))
	}

	//修改LFUMemLimit 为byte长度
	size, unit, err := utils.ExtractStoreUnit(Conf.Core.LFUMemLimit)
	if err != nil {
		panic(fmt.Sprintf("LFUMemLimit error:%s,should be '3K','3G','3T','3M'...", err))
	}
	Conf.Core.LFUMemLimitVal = utils.ToBytes(size, unit)

	//修改GOMemLimit 为byte长度
	size, unit, err = utils.ExtractStoreUnit(Conf.Core.GOMemLimit)
	if err != nil {
		panic(fmt.Sprintf("GOMemLimit error:%s,should be '3K','3G','3T','3M'...", err))
	}
	Conf.Core.GOMemLimitVal = utils.ToBytes(size, unit)

	//修改SegmentSize 为byte长度
	size, unit, err = utils.ExtractStoreUnit(Conf.Core.SegmentSize)
	if err != nil {
		panic(fmt.Sprintf("SegmentSize error:%s,should be '3K','3G','3T','3M'...", err))
	}
	Conf.Core.SegmentSizeVal = utils.ToBytes(size, unit)

}
