package rpc_api

import (
	"GoSegcache/config"
	"GoSegcache/grpc_service/crontab"
	"GoSegcache/pkg/glog"
	"GoSegcache/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"log"
	"net"
	"os"
	"testing"
)
import "context"

var BufListener *bufconn.Listener

func SetUp() {
	os.Setenv("CONFIG_PATH", "/Users/rgc/GolandProjects/GoSegcache/config/local.yml")
	config.SetUp()
	glog.SetUp()
}

func TestMain(m *testing.M) {
	SetUp()
	go crontab.CleanExpiredData()
	BufListener = bufconn.Listen(1024 * 1024)
	s := grpc.NewServer(grpc.MaxRecvMsgSize(int(config.Conf.Core.SegmentSizeVal)))

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
