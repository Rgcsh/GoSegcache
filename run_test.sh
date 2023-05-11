#!/bin/bash
rm -rf c.out
go test ./grpc_service/rpc_api -coverprofile=c.out -covermode=count
#红色表示没有调用，灰色表示频率较低，绿色随颜色深浅表示不同程度的频率调用
go tool cover -html=c.out
