package main

import (
	"flag"
	"net/http"

	"github.com/weeco/cortex-gateway/gateway"

	"github.com/cortexproject/cortex/pkg/util"
	"github.com/cortexproject/cortex/pkg/util/flagext"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	"github.com/weaveworks/common/middleware"
	"github.com/weaveworks/common/server"
	"github.com/weaveworks/common/tracing"
	"google.golang.org/grpc"
)

func main() {
	operationNameFunc := nethttp.OperationNameFunc(func(r *http.Request) string {
		return r.URL.RequestURI()
	})

	var (
		serverCfg = server.Config{
			MetricsNamespace: "cortex",
			HTTPMiddleware: []middleware.Interface{
				middleware.Func(func(handler http.Handler) http.Handler {
					return nethttp.Middleware(opentracing.GlobalTracer(), handler, operationNameFunc)
				}),
			},
			GRPCMiddleware: []grpc.UnaryServerInterceptor{
				middleware.ServerUserHeaderInterceptor,
			},
		}
		gatewayCfg gateway.Config
	)

	flagext.RegisterFlags(&serverCfg, &gatewayCfg)
	flag.Parse()

	util.InitLogger(&serverCfg)

	// Setting the environment variable JAEGER_AGENT_HOST enables tracing
	trace := tracing.NewFromEnv("cortex-gateway")
	defer trace.Close()

	svr, err := server.New(serverCfg)
	util.CheckFatal("initializing server", err)
	defer svr.Shutdown()

	// Setup proxy and register routes
	gateway, err := gateway.New(gatewayCfg, svr)
	util.CheckFatal("initializing gateway", err)
	gateway.Start()

	svr.Run()
}
