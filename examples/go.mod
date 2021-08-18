module github.com/onemoreteam/http_server_framework/examples

go 1.16

replace github.com/onemoreteam/http_server_framework => ../

require (
	github.com/onemoreteam/http_server_framework v0.0.0-00010101000000-000000000000
	google.golang.org/genproto v0.0.0-20210816143620-e15ff196659d // indirect
	google.golang.org/grpc v1.40.0
	google.golang.org/grpc/examples v0.0.0-20210812181202-a42567fe92f0
)
