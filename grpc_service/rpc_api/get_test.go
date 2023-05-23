package rpc_api

import (
	"GoSegcache/config"
	"GoSegcache/proto"
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGet1(t *testing.T) {
	c := proto.NewGoSegcacheApiClient(Connect())
	rGet, errGet := c.Get(context.Background(), &proto.GetReq{Key: "unexist key"})
	assert.Equal(t, errGet, nil)
	assert.Equal(t, rGet.Message, "no exist")
}

func TestGet(t *testing.T) {
	//测试 多次GET一个key时 访问频率的变化

	c := proto.NewGoSegcacheApiClient(Connect())
	key := "TEST KEY"
	value := []byte("value")
	expireTime := float32(24 * 60 * 60)
	setReq := &proto.SetReq{Key: key, Value: value, ExpireTime: &expireTime}

	r, err := c.Set(context.Background(), setReq)
	assert.Equal(t, err, nil)
	assert.Equal(t, r.Message, "ok")

	//先将 访问次数 根据算法 增长到 最大值 255
	for i := 0; i < 40000; i++ {
		rGet, errGet := c.Get(context.Background(), &proto.GetReq{Key: key})
		assert.Equal(t, errGet, nil)
		assert.Equal(t, rGet.Message, "ok")
	}

	//	再睡眠1min,检查 衰减结果 是否衰减 10
	config.Conf.Core.LFUDecayTime = 0.1
	time.Sleep(time.Minute)
	_, _ = c.Get(context.Background(), &proto.GetReq{Key: key})

	// 一次性将 衰减程度最大,检查 衰减结果 是否为0
	config.Conf.Core.LFUDecayTime = 0.001
	time.Sleep(time.Minute)
	_, _ = c.Get(context.Background(), &proto.GetReq{Key: key})

}
