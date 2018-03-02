/*
                      _ooOoo_
                     o8888888o
                     88" . "88
                     (| -_- |)
                     O\  =  /O
                  ____/`---'\____
                .'  \\|     |//  `.
               /  \\|||  :  |||//  \
              /  _||||| -:- |||||-  \
              |   | \\\  -  /// |   |
              | \_|  ''\---/''  |   |
              \  .-\__  `-`  ___/-. /
            ___`. .'  /--.--\  `. . __
         ."" '<  `.___\_<|>_/___.'  >'"".
        | | :  `- \`.;`\ _ /`;.`/ - ` : | |
        \  \ `-.   \_ __\ /__ _/   .-` /  /
   ======`-.____`-.___\_____/___.-`____.-'======
                      `=---='
   ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
				佛祖保佑       永无BUG
*/

package main

import (
	"git.oschina.net/kuaishangtong/common/utils/log"
	"git.oschina.net/kuaishangtong/navi/gateway/constants"
	_ "git.oschina.net/kuaishangtong/navi/gateway/routers"
	"git.oschina.net/kuaishangtong/navi/gateway/service"
	"git.oschina.net/kuaishangtong/navi/ipfilter"
	"github.com/astaxie/beego"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"

	//"github.com/opentracing/opentracing-go"
	//zipkin "github.com/openzipkin/zipkin-go-opentracing"

	//"github.com/opentracing/opentracing-go"
	//"github.com/uber/jaeger-client-go"
	//"github.com/uber/jaeger-client-go/transport/zipkin"
	_ "net/http/pprof"
	"time"
)

func main() {
	log.SetLevel(beego.AppConfig.DefaultInt("logLevel", 6))

	err := service.InitServiceManager()
	if err != nil {
		log.Fatal(err)
	}

	err = ipfilter.InitIpFilter(constants.ZookeeperHosts, constants.IPFilterPath)
	if err != nil {
		log.Fatal(err)
	}

	// Recommended configuration for production.
	cfg := jaegercfg.Configuration{
		Sampler: &jaegercfg.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
			LocalAgentHostPort:  "192.168.1.16:5775",
		},
	}

	// Example logger and metrics factory. Use github.com/uber/jaeger-client-go/log
	// and github.com/uber/jaeger-lib/metrics respectively to bind to real logging and metrics
	// frameworks.
	jLogger := jaegerlog.StdLogger
	jMetricsFactory := metrics.NullFactory

	// Initialize tracer with a logger and a metrics factory
	closer, err := cfg.InitGlobalTracer(
		"gateway",
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)
	if err != nil {
		log.Fatalf("Could not initialize jaeger tracer: %s", err.Error())
		return
	}
	defer closer.Close()

	/*#################################################################*/

	//collector, err := zipkin.NewHTTPCollector("http://192.168.1.16:9411/api/v1/spans")
	//if err != nil {
	//	log.Fatal(err)
	//	return
	//}
	//
	//tracer, err := zipkin.NewTracer(
	//	zipkin.NewRecorder(collector, false, "localhost:0", "gateway"),
	//	zipkin.ClientServerSameSpan(true),
	//	zipkin.TraceID128Bit(true),
	//)
	//if err != nil {
	//	log.Fatal(err)
	//	return
	//}
	//opentracing.InitGlobalTracer(tracer)

	//// Jaeger tracer can be initialized with a transport that will
	//// report tracing Spans to a Zipkin backend
	//transport, err := zipkin.NewHTTPTransport(
	//	"http://192.168.1.16:9411/api/v1/spans",
	//	zipkin.HTTPBatchSize(1),
	//	zipkin.HTTPLogger(jaeger.StdLogger),
	//)
	//if err != nil {
	//	log.Fatalf("Cannot initialize HTTP transport: %v", err)
	//}
	//// create Jaeger tracer
	//tracer, closer := jaeger.NewTracer(
	//	"gateway",
	//	jaeger.NewConstSampler(true), // sample all traces
	//	jaeger.NewRemoteReporter(transport),
	//)
	//
	//opentracing.InitGlobalTracer(tracer)

	beego.Run()

	//closer.Close()
}
