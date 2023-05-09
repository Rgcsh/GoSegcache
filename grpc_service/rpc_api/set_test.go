package rpc_api

import (
	"GoSegcache/proto"
	"context"
	"fmt"
	"testing"
)

func TestSet(t *testing.T) {
	c := proto.NewGoSegcacheApiClient(Connect())
	expire_time := float32(10000000)
	//expire_time := float32(10)
	value := []byte("abc")
	r, err := c.Set(context.Background(), &proto.SetReq{Key: "你好", Value: value, ExpireTime: &expire_time})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(r.Message)
	}
	//r, err = c.Set(context.Background(), &proto.SetReq{Key: "你好", Value: value, ExpireTime: &expire_time})
	//if err != nil {
	//	fmt.Println(err)
	//} else {
	//	fmt.Println(r.Message)
	//}
	r1, err := c.Get(context.Background(), &proto.GetReq{Key: "你好"})
	r1, err = c.Get(context.Background(), &proto.GetReq{Key: "你好"})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("success", r1)
	}
}
