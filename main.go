package main

import (
	"GoSegcache/config"
	"GoSegcache/grpc_service/rpc_api"
	"GoSegcache/pkg/glog"
	"GoSegcache/proto"
	"GoSegcache/utils"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"math"
	"net"
	"runtime"
	"runtime/debug"
	"strconv"
)

// setMemoryLimit
//
//	@Description: 设置内存使用限制
func setMemoryLimit() {
	size, unit, err := utils.ExtractStoreUnit(config.Conf.Core.GOMEMLIMIT)
	if err != nil {
		e := fmt.Sprintf("GOMEMLIMIT error:%s,should be '3K','3G','3T','3M'...", err)
		glog.Log.Error(e)
		panic(e)
	}
	memoryLimit := utils.ToBytes(size, unit)
	debug.SetMemoryLimit(memoryLimit)
	return
}

// init
//
//	@Description: 初始化一下启动项
func init() {
	config.SetUp()
	setMemoryLimit()
	glog.SetUp()
}

func main() {
	runtime.GC()
	//读取服务端口号配置,转为字符串类型
	serverPort := strconv.Itoa(config.Conf.Core.ServerPort)
	host := fmt.Sprintf(":%s", serverPort)
	//监听TCP服务
	tcp, err := net.Listen("tcp", host)
	if err != nil {
		glog.Log.Error(fmt.Sprintf("listen tcp error:%s", err))
		return
	}
	//最大接受消息大小为10M
	s := grpc.NewServer(grpc.MaxRecvMsgSize(int(math.Pow(2, 20)) * 10))
	proto.RegisterGoSegcacheApiServer(s, &rpc_api.Service{})

	//启动grpc服务
	glog.Log.Info("Grpc server start at", zap.String("HOST", host))
	if err := s.Serve(tcp); err != nil {
		glog.Log.Error("Start grpc server failed", zap.Error(err))
	}
}
