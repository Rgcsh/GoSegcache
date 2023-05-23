#!/bin/bash
rm -rf c.out
# -v 命令:实时显示日志输出
go test ./grpc_service/rpc_api -coverprofile=c.out -covermode=count -v
#红色表示没有调用，灰色表示频率较低，绿色随颜色深浅表示不同程度的频率调用
go tool cover -html=c.out
