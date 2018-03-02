package main

import (
	"encoding/json"
	"fmt"
	glog "git.oschina.net/kuaishangtong/common/utils/log"
	"github.com/gorilla/mux"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/transport/zipkin"
	"io/ioutil"
	"net/http"
	"time"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/ping", pingHandler)
	r.HandleFunc("/servicename", serviceNameHandler)
	r.HandleFunc("/servicemode", serviceModeHandler)
	r.HandleFunc("/hello", helloHandler)
	http.Handle("/", r)

	// Jaeger tracer can be initialized with a transport that will
	// report tracing Spans to a Zipkin backend
	transport, err := zipkin.NewHTTPTransport(
		"http://192.168.1.16:9411/api/v1/spans",
		zipkin.HTTPBatchSize(1),
		zipkin.HTTPLogger(jaeger.StdLogger),
	)
	if err != nil {
		glog.Fatalf("Cannot initialize HTTP transport: %v", err)
	}
	// create Jaeger tracer
	tracer, closer := jaeger.NewTracer(
		"MyTest",
		jaeger.NewConstSampler(true), // sample all traces
		jaeger.NewRemoteReporter(transport),
	)
	defer closer.Close()

	opentracing.InitGlobalTracer(tracer)

	glog.Info("MyTest Http server start")
	err = http.ListenAndServe(":8081",
		nil /*nethttp.Middleware(tracer, http.DefaultServeMux, nethttp.MWComponentName("MyTestHttpServer"))*/)
	if err != nil {
		glog.Fatal("ListenAndServe: ", err.Error())
	}
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	glog.Infof("req remoteAddr: %s for url /%s", r.RemoteAddr, r.URL.Path)
	fmt.Fprintf(w, "pong")
}

func serviceNameHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	glog.Infof("req remoteAddr: %s for url /%s", r.RemoteAddr, r.URL.Path)
	fmt.Fprintf(w, "MyHttpTest")
}

func serviceModeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	glog.Infof("req remoteAddr: %s for url /%s", r.RemoteAddr, r.URL.Path)
	fmt.Fprintf(w, "dev")
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
		glog.Errorf("req Method is not POST")
		return
	}

	glog.Info(r.Header)

	textCarrier := opentracing.HTTPHeadersCarrier(r.Header)
	wireSpanContext, err := opentracing.GlobalTracer().Extract(
		opentracing.HTTPHeaders, textCarrier)
	if err != nil {
		glog.Errorf("req remoteAddr: %s for url %s Extract err: %v", r.RemoteAddr, r.URL.Path, err)
	}

	serverSpan := opentracing.GlobalTracer().StartSpan(
		"POST sayHello",
		opentracing.ChildOf(wireSpanContext))
	serverSpan.SetTag("component", "mytest server")
	defer serverSpan.Finish()

	glog.Infof("req remoteAddr: %s for url %s", r.RemoteAddr, r.URL.Path)
	serverSpan.LogFields(log.String("remoteAddr", r.RemoteAddr))

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		glog.Errorf("read req body err: %v", err)
		serverSpan.LogFields(log.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	timeSpan := serverSpan.Tracer().StartSpan("time sleep")
	time.Sleep(time.Second)
	timeSpan.LogFields(log.String("sleep", "1 second"))
	timeSpan.Finish()

	var body map[string]interface{}

	err = json.Unmarshal(data, &body)
	if err != nil {
		glog.Errorf("Unmarshal req json body err: %v", err)
		serverSpan.LogFields(log.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	glog.Info(body)
	serverSpan.LogFields(log.String("request body", body["yourname"].(string)))

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "hello %s", body["yourname"].(string))
}
