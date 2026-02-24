package main

import (
	"log"
	"net"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	etcd "github.com/kitex-contrib/registry-etcd"

	"byte.dance/kitex_gen/user/userservice"
	"byte.dance/pkg/consts"
)

func main() {
	r, err := etcd.NewEtcdRegistry([]string{consts.EtcdEndpoint})
	if err != nil {
		log.Fatalf("etcd registry init failed: %v", err)
	}

	addr, _ := net.ResolveTCPAddr("tcp", consts.UserRPCAddr)
	svr := userservice.NewServer(
		new(UserServiceImpl),
		server.WithServiceAddr(addr),
		server.WithRegistry(r),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: consts.UserServiceName,
		}),
	)

	log.Printf("[%s] listening on %s", consts.UserServiceName, consts.UserRPCAddr)
	if err := svr.Run(); err != nil {
		log.Fatalf("%s server stopped: %v", consts.UserServiceName, err)
	}
}
