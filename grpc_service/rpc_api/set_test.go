package rpc_api

import (
	"GoSegcache/config"
	"GoSegcache/proto"
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
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
func TestSet1(t *testing.T) {
	var key string
	var valStr string

	c := proto.NewGoSegcacheApiClient(Connect())
	// 1s<=过期时间<1h 储存的过期时间在 秒(s)级的TTLMap中
	expireTime := float32(2)
	key = "S级过期"
	valStr = "1s<=过期时间<1h 储存的过期时间在 秒(s)级的TTLMap中"
	CheckSetGet(t, c, key, valStr, expireTime)

	// 1h<=过期时间<1d 储存的过期时间在 分钟(m)级的TTLMap中
	expireTime = float32(60*60 + 1)
	key = "M级过期"
	valStr = "1h<=过期时间<1d 储存的过期时间在 分钟(m)级的TTLMap中"
	CheckSetGet(t, c, key, valStr, expireTime)

	// 1d<=过期时间 储存的过期时间在 小时(h)级的TTLMap中
	expireTime = float32(24*60*60 + 1)
	key = "H级过期"
	valStr = "1d<=过期时间 储存的过期时间在 小时(h)级的TTLMap中"
	CheckSetGet(t, c, key, valStr, expireTime)

	// 过期时间 不填 储存的过期时间在 不过期的TTLMap中
	expireTime = 0
	key = "永不过期"
	valStr = "过期时间 不填 储存的过期时间在 不过期的TTLMap中"
	CheckSetGet(t, c, key, valStr, expireTime)
}

func TestSet2(t *testing.T) {
	//	同一个过期时间范围内的数据 存入到TTLMap中的key应该相同,且持续存入多个数据,需要开辟新的segment存新数据; 再检查 根据key 获取的值,过期时间范围 是否正确
	var key string
	var valStr string
	expireTime := float32(24*60*60 + 1)

	c := proto.NewGoSegcacheApiClient(Connect())
	config.Conf.Core.SegmentSizeVal = 10
	for i := 0; i < 100; i++ {
		key = fmt.Sprintf("key:%v", i)
		valStr = fmt.Sprintf("value is %v", key)
		CheckSetGet(t, c, key, valStr, expireTime)
	}
}

func TestSet(t *testing.T) {
	//测试用例部分
	//用例场景 普通的 set使用
	c := proto.NewGoSegcacheApiClient(Connect())
	expireTime := float32(2)
	//expire_time := float32(10)
	key := "你好"
	valStr := "abc"
	CheckSetGet(t, c, key, valStr, expireTime)
}
