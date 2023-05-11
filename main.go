package main

import (
	"GoSegcache/config"
	"GoSegcache/grpc_service/crontab"
	"GoSegcache/grpc_service/rpc_api"
	"GoSegcache/pkg/glog"
	"GoSegcache/proto"
	"GoSegcache/segcache_service"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
	"runtime/debug"
	"strconv"
)

// init
//
//	@Description: 初始化一下启动项
func init() {
	config.SetUp()
	//设置内存使用限制
	debug.SetMemoryLimit(config.Conf.Core.GOMemLimitVal)
	glog.SetUp()
}

func main() {
	//启动主动删除缓存功能 相关子协程
	go crontab.CleanExpiredData()
	if config.Conf.Core.LFUEnable == 1 {
		go crontab.CleanExpiringData()
	}

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
	s := grpc.NewServer(grpc.MaxRecvMsgSize(segcache_service.SegmentBodyLen))
	proto.RegisterGoSegcacheApiServer(s, &rpc_api.Service{})

	//启动grpc服务
	glog.Log.Info("Grpc server start at", zap.String("HOST", host))
	if err := s.Serve(tcp); err != nil {
		glog.Log.Error("Start grpc server failed", zap.Error(err))
	}
}
