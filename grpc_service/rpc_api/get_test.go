package rpc_api

import (
	"GoSegcache/config"
	"GoSegcache/proto"
	"GoSegcache/utils/time_util"
	"bou.ke/monkey"
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// FakeTimeNow
//
//	@Description: 通过 monkey patch伪造时间,模拟时间流逝使用
//	@param now:
func FakeTimeNow(now string) {
	println("造假时间为:", now)
	monkey.Patch(time.Now, func() time.Time {
		t, err := time_util.StringToTime(now, time_util.DateTimeFormat)
		if err != nil {
			panic(err)
		}
		return *t
	})
}

func TestGet(t *testing.T) {
	// 测试 key不存在的情况
	c := proto.NewGoSegcacheApiClient(Connect())
	GetCheck(t, c, "not exist key", "no exist")
}

func TestGet1(t *testing.T) {
	//测试 多次GET一个key时 访问频率的增长情况
	c := proto.NewGoSegcacheApiClient(Connect())
	key := "test key"
	value := []byte("value")
	expireTime := float32(24 * 60 * 60 * 90)
	setReq := &proto.SetReq{Key: key, Value: value, ExpireTime: &expireTime}

	r, err := c.Set(context.Background(), setReq)
	assert.Equal(t, err, nil)
	assert.Equal(t, r.Message, "ok")

	//访问100次,访问次数的值根据lfu算法应该有所增长
	loopCount := 100
	for i := 0; i < loopCount; i++ {
		GetCheck(t, c, key, "ok")
	}

}

func TestGet2(t *testing.T) {
	//测试 多次GET一个key后,间隔一段时间,其 访问次数的衰减情况
	FakeTimeNow("2023-01-01 12:00:00")

	c := proto.NewGoSegcacheApiClient(Connect())
	key := "test key"
	value := []byte("value")
	//90天过期
	expireTime := float32(24 * 60 * 60 * 90)
	setReq := &proto.SetReq{Key: key, Value: value, ExpireTime: &expireTime}

	r, err := c.Set(context.Background(), setReq)
	assert.Equal(t, err, nil)
	assert.Equal(t, r.Message, "ok")

	////访问100次,访问次数的值根据lfu算法应该有所增长
	monkey.Unpatch(time.Now)
	loopCount := 100
	for i := 0; i < loopCount; i++ {
		GetCheck(t, c, key, "ok")
	}

	//	再睡眠1min,检查 衰减结果 是否衰减 10
	FakeTimeNow("2023-01-01 12:10:00")
	_, _ = c.Get(context.Background(), &proto.GetReq{Key: key})

	// 一次性将 衰减程度最大,检查 衰减结果 是否为0
	FakeTimeNow("2023-01-01 12:15:00")
	config.Conf.Core.LFUDecayTime = 0.001
	_, _ = c.Get(context.Background(), &proto.GetReq{Key: key})

	// 超过 分钟循环周期,检查 衰减结果 是否为0
	FakeTimeNow("2023-02-14 12:16:00")
	_, _ = c.Get(context.Background(), &proto.GetReq{Key: key})

}

func TestGet3(t *testing.T) {
	//测试 当key过期后,再次获取数据,在没有清理过期缓存功能时,应该是能获取到 数据
	c := proto.NewGoSegcacheApiClient(Connect())
	key := "test key expire"
	value := []byte("value111")
	expireTime := float32(1)
	setReq := &proto.SetReq{Key: key, Value: value, ExpireTime: &expireTime}

	r, err := c.Set(context.Background(), setReq)
	assert.Equal(t, err, nil)
	assert.Equal(t, r.Message, "ok")

	GetCheck(t, c, key, "ok")

	time.Sleep(time.Second * 2)
	GetCheck(t, c, key, "ok")
}
