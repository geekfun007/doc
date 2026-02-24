package main

import (
	"log"

	"github.com/cloudwego/hertz/pkg/app/server"

	"byte.dance/api/biz/rpc"
	"byte.dance/pkg/consts"
)

func main() {
	rpc.Init()

	h := server.Default(server.WithHostPorts(consts.APIHTTPAddr))
	register(h)

	log.Printf("[%s] HTTP gateway listening on %s", consts.APIServiceName, consts.APIHTTPAddr)
	h.Spin()
}
