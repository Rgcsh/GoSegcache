package rpc_api

import (
	"GoSegcache/config"
	"GoSegcache/pkg/glog"
	"GoSegcache/proto"
	"GoSegcache/segcache_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"log"
	"net"
	"os"
	"runtime/debug"
	"testing"
)
import "context"

var BufListener *bufconn.Listener

// init
//
//	@Description: 初始化一下启动项
func init() {
	os.Setenv("CONFIG_PATH", "/Users/rgc/GolandProjects/GoSegcache/config/local.yml")
	config.SetUp()
	//设置内存使用限制
	debug.SetMemoryLimit(config.Conf.Core.GOMemLimitVal)
	glog.SetUp()
}

func TestMain(m *testing.M) {
	BufListener = bufconn.Listen(1024 * 1024)
	s := grpc.NewServer(grpc.MaxRecvMsgSize(segcache_service.SegmentBodyLen))

	proto.RegisterGoSegcacheApiServer(s, &Service{})

	go func() {
		if err := s.Serve(BufListener); err != nil {
			log.Fatalf("Server start failed")
		}
	}()

	code := m.Run()
	s.Stop()
	os.Exit(code)
}

func BufDialer(context.Context, string) (net.Conn, error) {
	return BufListener.Dial()
}

func Connect() *grpc.ClientConn {
	conn, _ := grpc.DialContext(context.Background(), "bufnet", grpc.WithContextDialer(BufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	return conn
}
