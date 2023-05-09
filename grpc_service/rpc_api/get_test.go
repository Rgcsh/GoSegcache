package rpc_api

import (
	"GoSegcache/proto"
	"context"
	"fmt"
	"testing"
)

func TestGet(t *testing.T) {
	c := proto.NewGoSegcacheApiClient(Connect())
	r, err := c.Get(context.Background(), &proto.GetReq{Key: "你好"})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(r.Message)
	}
}
