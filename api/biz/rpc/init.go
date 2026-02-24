package rpc

import (
	"sync"

	"github.com/cloudwego/kitex/client"
	etcd "github.com/kitex-contrib/registry-etcd"

	"byte.dance/kitex_gen/article/articleservice"
	"byte.dance/kitex_gen/user/userservice"
	"byte.dance/pkg/consts"
)

var (
	userClient    userservice.Client
	articleClient articleservice.Client
	once          sync.Once
)

func Init() {
	once.Do(func() {
		initUserClient()
		initArticleClient()
	})
}

func initUserClient() {
	r, err := etcd.NewEtcdResolver([]string{consts.EtcdEndpoint})
	if err != nil {
		panic(err)
	}
	c, err := userservice.NewClient(consts.UserServiceName, client.WithResolver(r))
	if err != nil {
		panic(err)
	}
	userClient = c
}

func initArticleClient() {
	r, err := etcd.NewEtcdResolver([]string{consts.EtcdEndpoint})
	if err != nil {
		panic(err)
	}
	c, err := articleservice.NewClient(consts.ArticleServiceName, client.WithResolver(r))
	if err != nil {
		panic(err)
	}
	articleClient = c
}

func UserClient() userservice.Client {
	return userClient
}

func ArticleClient() articleservice.Client {
	return articleClient
}
