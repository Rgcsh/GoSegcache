# 根据 go_segcache.proto生成 2哥go文件 命令

* 在 proto目录下执行

```shell
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative go_segcache.proto
```

