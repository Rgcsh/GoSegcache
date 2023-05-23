package crontab

//测试场景
//新建多个M级别的数据,在超过 过期时间后,尝试获取缓存值,应该显示 不存在
import (
	"GoSegcache/config"
	"GoSegcache/proto"
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func CheckSetGet(t *testing.T, c proto.GoSegcacheApiClient, key, valStr string, expireTime float32) {
	value := []byte(valStr)
	setReq := &proto.SetReq{Key: key, Value: value, ExpireTime: &expireTime}
	if expireTime == 0 {
		setReq.ExpireTime = nil
	}
	r, err := c.Set(context.Background(), setReq)
	assert.Equal(t, err, nil)
	assert.Equal(t, r.Message, "ok")

	rGet, err := c.Get(context.Background(), &proto.GetReq{Key: key})
	assert.Equal(t, err, nil)
	assert.Equal(t, rGet.Message, "ok")
	assert.Equal(t, rGet.Value, value)

}

// TestSet1
//
//	@Description: 不同过期时间范围(秒,分钟,小时,永不过期 4种),检查 存入的TTLMap级别是否正确
//	@param t:
func TestCleanExpiredData(t *testing.T) {
	var key string
	var valStr string
	expireTime := float32(2 * 60 * 60)

	FakeTimeNow("2023-01-01 12:10:00")
	c := proto.NewGoSegcacheApiClient(Connect())
	config.Conf.Core.SegmentSizeVal = 60
	for i := 0; i < 6; i++ {
		key = fmt.Sprintf("key:%v", i)
		valStr = fmt.Sprintf("value is %v", key)
		CheckSetGet(t, c, key, valStr, expireTime)
	}

	FakeTimeNow("2023-01-01 12:12:00")
	time.Sleep(time.Second * 3)

	for i := 0; i < 100; i++ {
		key = fmt.Sprintf("key:%v", i)
		rGet, err := c.Get(context.Background(), &proto.GetReq{Key: key})
		assert.Equal(t, err, nil)
		assert.Equal(t, rGet.Message, "no exist")
	}
}
