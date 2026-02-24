package consts

import "os"

const (
	UserServiceName    = "byte.dance.user"
	ArticleServiceName = "byte.dance.article"
	APIServiceName     = "byte.dance.api"

	DefaultEtcdEndpoint = "127.0.0.1:2379"
	UserRPCAddr         = ":8881"
	ArticleRPCAddr      = ":8882"
	APIHTTPAddr         = ":8080"

	JWTSecret = "byte-dance-jwt-secret-key"
	JWTExpire = 24 // hours
)

var EtcdEndpoint = defaultEnv("ETCD_ENDPOINT", DefaultEtcdEndpoint)

func defaultEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
