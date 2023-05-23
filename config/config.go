package config

import (
	"GoSegcache/utils"
	"fmt"
	"github.com/creasty/defaults"
	"gopkg.in/dealancer/validate.v2"
	"gopkg.in/yaml.v2"
	"os"
	"reflect"
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
	LFUDecayTime float64 `validate:"gte=0 & lte=65535" default:"1" yaml:"LFUDecayTime"`
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
	//单个segment大小;如 1B,2K,3M,4G,5T;默认值为10M
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

	HandlerConf("GOMemLimit")
	HandlerConf("LFUMemLimit")
	HandlerConf("SegmentSize")
}

// HandlerConf
//
//	@Description: 通过反射 处理字段 并给对应的字段赋值
//	@param s:
func HandlerConf(s string) {
	core := &Conf.Core
	v := reflect.ValueOf(core)
	//根据字符串获取 在结构体中的字段
	field := v.Elem().FieldByName(s)
	//获取字段对应的值
	fieldVale := field.Interface().(string)

	//数据处理部分
	size, unit, err := utils.ExtractStoreUnit(fieldVale)
	if err != nil {
		panic(fmt.Sprintf("GOMemLimit error:%s,should be '3K','3G','3T','3M'...", err))
	}
	lfuMemLimitVal := utils.ToBytes(size, unit)

	//获取 结构体中的新字段
	newField := v.Elem().FieldByName(s + "Val")
	//给新字段 赋值
	newField.SetInt(lfuMemLimitVal)
}
