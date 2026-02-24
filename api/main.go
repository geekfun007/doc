package main

import (
	"log"

	"github.com/cloudwego/hertz/pkg/app/server"

	"byte.dance/api/biz/router"
	"byte.dance/api/biz/rpc"
	"byte.dance/pkg/consts"
)

func main() {
	rpc.Init()

	h := server.Default(server.WithHostPorts(consts.APIHTTPAddr))
	router.Register(h)

	log.Printf("[%s] HTTP gateway listening on %s", consts.APIServiceName, consts.APIHTTPAddr)
	h.Spin()
}
